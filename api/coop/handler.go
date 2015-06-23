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

// TA 間連携受け入れ代行。
package coop

import (
	"bytes"
	"encoding/json"
	"github.com/realglobe-Inc/edo-auth/database/token"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"time"
)

type handler struct {
	stopper *server.Stopper

	selfId string
	sigAlg string
	sigKid string

	sessLabel  string
	sessLen    int
	tokTagLen  int
	tokDbExpIn time.Duration
	jtiLen     int
	jtiExpIn   time.Duration

	keyDb keydb.Db
	idpDb idpdb.Db
	tokDb token.Db
	idGen rand.Generator
	tr    http.RoundTripper

	cookPath string
	cookSec  bool
	debug    bool
}

func New(
	stopper *server.Stopper,
	selfId string,
	sigAlg string,
	sigKid string,
	sessLabel string,
	sessLen int,
	tokTagLen int,
	tokDbExpIn time.Duration,
	jtiLen int,
	jtiExpIn time.Duration,
	keyDb keydb.Db,
	idpDb idpdb.Db,
	tokDb token.Db,
	idGen rand.Generator,
	tr http.RoundTripper,
	cookPath string,
	cookSec bool,
	debug bool,
) http.Handler {
	return &handler{
		stopper,
		selfId,
		sigAlg,
		sigKid,
		sessLabel,
		sessLen,
		tokTagLen,
		tokDbExpIn,
		jtiLen,
		jtiExpIn,
		keyDb,
		idpDb,
		tokDb,
		idGen,
		tr,
		cookPath,
		cookSec,
		debug,
	}
}

func (this *handler) httpClient() *http.Client {
	return &http.Client{Transport: this.tr}
}

func (this *handler) newCookie(id string, exp time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     this.sessLabel,
		Value:    id,
		Path:     this.cookPath,
		Expires:  exp,
		Secure:   this.cookSec,
		HttpOnly: true,
	}
}

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var sender *requtil.Request

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			idperr.RespondJson(w, r, erro.New(rcv), sender)
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

	sender = requtil.Parse(r, "")
	log.Info(sender, ": Received cooperation request")
	defer log.Info(sender, ": Handled cooperation request")

	if err := (&environment{this, sender}).serve(w, r); err != nil {
		idperr.RespondJson(w, r, erro.Wrap(err), sender)
		return
	}
}

// environment のメソッドは idperr.Error を返す。
type environment struct {
	*handler

	sender *requtil.Request
}

