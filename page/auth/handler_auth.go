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
	"github.com/realglobe-Inc/edo-auth/database/asession"
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
	var logPref string

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			server.RespondErrorHtml(w, r, erro.New(rcv), this.errTmpl, logPref)
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

	logPref = server.ParseSender(r) + ": "

	log.Info(logPref, "Received authentication request")
	defer log.Info(logPref, "Handled authentication request")

	if err := (&environment{this, logPref, "", nil}).authServe(w, r); err != nil {
		server.RespondErrorHtml(w, r, erro.Wrap(err), this.errTmpl, logPref)
		return
	}
	return
}

func (this *environment) authServe(w http.ResponseWriter, r *http.Request) error {
	req, err := parseAuthRequest(r)
	if err != nil {
		return erro.Wrap(server.NewError(http.StatusBadRequest, erro.Unwrap(err).Error(), err))
	}

	log.Debug(this.logPref, "Parsed authentication request")

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
		log.Debug(this.logPref, "Destination is in ID provider "+idp)
	} else {
		// ID プロバイダ選択サービスか何か。
		respType[tagId_token] = true
		log.Debug(this.logPref, "Destination "+rawAuthUri+" is not ID provider")
	}
	queries.Set(tagResponse_type, request.ValueSetForm(respType))

	// scope
	if scop := request.FormValueSet(queries.Get(tagScope)); !scop[tagOpenid] {
		scop[tagOpenid] = true
		queries.Set(tagScope, request.ValueSetForm(scop))
		log.Debug(this.logPref, `: Added scope "`+tagOpenid+`"`)
	}

	// client_id
	ta := queries.Get(tagClient_id)
	if ta == "" {
		ta = this.selfId
		queries.Set(tagClient_id, ta)
		log.Debug(this.logPref, "Act as default TA "+ta)
	} else {
		log.Debug(this.logPref, "Act as TA "+ta)
	}

	// redirect_uri
	rediUri := queries.Get(tagRedirect_uri)
	if rediUri == "" {
		rediUri = this.rediUri
		queries.Set(tagRedirect_uri, rediUri)
		log.Debug(this.logPref, "Use default redirect uri "+rediUri)
	} else {
		log.Debug(this.logPref, "Use redirect uri "+rediUri)
	}

	// state
	stat := this.idGen.String(this.statLen)
	queries.Set(tagState, stat)
	log.Debug(this.logPref, "Use state "+logutil.Mosaic(stat))

	// nonce
	nonc := this.idGen.String(this.noncLen)
	queries.Set(tagNonce, nonc)
	log.Debug(this.logPref, "Use nonce "+logutil.Mosaic(nonc))

	authUri.RawQuery = queries.Encode()

	sess := asession.New(
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
	log.Info(this.logPref, "Generated user session "+logutil.Mosaic(sess.Id()))

	http.SetCookie(w, this.newCookie(sess.Id(), sess.Expires()))
	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)

	uri := authUri.String()
	http.Redirect(w, r, uri, http.StatusFound)
	log.Info(this.logPref, "Redirect to "+uri)
	return nil
}
