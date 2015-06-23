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

// ユーザー認証代行ページ。
package auth

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-auth/database/asession"
	"github.com/realglobe-Inc/edo-auth/database/token"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Page struct {
	stopper *server.Stopper

	selfId  string
	rediUri string
	sigAlg  string

	errTmpl *template.Template

	sessLabel   string
	sessLen     int
	sessExpIn   time.Duration
	sessDbExpIn time.Duration
	fsessLabel  string
	fsessLen    int
	fsessExpIn  time.Duration
	statLen     int
	noncLen     int
	tokTagLen   int
	tokDbExpIn  time.Duration
	jtiLen      int
	jtiExpIn    time.Duration

	keyDb  keydb.Db
	idpDb  idpdb.Db
	sessDb asession.Db
	tokDb  token.Db
	idGen  rand.Generator

	cookPath string
	cookSec  bool
	debug    bool
}

func New(
	stopper *server.Stopper,
	selfId string,
	rediUri string,
	sigAlg string,
	errTmpl *template.Template,
	sessLabel string,
	sessLen int,
	sessExpIn time.Duration,
	sessDbExpIn time.Duration,
	fsessLabel string,
	fsessLen int,
	fsessExpIn time.Duration,
	statLen int,
	noncLen int,
	tokTagLen int,
	tokDbExpIn time.Duration,
	jtiLen int,
	jtiExpIn time.Duration,
	keyDb keydb.Db,
	idpDb idpdb.Db,
	sessDb asession.Db,
	tokDb token.Db,
	idGen rand.Generator,
	cookPath string,
	cookSec bool,
	debug bool,
) *Page {
	return &Page{
		stopper,
		selfId,
		rediUri,
		sigAlg,
		errTmpl,
		sessLabel,
		sessLen,
		sessExpIn,
		sessDbExpIn,
		fsessLabel,
		fsessLen,
		fsessExpIn,
		statLen,
		noncLen,
		tokTagLen,
		tokDbExpIn,
		jtiLen,
		jtiExpIn,
		keyDb,
		idpDb,
		sessDb,
		tokDb,
		idGen,
		cookPath,
		cookSec,
		debug,
	}
}

func (this *Page) newCookie(id string, exp time.Time) *http.Cookie {
	return this._newCookie(this.sessLabel, id, exp)
}

func (this *Page) newFrontCookie(id string, exp time.Time) *http.Cookie {
	return this._newCookie(this.fsessLabel, id, exp)
}

func (this *Page) _newCookie(label, id string, exp time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     label,
		Value:    id,
		Path:     this.cookPath,
		Expires:  exp,
		Secure:   this.cookSec,
		HttpOnly: true,
	}
}

// environment のメソッドは server.Error を返す。
type environment struct {
	*Page

	sender *request.Request
	sess   *asession.Element
}

// 認可コードを使って、ID プロバイダからアクセストークンを取得する。
func (this *environment) getAccessToken(req *callbackRequest, idp idpdb.Element) (*token.Element, *idToken, error) {
	keys, err := this.keyDb.Get()
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}

	queries := url.Values{}
	// grant_type
	queries.Set(tagGrant_type, tagAuthorization_code)
	// code
	queries.Set(tagCode, req.code())
	// redirect_uri
	queries.Set(tagRedirect_uri, this.sess.RedirectUri())
	// client_id
	queries.Set(tagClient_id, this.sess.Ta())
	// client_assertion_type
	queries.Set(tagClient_assertion_type, cliAssTypeJwt_bearer)
	// client_assertion
	ass := jwt.New()
	now := time.Now()
	ass.SetHeader(tagAlg, this.sigAlg)
	ass.SetClaim(tagIss, this.sess.Ta())
	ass.SetClaim(tagSub, this.sess.Ta())
	ass.SetClaim(tagAud, idp.TokenUri())
	ass.SetClaim(tagJti, this.idGen.String(this.jtiLen))
	ass.SetClaim(tagExp, now.Add(this.jtiExpIn).Unix())
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

	server.LogRequest(level.DEBUG, tokReq, this.debug)
	resp, err := (&http.Client{}).Do(tokReq)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}
	defer resp.Body.Close()
	server.LogResponse(level.DEBUG, resp, this.debug)
	log.Info(this.sender, ": Got token response from "+idp.Id())

	tokResp, err := parseTokenResponse(resp)
	if err != nil {
		return nil, nil, erro.Wrap(server.NewError(http.StatusForbidden, "cannot get access token", nil))
	}
	tok := token.New(tokResp.token(), this.idGen.String(this.tokTagLen), tokResp.expires(), idp.Id(), tokResp.scope())
	log.Info(this.sender, ": Got access token "+logutil.Mosaic(tok.Id()))

	if err := this.tokDb.Save(tok, time.Now().Add(this.tokDbExpIn)); err != nil {
		return nil, nil, erro.Wrap(err)
	}
	log.Info(this.sender, ": Saved access token "+logutil.Mosaic(tok.Id()))

	idTok, err := parseIdToken(tokResp.idToken())
	if err != nil {
		return nil, nil, erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
	} else if idTok.nonce() != this.sess.Nonce() {
		return nil, nil, erro.Wrap(server.NewError(http.StatusForbidden, "invalid nonce", nil))
	} else if err := idTok.verify(idp.Keys()); err != nil {
		return nil, nil, erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
	} else if idTok.tokenHash() != nil {
		if err := idTok.verifyTokenHash(tok.Id()); err != nil {
			return nil, nil, erro.Wrap(server.NewError(http.StatusForbidden, erro.Unwrap(err).Error(), err))
		}
	}
	log.Info(this.sender, ": ID token is OK")

	return tok, idTok, nil
}

// アクセストークンを使って、ID プロバイダからアカウント情報を取得する。
func (this *environment) getAccountInfo(req *callbackRequest, tok *token.Element, idp idpdb.Element) (attrs map[string]interface{}, err error) {
	acntReq, err := http.NewRequest("GET", idp.AccountUri(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	acntReq.Header.Set(tagAuthorization, tagBearer+" "+tok.Id())

	server.LogRequest(level.DEBUG, acntReq, this.debug)
	resp, err := (&http.Client{}).Do(acntReq)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer resp.Body.Close()
	server.LogResponse(level.DEBUG, resp, this.debug)
	log.Info(this.sender, ": Got account info response from "+idp.Id())

	if err := json.NewDecoder(resp.Body).Decode(&attrs); err != nil {
		return nil, erro.Wrap(err)
	}
	log.Info(this.sender, ": Got account info")

	return attrs, nil
}
