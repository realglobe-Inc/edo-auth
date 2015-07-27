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

package coop

import (
	"encoding/json"
	"net/http"
	"time"

	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/go-lib/erro"
)

const (
	test_idpSigAlg = "ES256"
	test_cod       = "1SblzkyNc6O867zqdZYPM0T-a7g1n5"
	test_tok       = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"

	test_subIdpSigAlg = "ES256"
	test_subCod       = "iIxcrSU4j4ulTBIy4bOVjX9Epe5k5a"

	test_refHash = "q1wfMtaj8DL_Z2iyDPHEuA"

	test_toTaSigAlg = "ES384"

	test_acntTag   = "main-user"
	test_acntId    = "EYClXo4mQKwSgPel"
	test_acntEmail = "tester@example.org"

	test_subAcnt1Tag   = "sub-user1"
	test_subAcnt1Id    = "U7pdvT8dYbBFWXdc"
	test_subAcnt1Email = "subtester1@example.org"

	test_subAcnt2Tag   = "sub-user2"
	test_subAcnt2Id    = "lgmxuHfXfSTB-1js"
	test_subAcnt2Email = "subtester2@example.org"
)

var (
	test_idpKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
	})
	test_subIdpKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "vQ3EYqVi30Zd4NF0hbKdHIMZAngSrhwa3mxx74zXkDc",
		"y":   "OwPvhvTL0SlgB7SpucwBOyjbbY0V8M1-dS6FwkMPGD8",
		"d":   "Y4YXo4D_B5FMj_5oXizubBDWRWETRpWr8jX969odblA",
	})
	test_toTaKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
		"d":   "Gp-7eC0G7PjGzKoiAmTQ1iLsLU3AEy3h-bKFWSZOanXqSWI6wqJVPEUsatNYBJoG",
	})
	test_frTa = tadb.New("https://from.example.org", nil, nil, nil, false, "")
)

// コードトークン 1 つのリクエスト。
func newTestSingleRequest(hndl *handler, idp idpdb.Element) (*http.Request, error) {
	return newTestSingleRequestWithParams(hndl, idp, nil)
}

