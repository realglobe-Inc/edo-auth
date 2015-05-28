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
	"github.com/realglobe-Inc/edo-auth/database/usession"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"time"
)

// ユーザー認証開始。
func (this *Page) HandleAuth(w http.ResponseWriter, r *http.Request) {
	var sender *request.Request

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			server.RespondPageError(w, r, erro.New(rcv), this.errTmpl, sender.String()+": ")
			return
		}
	}()

	if this.stopper != nil {
		this.stopper.Stop()
		defer this.stopper.Unstop()
	}

	//////////////////////////////
	server.LogRequest(level.DEBUG, r, true)
	//////////////////////////////

	sender = request.Parse(r, this.sessLabel)
	log.Info(sender, ": Received authentication request")
	defer log.Info(sender, ": Handled authentication request")

	if err := this.authServe(w, r, sender); err != nil {
		server.RespondPageError(w, r, erro.Wrap(err), this.errTmpl, sender.String()+": ")
		return
	}
	return
}

func (this *Page) authServe(w http.ResponseWriter, r *http.Request, sender *request.Request) error {
	req, err := parseAuthRequest(r, sender)
	if err != nil {
		return erro.Wrap(server.NewError(http.StatusBadRequest, erro.Unwrap(err).Error(), err))
	}

	log.Debug(req, ": Parsed authentication request")

	authUri := req.authUri()
	queries := authUri.Query()

	// response_type
	var idp string
	respType := map[string]bool{tagCode: true}
	rawAuthUri := authUri.Scheme + "://" + authUri.Host + authUri.Path
	if idps, err := this.idpDb.Search(map[string]string{
		tagAuthorization_endpoint: "^" + rawAuthUri + "$",
	}); err != nil {
		return erro.Wrap(err)
	} else if len(idps) == 1 {
		idp = idps[0].Id()
		log.Debug(req, ": Destination is in ID provider "+idp)
	} else {
		// ID プロバイダ選択サービスか何か。
		respType[tagId_token] = true
		log.Debug(req, ": Destination "+rawAuthUri+" is not ID provider")
	}
	queries.Set(tagResponse_type, request.ValueSetForm(respType))

	// scope
	if scop := request.FormValueSet(queries.Get(tagScope)); !scop[tagOpenid] {
		scop[tagOpenid] = true
		queries.Set(tagScope, request.ValueSetForm(scop))
		log.Debug(req, `: Added scope "`+tagOpenid+`"`)
	}

	// client_id
	ta := queries.Get(tagClient_id)
	if ta == "" {
		ta = this.selfId
		queries.Set(tagClient_id, ta)
		log.Debug(req, ": Act as default TA "+ta)
	} else {
		log.Debug(req, ": Act as TA "+ta)
	}

	// redirect_uri
	rediUri := queries.Get(tagRedirect_uri)
	if rediUri == "" {
		rediUri = this.rediUri
		queries.Set(tagRedirect_uri, rediUri)
		log.Debug(req, ": Use default redirect uri "+rediUri)
	} else {
		log.Debug(req, ": Use redirect uri "+rediUri)
	}

	// state
	stat := this.idGen.String(this.statLen)
	queries.Set(tagState, stat)
	log.Debug(req, ": Use state "+logutil.Mosaic(stat))

	// nonce
	nonc := this.idGen.String(this.noncLen)
	queries.Set(tagNonce, nonc)
	log.Debug(req, ": Use nonce "+logutil.Mosaic(nonc))

	authUri.RawQuery = queries.Encode()

	sess := usession.New(
		this.idGen.String(this.sessLen),
		time.Now().Add(this.sessExpIn),
		req.path(),
		idp,
		ta,
		rediUri,
		stat,
		nonc,
	)
	if err := this.sessDb.Save(sess, sess.Expires().Add(this.sessDbExpIn-this.sessExpIn)); err != nil {
		return erro.Wrap(err)
	}
	log.Info(req, ": Generated user session "+logutil.Mosaic(sess.Id()))

	http.SetCookie(w, this.newCookie(sess.Id(), sess.Expires()))
	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)

	uri := authUri.String()
	http.Redirect(w, r, uri, http.StatusFound)
	log.Info(req, ": Redirect to "+uri)
	return nil
}
