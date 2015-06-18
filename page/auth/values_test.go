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
	"encoding/json"
	hashutil "github.com/realglobe-Inc/edo-id-provider/hash"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
	"time"
)

const (
	test_idpSigAlg = "ES256"
	test_taSigAlg  = "ES384"

	test_reqPath   = "/a/b/c/d/e"
	test_sessId    = "EBBR9STJ-hUL9PzniRasrc-qeFxJ9m"
	test_stat      = "YJgUit_Wx5"
	test_nonc      = "Wjj1_YUOlR"
	test_cod       = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	test_tok       = "TM4CmjXyWQeqtasbRDqwSN80n26vuV"
	test_acntId    = "EYClXo4mQKwSgPel"
	test_acntEmail = "tester@example.org"
)

var (
	test_idpKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
	})
	test_taKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
		"d":   "Gp-7eC0G7PjGzKoiAmTQ1iLsLU3AEy3h-bKFWSZOanXqSWI6wqJVPEUsatNYBJoG",
	})
	test_idp = idpdb.New(
		"https://idp.example.org", nil,
		"https://idp.example.org/auth", "", "", "", "", nil,
	)
)

func newAuthRequest() (*http.Request, error) {
	r, err := http.NewRequest("GET", "http://localhost/", nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Set("X-Auth-Uri", test_idp.AuthUri())
	return r, nil
}

func newCallbackRequest(page *Page) (*http.Request, error) {
	q := url.Values{}
	q.Set("code", test_cod)
	q.Set("state", test_stat)
	r, err := http.NewRequest("GET", "http://localhost/callback?"+q.Encode(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: test_sessId,
	})
	return r, nil
}

func newCallbackRequestWithIdToken(page *Page, idp idpdb.Element, clms map[string]interface{}) (*http.Request, error) {
	q := url.Values{}
	q.Set("code", test_cod)
	q.Set("state", test_stat)
	idTok := jwt.New()
	idTok.SetHeader("alg", test_idpSigAlg)
	idTok.SetClaim("iss", idp.Id())
	idTok.SetClaim("sub", test_acntId)
	idTok.SetClaim("aud", page.selfId)
	now := time.Now()
	idTok.SetClaim("exp", now.Add(time.Minute).Unix())
	idTok.SetClaim("iat", now.Unix())
	idTok.SetClaim("nonce", test_nonc)
	hash, err := jwt.HashGenerator(test_idpSigAlg)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	idTok.SetClaim("c_hash", hashutil.Hashing(hash.New(), []byte(test_cod)))
	for k, v := range clms {
		idTok.SetClaim(k, v)
	}
	if err := idTok.Sign(idp.Keys()); err != nil {
		return nil, erro.Wrap(err)
	}
	data, err := idTok.Encode()
	if err != nil {
		return nil, erro.Wrap(err)
	}
	q.Set("id_token", string(data))
	r, err := http.NewRequest("GET", "http://localhost/callback?"+q.Encode(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: test_sessId,
	})
	return r, nil
}

func newTestTokenResponse(page *Page, idp idpdb.Element, clms map[string]interface{}) (status int, hader http.Header, body []byte, err error) {
	m := map[string]interface{}{
		"access_token": test_tok,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        "openid email",
	}
	idTok := jwt.New()
	idTok.SetHeader("alg", test_idpSigAlg)
	idTok.SetClaim("iss", idp.Id())
	idTok.SetClaim("sub", test_acntId)
	idTok.SetClaim("aud", page.selfId)
	now := time.Now()
	idTok.SetClaim("exp", now.Add(time.Minute).Unix())
	idTok.SetClaim("iat", now.Unix())
	idTok.SetClaim("nonce", test_nonc)
	hash, err := jwt.HashGenerator(test_idpSigAlg)
	if err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}
	idTok.SetClaim("at_hash", hashutil.Hashing(hash.New(), []byte(test_tok)))
	for k, v := range clms {
		idTok.SetClaim(k, v)
	}
	if err := idTok.Sign(idp.Keys()); err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}
	data, err := idTok.Encode()
	if err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}
	m["id_token"] = string(data)

	body, err = json.Marshal(m)
	if err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}

	return http.StatusOK, nil, body, nil
}

func newTestAccountResponse(page *Page, idp idpdb.Element) (status int, hader http.Header, body []byte, err error) {
	m := map[string]interface{}{
		"sub":   test_acntId,
		"email": test_acntEmail,
	}
	body, err = json.Marshal(m)
	if err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}

	return http.StatusOK, nil, body, nil
}
