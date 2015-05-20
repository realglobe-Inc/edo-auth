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
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

// ユーザー認証開始。
func (sys *system) authPage(w http.ResponseWriter, r *http.Request) error {
	req := request.Parse(r, tagAuth_user)
	log.Info(req, ": Received authentication request")
	defer log.Info(req, ": Handled authentication request")

	if err := sys.authServe(w, r, req); err != nil {
		return sys.responseError(w, r, erro.Wrap(err), req)
	}
	return nil
}

func (sys *system) authServe(w http.ResponseWriter, r *http.Request, sender *request.Request) error {
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
	if idps, err := sys.idpDb.Search(map[string]string{
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
		ta = sys.taId
		queries.Set(tagClient_id, ta)
		log.Debug(req, ": Act as default TA "+ta)
	} else {
		log.Debug(req, ": Act as TA "+ta)
	}

	// redirect_uri
	rediUri := queries.Get(tagRedirect_uri)
	if rediUri == "" {
		rediUri = sys.rediUri
		queries.Set(tagRedirect_uri, rediUri)
		log.Debug(req, ": Use default redirect uri "+rediUri)
	} else {
		log.Debug(req, ": Use redirect uri "+rediUri)
	}

	// state
	stat := randomString(sys.statLen)
	queries.Set(tagState, stat)
	log.Debug(req, ": Use state "+mosaic(stat))

	// nonce
	nonc := randomString(sys.noncLen)
	queries.Set(tagNonce, nonc)
	log.Debug(req, ": Use nonce "+mosaic(nonc))

	authUri.RawQuery = queries.Encode()

	sess := usession.New(
		randomString(sys.usessLen),
		time.Now().Add(sys.authExpIn),
		req.path(),
		idp,
		ta,
		rediUri,
		stat,
		nonc,
	)
	if err := sys.usessDb.Save(sess, sess.Expires().Add(sys.usessDbExpIn-sys.usessExpIn)); err != nil {
		return erro.Wrap(err)
	}
	log.Info(req, ": Generated user session "+mosaic(sess.Id()))

	http.SetCookie(w, sys.newUserCookie(sess))
	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)

	uri := authUri.String()
	http.Redirect(w, r, uri, http.StatusFound)
	log.Info(req, ": Redirect to "+uri)
	return nil
}
