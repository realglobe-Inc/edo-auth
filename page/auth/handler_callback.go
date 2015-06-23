// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"time"
)

func (this *Page) HandleCallback(w http.ResponseWriter, r *http.Request) {
	var sender *request.Request

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			server.RespondErrorHtml(w, r, erro.New(rcv), this.errTmpl, sender.String()+": ")
			return
		}
	}()

	if this.stopper != nil {
		this.stopper.Stop()
		defer this.stopper.Unstop()
	}

	//////////////////////////////
	server.LogRequest(level.DEBUG, r, this.debug)
	//////////////////////////////

	sender = request.Parse(r, this.sessLabel)
	log.Info(sender, ": Received callback request")
	defer log.Info(sender, ": Handled callback request")

	if err := (&environment{this, sender, nil}).callbackServe(w, r); err != nil {
		server.RespondErrorHtml(w, r, erro.Wrap(err), this.errTmpl, sender.String()+": ")
		return
	}
	return
}

func (this *environment) callbackServe(w http.ResponseWriter, r *http.Request) error {
	if this.sender.Session() == "" {
		return erro.Wrap(server.NewError(http.StatusBadRequest, "no session ", nil))
	}

	sess, err := this.sessDb.Get(this.sender.Session())
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		return erro.Wrap(server.NewError(http.StatusBadRequest, "declared user session is not exist", nil))
	}
	this.sess = sess
	log.Debug(this.sender, ": Declared user session is exist")

	savedDate := sess.Date()
	sess.Invalidate()
	if ok, err := this.sessDb.Replace(sess, savedDate); err != nil {
		return erro.Wrap(err)
	} else if !ok {
		return erro.Wrap(server.NewError(http.StatusBadRequest, "reused user session", nil))
	}

	req, err := parseCallbackRequest(r)
	if err != nil {
		return erro.Wrap(server.NewError(http.StatusBadRequest, erro.Unwrap(err).Error(), err))
	}

	log.Debug(this.sender, ": Parsed callback request")

	if req.state() != sess.State() {
		return erro.Wrap(server.NewError(http.StatusForbidden, "invalid state", nil))
	}

	var idp idpdb.Element
	var attrs1 map[string]interface{}

	if sess.IdProvider() != "" {
		idp, err = this.idpDb.Get(sess.IdProvider())
		if err != nil {
			return erro.Wrap(err)
		} else if idp == nil {
			return erro.Wrap(server.NewError(http.StatusBadRequest, "ID provider "+sess.IdProvider()+" is not exist", nil))
		}
		log.Debug(this.sender, ": ID provider "+idp.Id()+" is exist")
	} else {
		idTok, err := parseIdToken(req.idToken())
		if err != nil {
			return erro.Wrap(server.NewError(http.StatusBadRequest, erro.Unwrap(err).Error(), err))
		}

		idp, err = this.idpDb.Get(idTok.idProvider())
		if err != nil {
			return erro.Wrap(err)
		} else if idp == nil {
			return erro.Wrap(server.NewError(http.StatusBadRequest, "ID provider "+idTok.idProvider()+" is not exist", nil))
		}
		log.Debug(this.sender, ": ID provider "+idp.Id()+" is exist")

		if idTok.nonce() != sess.Nonce() {
			return erro.Wrap(server.NewError(http.StatusForbidden, "invalid nonce", nil))
		} else if err := idTok.verify(idp.Keys()); err != nil {
			return erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
		} else if err := idTok.verifyCodeHash(req.code()); err != nil {
			return erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
		}
		attrs1 = idTok.attributes()
		log.Debug(this.sender, ": ID token is OK")
	}

	// アクセストークンを取得する。
	tok, idTok, err := this.getAccessToken(req, idp)
	if err != nil {
		return erro.Wrap(err)
	}

	// アカウント情報を取得する。
	attrs2, err := this.getAccountInfo(req, tok, idp)
	if err != nil {
		return erro.Wrap(err)
	}

	// アカウント情報をまとめる。
	jt := jwt.New()
	jt.SetHeader(tagAlg, tagNone)
	for _, m := range []map[string]interface{}{attrs1, idTok.attributes(), attrs2} {
		for k, v := range m {
			jt.SetClaim(k, v)
		}
	}
	jt.SetClaim(tagAt_tag, tok.Tag())
	jt.SetClaim(tagAt_exp, tok.Expires().Unix())
	buff, err := jt.Encode()
	if err != nil {
		return erro.Wrap(err)
	}

	// フロントエンドのためにセッション期限を延長する。
	now := time.Now()
	http.SetCookie(w, this.newCookie(sess.Id(), now.Add(-time.Second)))
	http.SetCookie(w, this.newFrontCookie(this.idGen.String(this.fsessLen), now.Add(this.fsessExpIn)))
	log.Info(this.sender, ": Upgrade user session to frontend session")

	// フロントエンドが使うので保存しなくて良い。

	w.Header().Add(tagX_auth_user, string(buff))
	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)

	http.Redirect(w, r, sess.Path(), http.StatusFound)
	log.Info(this.sender, ": Redirect to "+sess.Path())
	return nil
}
