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

package main

import (
	"github.com/realglobe-Inc/edo-auth/database/usession"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

// ユーザー情報。
func (sys *system) callbackPage(w http.ResponseWriter, r *http.Request) (err error) {
	sender := request.Parse(r, sys.usessLabel)
	log.Info(sender, ": Received callback request")
	defer log.Info(sender, ": Handled callback request")

	if err := sys.callbackServe(w, r, sender); err != nil {
		return server.RespondPageError(w, r, erro.Wrap(err), sys.errTmpl, sender.String()+": ")
	}
	return nil
}

func (sys *system) callbackServe(w http.ResponseWriter, r *http.Request, sender *request.Request) error {
	if sender.Session() == "" {
		return erro.Wrap(server.NewError(http.StatusBadRequest, "no session ", nil))
	}

	sess, err := sys.usessDb.Get(sender.Session())
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		return erro.Wrap(server.NewError(http.StatusBadRequest, "declared user session is not exist", nil))
	}
	log.Debug(sender, ": Declared user session is exist")

	savedDate := sess.Date()
	sess.Invalidate()
	if ok, err := sys.usessDb.Replace(sess, savedDate); err != nil {
		return erro.Wrap(err)
	} else if !ok {
		return erro.Wrap(server.NewError(http.StatusBadRequest, "reused user session", nil))
	}

	req, err := parseCallbackRequest(r, sender)
	if err != nil {
		return erro.Wrap(server.NewError(http.StatusBadRequest, erro.Unwrap(err).Error(), err))
	}

	log.Debug(req, ": Parsed callback request")

	if req.state() != sess.State() {
		return erro.Wrap(server.NewError(http.StatusForbidden, "invalid nonce", nil))
	}

	var idp idpdb.Element
	var attrs1 map[string]interface{}

	if sess.IdProvider() != "" {
		idp, err = sys.idpDb.Get(sess.IdProvider())
		if err != nil {
			return erro.Wrap(err)
		} else if idp == nil {
			return erro.Wrap(server.NewError(http.StatusBadRequest, "ID provider "+sess.IdProvider()+" is not exist", nil))
		}
		log.Debug(req, ": ID provider "+idp.Id()+" is exist")
	} else {
		idTok, err := parseIdToken(req.idToken())
		if err != nil {
			return erro.Wrap(server.NewError(http.StatusBadRequest, erro.Unwrap(err).Error(), err))
		}

		idp, err = sys.idpDb.Get(idTok.idProvider())
		if err != nil {
			return erro.Wrap(err)
		} else if idp == nil {
			return erro.Wrap(server.NewError(http.StatusBadRequest, "ID provider "+idTok.idProvider()+" is not exist", nil))
		}
		log.Debug(req, ": ID provider "+idp.Id()+" is exist")

		if idTok.nonce() != sess.Nonce() {
			return erro.Wrap(server.NewError(http.StatusForbidden, "invalid nonce", nil))
		} else if err := idTok.verify(idp.Keys()); err != nil {
			return erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
		} else if err := idTok.verifyCodeHash(req.code()); err != nil {
			return erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
		}
		attrs1 = idTok.attributes()
		log.Debug(req, ": ID token is OK")
	}

	// アクセストークンを取得する。
	tok, idTok, err := sys.getAccessToken(req, idp, sess)
	if err != nil {
		return erro.Wrap(err)
	}
	log.Debug(req, ": Got access token "+mosaic(tok.Id()))

	// アカウント情報を取得する。
	attrs2, err := sys.getAccountInfo(req, tok, idp, sess)
	if err != nil {
		return erro.Wrap(err)
	}
	log.Debug(req, ": Got account info")

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
	sess2 := usession.New(
		sess.Id(),
		time.Now().Add(sys.usessExpIn),
		"",
		"",
		"",
		"",
		"",
		"",
	)
	log.Info(req, ": Rescheduled user session "+mosaic(sess2.Id()))
	http.SetCookie(w, sys.newUserCookie(sess2))

	w.Header().Add(tagX_auth_user, string(buff))
	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)

	http.Redirect(w, r, sess.Path(), http.StatusFound)
	log.Info(req, ": Redirect to "+sess.Path())
	return nil
}
