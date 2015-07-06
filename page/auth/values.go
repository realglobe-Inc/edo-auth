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

const (
	// アンダースコア。
	tagAlg                    = "alg"
	tagAmr                    = "amr"
	tagAt_exp                 = "at_exp"
	tagAt_hash                = "at_hash"
	tagAt_tag                 = "at_tag"
	tagAud                    = "aud"
	tagAuthorization_code     = "authorization_code"
	tagAuthorization_endpoint = "authorization_endpoint"
	tagAzp                    = "azp"
	tagC_hash                 = "c_hash"
	tagClient_assertion       = "client_assertion"
	tagClient_assertion_type  = "client_assertion_type"
	tagClient_id              = "client_id"
	tagCode                   = "code"
	tagExp                    = "exp"
	tagGrant_type             = "grant_type"
	tagIat                    = "iat"
	tagId_token               = "id_token"
	tagIss                    = "iss"
	tagJti                    = "jti"
	tagNbf                    = "nbf"
	tagNonce                  = "nonce"
	tagNone                   = "none"
	tagOpenid                 = "openid"
	tagRedirect_uri           = "redirect_uri"
	tagResponse_type          = "response_type"
	tagScope                  = "scope"
	tagState                  = "state"
	tagSub                    = "sub"
	tagTyp                    = "typ"

	// ハイフン。
	tagNo_cache = "no-cache"
	tagNo_store = "no-store"

	// 頭大文字、ハイフン。
	tagAuthorization = "Authorization"
	tagBearer        = "Bearer"
	tagCache_control = "Cache-Control"
	tagContent_type  = "Content-Type"
	tagPragma        = "Pragma"
	tagX_auth_uri    = "X-Auth-Uri"
	tagX_auth_user   = "X-Auth-User"
)

const (
	cliAssTypeJwt_bearer = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"

	contTypeForm = "application/x-www-form-urlencoded"
)