func newTestSingleRequestWithParams(hndl *handler, idp idpdb.Element, params map[string]interface{}) (*http.Request, error) {
	r, err := http.NewRequest("GET", "http://localhost/coop", nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	codTok := jwt.New()
	codTok.SetHeader("alg", test_idpSigAlg)
	codTok.SetClaim("iss", idp.Id())
	codTok.SetClaim("sub", test_cod)
	codTok.SetClaim("aud", audience.New(hndl.selfId))
	codTok.SetClaim("from_client", test_frTa.Id())
	codTok.SetClaim("user_tag", test_acntTag)
	codTok.SetClaim("user_tags", []string{test_subAcnt1Tag})
	for k, v := range params {
		codTok.SetClaim(k, v)
	}
	if err := codTok.Sign(idp.Keys()); err != nil {
		return nil, erro.Wrap(err)
	}
	data, err := codTok.Encode()
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Set("X-Edo-Code-Tokens", string(data))

	return r, nil
}

// コードトークン 1 つの場合の ID プロバイダからのレスポンス。
func newTestSingleIdpResponse(hndl *handler, idp idpdb.Element) (status int, header http.Header, body []byte, err error) {
	return newTestMainIdpResponse(hndl, idp)
}

func newTestSingleIdpResponseWithParams(hndl *handler, idp idpdb.Element, params, idsTokParams map[string]interface{}) (status int, header http.Header, body []byte, err error) {
	return newTestMainIdpResponseWithParams(hndl, idp, params, idsTokParams)
}

// コードトークン 2 つ以上のリクエスト。
func newTestRequest(hndl *handler, idp, subIdp idpdb.Element) (*http.Request, error) {
	return newTestRequestWithParams(hndl, idp, subIdp, nil, nil)
}

func newTestRequestWithParams(hndl *handler, idp, subIdp idpdb.Element, params, subParams map[string]interface{}) (*http.Request, error) {
	r, err := http.NewRequest("GET", "http://localhost/coop", nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	codTok := jwt.New()
	codTok.SetHeader("alg", test_idpSigAlg)
	codTok.SetClaim("iss", idp.Id())
	codTok.SetClaim("sub", test_cod)
	codTok.SetClaim("aud", audience.New(hndl.selfId))
	codTok.SetClaim("from_client", test_frTa.Id())
	codTok.SetClaim("user_tag", test_acntTag)
	codTok.SetClaim("user_tags", []string{test_subAcnt1Tag})
	codTok.SetClaim("ref_hash", test_refHash)
	for k, v := range params {
		codTok.SetClaim(k, v)
	}
	if err := codTok.Sign(idp.Keys()); err != nil {
		return nil, erro.Wrap(err)
	}
	data, err := codTok.Encode()
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Set("X-Edo-Code-Tokens", string(data))

	subCodTok := jwt.New()
	subCodTok.SetHeader("alg", test_subIdpSigAlg)
	subCodTok.SetClaim("iss", subIdp.Id())
	subCodTok.SetClaim("sub", test_subCod)
	subCodTok.SetClaim("aud", audience.New(hndl.selfId))
	subCodTok.SetClaim("user_tags", []string{test_subAcnt2Tag})
	subCodTok.SetClaim("ref_hash", test_refHash)
	for k, v := range subParams {
		subCodTok.SetClaim(k, v)
	}
	if err := subCodTok.Sign(subIdp.Keys()); err != nil {
		return nil, erro.Wrap(err)
	}
	subData, err := subCodTok.Encode()
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Add("X-Edo-Code-Tokens", string(subData))

	return r, nil
}

func newTestMainIdpResponse(hndl *handler, idp idpdb.Element) (status int, header http.Header, body []byte, err error) {
	return newTestMainIdpResponseWithParams(hndl, idp, nil, nil)
}

func newTestMainIdpResponseWithParams(hndl *handler, idp idpdb.Element, params, idsTokParams map[string]interface{}) (status int, header http.Header, body []byte, err error) {
	now := time.Now()

	idsTok := jwt.New()
	idsTok.SetHeader("alg", test_idpSigAlg)
	idsTok.SetClaim("iss", idp.Id())
	idsTok.SetClaim("sub", test_frTa.Id())
	idsTok.SetClaim("aud", audience.New(hndl.selfId))
	idsTok.SetClaim("exp", now.Add(time.Minute).Unix())
	idsTok.SetClaim("iat", now.Unix())
	idsTok.SetClaim("ids", map[string]map[string]interface{}{
		test_acntTag: {
			"sub":   test_acntId,
			"email": test_acntEmail,
		},
		test_subAcnt1Tag: {
			"sub":   test_subAcnt1Id,
			"email": test_subAcnt1Email,
		},
	})
	for k, v := range idsTokParams {
		idsTok.SetClaim(k, v)
	}
	if err := idsTok.Sign(idp.Keys()); err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}
	data, err := idsTok.Encode()
	if err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}

	m := map[string]interface{}{
		"access_token": test_tok,
		"token_type":   "Bearer",
		"expires_in":   1234,
		"scope":        "openid email",
		"ids_token":    string(data),
	}
	for k, v := range params {
		if v == nil {
			delete(m, k)
		} else {
			m[k] = v
		}
	}
	body, err = json.Marshal(m)
	if err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}

	return http.StatusOK, http.Header{"Content-Type": {"application/json"}}, body, nil
}

func newTestSubIdpResponse(hndl *handler, idp idpdb.Element) (status int, header http.Header, body []byte, err error) {
	return newTestSubIdpResponseWithParams(hndl, idp, nil, nil)
}

func newTestSubIdpResponseWithParams(hndl *handler, idp idpdb.Element, params, idsTokParams map[string]interface{}) (status int, header http.Header, body []byte, err error) {
	now := time.Now()

	idsTok := jwt.New()
	idsTok.SetHeader("alg", test_idpSigAlg)
	idsTok.SetClaim("iss", idp.Id())
	idsTok.SetClaim("sub", test_frTa.Id())
	idsTok.SetClaim("aud", audience.New(hndl.selfId))
	idsTok.SetClaim("exp", now.Add(time.Minute).Unix())
	idsTok.SetClaim("iat", now.Unix())
	idsTok.SetClaim("ids", map[string]map[string]interface{}{
		test_subAcnt2Tag: {
			"sub":   test_subAcnt2Id,
			"email": test_subAcnt2Email,
		},
	})
	for k, v := range idsTokParams {
		idsTok.SetClaim(k, v)
	}
	if err := idsTok.Sign(idp.Keys()); err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}
	data, err := idsTok.Encode()
	if err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}

	m := map[string]interface{}{
		"ids_token": string(data),
	}
	for k, v := range params {
		if v == nil {
			delete(m, k)
		} else {
			m[k] = v
		}
	}
	body, err = json.Marshal(m)
	if err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}

	return http.StatusOK, http.Header{"Content-Type": {"application/json"}}, body, nil
}
