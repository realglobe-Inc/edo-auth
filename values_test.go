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
	"net/http"
	"net/url"
	"time"

	hashutil "github.com/realglobe-Inc/edo-id-provider/hash"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/go-lib/erro"
)

const (
	test_logPath = "/tmp/edo-access-proxy.log"
	test_logSize = 100000000
	test_logNum  = 10
	test_logAddr = "127.0.0.1:24224"
	test_logTag  = "edo-access-proxy"
	test_socPort = 12345
	test_socPath = "/tmp/edo-access-proxy.sock"
	test_monAddr = "localhost"

	test_idpSigAlg = "ES256"
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
	test_key, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
		"d":   "Gp-7eC0G7PjGzKoiAmTQ1iLsLU3AEy3h-bKFWSZOanXqSWI6wqJVPEUsatNYBJoG",
	})
)

func newAuthRequest(uri, idpUri string) (*http.Request, error) {
	r, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Set("X-Auth-Uri", idpUri)
	r.Header.Set("Connection", "close")
	return r, nil
}

func newCallbackRequest(uri, stat string) (*http.Request, error) {
	q := url.Values{}
	q.Set("code", test_cod)
	q.Set("state", stat)
	r, err := http.NewRequest("GET", uri+"?"+q.Encode(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Set("Connection", "close")
	return r, nil
}

func newTestTokenResponse(selfId, nonc string, idp idpdb.Element) (status int, hader http.Header, body []byte, err error) {
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
	idTok.SetClaim("aud", selfId)
	now := time.Now()
	idTok.SetClaim("exp", now.Add(time.Minute).Unix())
	idTok.SetClaim("iat", now.Unix())
	idTok.SetClaim("nonce", nonc)
	hGen := jwt.HashGenerator(test_idpSigAlg)
	if !hGen.Available() {
		return 0, nil, nil, erro.New("unsupported algorithm " + test_idpSigAlg)
	}
	idTok.SetClaim("at_hash", hashutil.Hashing(hGen.New(), []byte(test_tok)))
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

func newTestAccountResponse(idp idpdb.Element) (status int, hader http.Header, body []byte, err error) {
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
