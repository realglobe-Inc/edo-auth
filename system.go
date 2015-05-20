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
	"encoding/json"
	"github.com/realglobe-Inc/edo-auth/database/token"
	"github.com/realglobe-Inc/edo-auth/database/usession"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type system struct {
	taId    string
	rediUri string
	sigAlg  string
	sigKid  string

	errTmpl *template.Template

	usessLabel   string
	usessLen     int
	authExpIn    time.Duration
	usessExpIn   time.Duration
	usessDbExpIn time.Duration
	statLen      int
	noncLen      int
	tokTagLen    int
	tokDbExpIn   time.Duration
	jtiLen       int
	jtiExpIn     time.Duration

	keyDb   keydb.Db
	webDb   webdb.Db
	idpDb   idpdb.Db
	usessDb usession.Db
	tokDb   token.Db

	cookPath string
	cookSec  bool
}

func (sys *system) newUserCookie(sess *usession.Element) *http.Cookie {
	return &http.Cookie{
		Name:     sys.usessLabel,
		Value:    sess.Id(),
		Path:     sys.cookPath,
		Expires:  sess.Expires(),
		Secure:   sys.cookSec,
		HttpOnly: true,
	}
}

// 認可コードを使って、ID プロバイダからアクセストークンを取得する。
func (sys *system) getAccessToken(req *callbackRequest, idp idpdb.Element, sess *usession.Element) (*token.Element, *idToken, error) {
	keys, err := sys.keyDb.Get()
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}

	queries := url.Values{}
	// grant_type
	queries.Set(tagGrant_type, tagAuthorization_code)
	// code
	queries.Set(tagCode, req.code())
	// redirect_uri
	queries.Set(tagRedirect_uri, sess.RedirectUri())
	// client_id
	queries.Set(tagClient_id, sess.Ta())
	// client_assertion_type
	queries.Set(tagClient_assertion_type, cliAssTypeJwt_bearer)
	// client_assertion
	ass := jwt.New()
	now := time.Now()
	ass.SetHeader(tagAlg, sys.sigAlg)
	ass.SetClaim(tagIss, sess.Ta())
	ass.SetClaim(tagSub, sess.Ta())
	ass.SetClaim(tagAud, idp.TokenUri())
	ass.SetClaim(tagJti, randomString(sys.jtiLen))
	ass.SetClaim(tagExp, now.Add(sys.jtiExpIn).Unix())
	ass.SetClaim(tagIat, now.Unix())
	if err := ass.Sign(keys); err != nil {
		return nil, nil, erro.Wrap(err)
	}
	buff, err := ass.Encode()
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}
	queries.Set(tagClient_assertion, string(buff))

	tokReq, err := http.NewRequest("POST", idp.TokenUri(), strings.NewReader(queries.Encode()))
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}
	tokReq.Header.Set(tagContent_type, contTypeForm)

	server.LogRequest(level.DEBUG, tokReq, true)
	resp, err := (&http.Client{}).Do(tokReq)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}
	defer resp.Body.Close()
	server.LogResponse(level.DEBUG, resp, true)
	log.Info(req, ": Got token response from "+idp.Id())

	tokResp, err := parseTokenResponse(resp)
	if err != nil {
		return nil, nil, erro.Wrap(server.NewError(http.StatusForbidden, "cannot get access token", nil))
	}
	tok := token.New(tokResp.token(), randomString(sys.tokTagLen), tokResp.expires(), idp.Id(), tokResp.scope())
	log.Info(req, ": Got access token "+mosaic(tok.Id()))

	if err := sys.tokDb.Save(tok, time.Now().Add(sys.tokDbExpIn)); err != nil {
		return nil, nil, erro.Wrap(err)
	}
	log.Info(req, ": Saved access token "+mosaic(tok.Id()))

	idTok, err := parseIdToken(tokResp.idToken())
	if err != nil {
		return nil, nil, erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
	} else if idTok.nonce() != sess.Nonce() {
		return nil, nil, erro.Wrap(server.NewError(http.StatusBadRequest, "invalid nonce", nil))
	} else if err := idTok.verify(idp.Keys()); err != nil {
		return nil, nil, erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
	} else if err := idTok.verifyTokenHash(tok.Id()); err != nil {
		return nil, nil, erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
	}
	log.Info(req, ": ID token is OK")

	return tok, idTok, nil
}

// アクセストークンを使って、ID プロバイダからアカウント情報を取得する。
func (sys *system) getAccountInfo(req *callbackRequest, tok *token.Element, idp idpdb.Element, sess *usession.Element) (attrs map[string]interface{}, err error) {
	acntReq, err := http.NewRequest("GET", idp.AccountUri(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	acntReq.Header.Set(tagAuthorization, tagBearer+" "+tok.Id())

	server.LogRequest(level.DEBUG, acntReq, true)
	resp, err := (&http.Client{}).Do(acntReq)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer resp.Body.Close()
	server.LogResponse(level.DEBUG, resp, true)
	log.Info(req, ": Got account info response from "+idp.Id())

	if err := json.NewDecoder(resp.Body).Decode(&attrs); err != nil {
		return nil, erro.Wrap(err)
	}
	log.Info(req, ": Got account info")

	return attrs, nil
}
