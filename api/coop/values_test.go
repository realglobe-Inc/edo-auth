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
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

const (
	test_idpSigAlg  = "ES256"
	test_toTaSigAlg = "ES384"

	test_cod = "1SblzkyNc6O867zqdZYPM0T-a7g1n5"
	test_tok = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"

	test_acntTag   = "main-user"
	test_acntId    = "EYClXo4mQKwSgPel"
	test_acntEmail = "tester@example.org"

	test_subAcnt1Tag   = "sub-user1"
	test_subAcnt1Id    = "U7pdvT8dYbBFWXdc"
	test_subAcnt1Email = "subtester1@example.org"
)

var (
	test_idpKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
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
	body, err = json.Marshal(m)
	if err != nil {
		return 0, nil, nil, erro.Wrap(err)
	}

	return http.StatusOK, http.Header{"Content-Type": {"application/json"}}, body, nil
}