func (this *environment) serve(w http.ResponseWriter, r *http.Request) error {
	req, err := parseRequest(r)
	if err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	}

	var acntTag string
	tags := map[string]bool{}
	var refHash string
	type idpUnit struct {
		idp    idpdb.Element
		codTok *codeToken
	}
	units := []*idpUnit{}
	for _, rawCodTok := range req.codeTokens() {
		codTok, err := parseCodeToken(rawCodTok)
		if err != nil {
			return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
		} else if !codTok.audience()[this.selfId] {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "invalid audience", http.StatusBadRequest, nil))
		} else if codTok.referralHash() == "" && len(req.codeTokens()) > 1 {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "no referral hash", http.StatusBadRequest, nil))
		} else if codTok.accountTag() != "" {
			if acntTag != "" {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "two main account tags", http.StatusBadRequest, nil))
			}
			acntTag = codTok.accountTag()
			log.Debug(this.sender, ": Main account tag is "+acntTag)
			if codTok.fromTa() == "" {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "no from-TA in main token", http.StatusBadRequest, nil))
			}
			log.Debug(this.sender, ": From-TA is "+codTok.fromTa())
		} else if len(codTok.accountTags()) == 0 {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "no account tags in sub token", http.StatusBadRequest, nil))
		}

		for tag := range codTok.accountTags() {
			if tags[tag] {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "tag "+tag+" overlaps", http.StatusBadRequest, nil))
			}
			tags[tag] = true
			log.Debug(this.sender, ": Account tag is "+tag)
		}

		if codTok.referralHash() != "" {
			if refHash == "" {
				refHash = codTok.referralHash()
			} else if codTok.referralHash() != refHash {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "invalid referral hash", http.StatusBadRequest, nil))
			}
		}

		var idp idpdb.Element
		if idp, err = this.idpDb.Get(codTok.idProvider()); err != nil {
			return erro.Wrap(err)
		} else if idp == nil {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "ID provider "+codTok.idProvider()+" is not exist", http.StatusBadRequest, nil))
		}

		log.Debug(this.sender, ": ID provider "+idp.Id()+" is exist")

		if err := codTok.verify(idp.Keys()); err != nil {
			return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
		}

		log.Debug(this.sender, ": Verified cooperation code")

		units = append(units, &idpUnit{idp, codTok})
	}
	if acntTag == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no main account tag", http.StatusBadRequest, nil))
	}

	log.Debug(this.sender, ": Cooperation codes are OK")

	var tok *token.Element
	var mainAttrs map[string]interface{}
	tagToAttrs := map[string]map[string]interface{}{}
	var frTa string
	for _, unit := range units {
		var tToA map[string]map[string]interface{}
		var fT string
		if unit.codTok.accountTag() != "" {
			fT, tok, tToA, err = this.getInfoFromMainIdProvider(unit.idp, unit.codTok)
			if err != nil {
				return erro.Wrap(err)
			}
			log.Debug(this.sender, ": Got account info from main ID provider "+unit.idp.Id())
		} else {
			fT, tToA, err = this.getInfoFromSubIdProvider(unit.idp, unit.codTok)
			if err != nil {
				return erro.Wrap(err)
			}
			log.Debug(this.sender, ": Got account info from sub ID provider "+unit.idp.Id())
		}
		for tag, attrs := range tToA {
			if tag == acntTag {
				attrs[tagIss] = unit.idp.Id()
				attrs[tagAt_tag] = tok.Tag()
				attrs[tagAt_exp] = tok.Expires().Unix()
				mainAttrs = attrs
			} else {
				attrs[tagIss] = unit.idp.Id()
				tagToAttrs[tag] = attrs
			}
		}
		if frTa == "" {
			frTa = fT
		} else if frTa != fT {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "two from-TA ID", http.StatusBadRequest, nil))
		}
	}

	log.Debug(this.sender, ": Got all account info")

	jt := jwt.New()
	jt.SetHeader(tagAlg, tagNone)
	for k, v := range mainAttrs {
		jt.SetClaim(k, v)
	}
	mainInfo, err := jt.Encode()
	if err != nil {
		return erro.Wrap(err)
	}

	var relInfo []byte
	sessFlag := true
	if len(tagToAttrs) > 0 {
		jt = jwt.New()
		jt.SetHeader(tagAlg, tagNone)
		for acntTag, attrs := range tagToAttrs {
			jt.SetClaim(acntTag, attrs)
			if sessFlag {
				for k := range attrs {
					if !sessionEnable(k) {
						sessFlag = false
					}
				}
			}
		}
		relInfo, err = jt.Encode()
		if err != nil {
			return erro.Wrap(err)
		}
	}

	w.Header().Set(tagX_auth_user, string(mainInfo))
	w.Header().Set(tagX_auth_user_tag, acntTag)
	w.Header().Set(tagX_auth_from_id, frTa)
	if relInfo != nil {
		w.Header().Set(tagX_auth_users, string(relInfo))
	}
	if sessFlag {
		sessId := this.idGen.String(this.sessLen)
		http.SetCookie(w, this.handler.newCookie(sessId, tok.Expires()))
		log.Debug(this.sender, ": Report session "+logutil.Mosaic(sessId))
	}
	return nil
}

func (this *environment) getInfoFromMainIdProvider(idp idpdb.Element, codTok *codeToken) (frTa string, tok *token.Element, tagToAttrs map[string]map[string]interface{}, err error) {
	return this.getInfo(true, idp, codTok)
}

func (this *environment) getInfoFromSubIdProvider(idp idpdb.Element, codTok *codeToken) (frTa string, tagToAttrs map[string]map[string]interface{}, err error) {
	frTa, _, tagToAttrs, err = this.getInfo(false, idp, codTok)
	return frTa, tagToAttrs, err
}

