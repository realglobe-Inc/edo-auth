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
	"crypto/tls"
	"encoding/json"
	"github.com/realglobe-Inc/edo-auth/database/token"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net"
	"net/http"
	"time"
)

type handler struct {
	stopper *server.Stopper

	selfId string
	sigAlg string
	sigKid string

	sessLabel  string
	tokTagLen  int
	tokDbExpIn time.Duration
	jtiLen     int
	jtiExpIn   time.Duration

	keyDb keydb.Db
	idpDb idpdb.Db
	tokDb token.Db
	idGen rand.Generator

	noVeri bool
	debug  bool
}

func New(
	stopper *server.Stopper,
	selfId string,
	sigAlg string,
	sigKid string,
	sessLabel string,
	tokTagLen int,
	tokDbExpIn time.Duration,
	jtiLen int,
	jtiExpIn time.Duration,
	keyDb keydb.Db,
	idpDb idpdb.Db,
	tokDb token.Db,
	idGen rand.Generator,
	noVeri bool,
	debug bool,
) http.Handler {
	return &handler{
		stopper:    stopper,
		selfId:     selfId,
		sigAlg:     sigAlg,
		sigKid:     sigKid,
		sessLabel:  sessLabel,
		tokTagLen:  tokTagLen,
		tokDbExpIn: tokDbExpIn,
		jtiLen:     jtiLen,
		jtiExpIn:   jtiExpIn,
		keyDb:      keyDb,
		idpDb:      idpDb,
		tokDb:      tokDb,
		idGen:      idGen,
		noVeri:     noVeri,
		debug:      debug,
	}
}

// http.DefaultTransport を参考にした。
var noVeriTr = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 10 * time.Second,
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
}

func (this *handler) httpClient() *http.Client {
	if this.noVeri {
		return &http.Client{Transport: noVeriTr}
	} else {
		return &http.Client{}
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

	sender = requtil.Parse(r, this.sessLabel)
	log.Info(sender, ": Received cooperation request")
	defer log.Info(sender, ": Handled cooperation request")

	if err := this.serve(w, r, sender); err != nil {
		idperr.RespondJson(w, r, erro.Wrap(err), sender)
		return
	}
}

func (this *handler) serve(w http.ResponseWriter, r *http.Request, sender *requtil.Request) error {
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
			log.Debug(sender, ": Main account tag is "+acntTag)
		}

		for tag := range codTok.accountTags() {
			if tags[tag] {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "tag "+tag+" overlaps", http.StatusBadRequest, nil))
			}
			tags[tag] = true
			log.Debug(sender, ": Account tag is "+tag)
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

		log.Debug(sender, ": ID provider "+idp.Id()+" is exist")

		if err := codTok.verify(idp.Keys()); err != nil {
			return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
		}

		log.Debug(sender, ": Verified cooperation code")

		units = append(units, &idpUnit{idp, codTok})
	}
	if acntTag == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no main account tag", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Cooperation codes are OK")

	var tok *token.Element
	var mainAttrs map[string]interface{}
	tagToAttrs := map[string]map[string]interface{}{}
	var frTa string
	for _, unit := range units {
		var tToA map[string]map[string]interface{}
		var fT string
		if unit.codTok.accountTag() != "" {
			fT, tok, tToA, err = this.getInfoFromMainIdProvider(unit.idp, unit.codTok, sender)
			if err != nil {
				return erro.Wrap(err)
			}
			log.Debug(sender, ": Got account info from main ID provider "+unit.idp.Id())
		} else {
			fT, tToA, err = this.getInfoFromSubIdProvider(sender)
			if err != nil {
				return erro.Wrap(err)
			}
			log.Debug(sender, ": Got account info from sub ID provider "+unit.idp.Id())
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

	log.Debug(sender, ": Got all account info")

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
	if len(tagToAttrs) > 0 {
		jt = jwt.New()
		jt.SetHeader(tagAlg, tagNone)
		for k, v := range tagToAttrs {
			jt.SetClaim(k, v)
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

	return nil
}

func (this *handler) getInfoFromMainIdProvider(idp idpdb.Element, codTok *codeToken, sender *requtil.Request) (frTa string, tok *token.Element, tagToAttrs map[string]map[string]interface{}, err error) {
	params := map[string]interface{}{}

	// grant_type
	params[tagGrant_type] = tagCooperation_code

	// code
	params[tagCode] = codTok.code()

	// claims
	// TODO 受け取り方を考えないと。

	// user_claims
	// TODO 受け取り方を考えないと。

	// client_assertion_type
	params[tagClient_assertion_type] = cliAssTypeJwt_bearer

	// client_assertion
	keys, err := this.keyDb.Get()
	if err != nil {
		return "", nil, nil, erro.Wrap(err)
	}

	{
		jt := jwt.New()
		jt.SetHeader(tagAlg, this.sigAlg)
		if this.sigKid != "" {
			jt.SetHeader(tagKid, this.sigKid)
		}
		jt.SetClaim(tagIss, this.selfId)
		jt.SetClaim(tagSub, this.selfId)
		jt.SetClaim(tagAud, idp.CoopToUri())
		jt.SetClaim(tagJti, this.idGen.String(this.jtiLen))
		now := time.Now()
		jt.SetClaim(tagExp, now.Add(this.jtiExpIn).Unix())
		jt.SetClaim(tagIat, now.Unix())
		if err := jt.Sign(keys); err != nil {
			return "", nil, nil, erro.Wrap(err)
		}
		assData, err := jt.Encode()
		if err != nil {
			return "", nil, nil, erro.Wrap(err)
		}
		params[tagClient_assertion] = string(assData)
	}

	data, err := json.Marshal(params)
	if err != nil {
		return "", nil, nil, erro.Wrap(err)
	}

	r, err := http.NewRequest("POST", idp.CoopToUri(), bytes.NewReader(data))
	r.Header.Set(tagContent_type, contTypeJson)
	log.Debug(sender, ": Made main cooperation-to request")

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

	now := time.Now()
	tok = token.New(coopResp.token(), this.idGen.String(this.tokTagLen), now.Add(coopResp.expiresIn()), idsTok.idProvider(), coopResp.scope())
	log.Info(sender, ": Got access token "+logutil.Mosaic(tok.Id()))

	if err := this.tokDb.Save(tok, now.Add(this.tokDbExpIn)); err != nil {
		return "", nil, nil, erro.Wrap(err)
	}
	log.Info(sender, ": Saved access token "+logutil.Mosaic(tok.Id()))

	return idsTok.fromTa(), tok, idsTok.attributes(), nil
}

func (this *handler) getInfoFromSubIdProvider(sender *requtil.Request) (frTa string, tagToAttrs map[string]map[string]interface{}, err error) {
	panic("not yet implemented")
}