func (this *environment) getInfo(isMain bool, idp idpdb.Element, codTok *codeToken) (frTa string, tok *token.Element, tagToAttrs map[string]map[string]interface{}, err error) {
	params := map[string]interface{}{}

	// grant_type
	params[tagGrant_type] = tagCooperation_code

	// code
	params[tagCode] = codTok.code()

	// claims
	if isMain {
		// TODO 受け取り方を考えないと。
	}

	// user_claims
	// TODO 受け取り方を考えないと。

	// client_assertion_type
	params[tagClient_assertion_type] = cliAssTypeJwt_bearer

	// client_assertion
	keys, err := this.keyDb.Get()
	if err != nil {
		return "", nil, nil, erro.Wrap(err)
	}
	ass, err := makeAssertion(this.handler, keys, idp.CoopToUri())
	if err != nil {
		return "", nil, nil, erro.Wrap(err)
	}
	params[tagClient_assertion] = string(ass)

	data, err := json.Marshal(params)
	if err != nil {
		return "", nil, nil, erro.Wrap(err)
	}

	r, err := http.NewRequest("POST", idp.CoopToUri(), bytes.NewReader(data))
	r.Header.Set(tagContent_type, contTypeJson)
	log.Debug(this.sender, ": Made main cooperation-to request")

	server.LogRequest(level.DEBUG, r, this.debug)
	resp, err := this.httpClient().Do(r)
	if err != nil {
		return "", nil, nil, erro.Wrap(err)
	}
	defer resp.Body.Close()
	server.LogResponse(level.DEBUG, resp, this.debug)

	if resp.StatusCode != http.StatusOK {
		var buff struct {
			Error             string
			Error_description string
		}
		if err := json.NewDecoder(resp.Body).Decode(&buff); err != nil {
			return "", nil, nil, erro.Wrap(err)
		}
		return "", nil, nil, erro.Wrap(idperr.New(buff.Error, buff.Error_description, resp.StatusCode, nil))
	}
	coopResp, err := parseCoopResponse(resp)
	if err != nil {
		return "", nil, nil, erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
	}

	idsTok, err := parseIdsToken(coopResp.idsToken())
	if err != nil {
		return "", nil, nil, erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
	} else if err := idsTok.verify(idp.Keys()); err != nil {
		return "", nil, nil, erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
	}

	if isMain {
		if coopResp.token() == "" {
			return "", nil, nil, erro.Wrap(idperr.New(idperr.Access_denied, "cannot get token", http.StatusForbidden, nil))
		}
		now := time.Now()
		tok = token.New(coopResp.token(), this.idGen.String(this.tokTagLen), now.Add(coopResp.expiresIn()), idsTok.idProvider(), coopResp.scope())
		log.Info(this.sender, ": Got access token "+logutil.Mosaic(tok.Id()))

		if err := this.tokDb.Save(tok, now.Add(this.tokDbExpIn)); err != nil {
			return "", nil, nil, erro.Wrap(err)
		}
		log.Info(this.sender, ": Saved access token "+logutil.Mosaic(tok.Id()))
	}

	return idsTok.fromTa(), tok, idsTok.attributes(), nil
}

// TA 認証用署名をつくる。
func makeAssertion(hndl *handler, keys []jwk.Key, aud string) ([]byte, error) {
	ass := jwt.New()
	ass.SetHeader(tagAlg, hndl.sigAlg)
	if hndl.sigKid != "" {
		ass.SetHeader(tagKid, hndl.sigKid)
	}
	ass.SetClaim(tagIss, hndl.selfId)
	ass.SetClaim(tagSub, hndl.selfId)
	ass.SetClaim(tagAud, aud)
	ass.SetClaim(tagJti, hndl.idGen.String(hndl.jtiLen))
	now := time.Now()
	ass.SetClaim(tagExp, now.Add(hndl.jtiExpIn).Unix())
	ass.SetClaim(tagIat, now.Unix())
	if err := ass.Sign(keys); err != nil {
		return nil, erro.Wrap(err)
	}
	data, err := ass.Encode()
	if err != nil {
		return nil, erro.Wrap(err)
	}

	return data, nil
}

// セッションを発行しても良い属性かどうか。
func sessionEnable(attrName string) bool {
	return sessEnAttrs[attrName]
}

var sessEnAttrs = strsetutil.New(
	tagIss,
	tagSub,
)
