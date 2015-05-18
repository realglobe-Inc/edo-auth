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

import ()

const (
	// コンパイル時に打ち間違いに気付けるように。
	// それ以上の意味は無い。

	// HTTP ヘッダ。
	tagAuthorization             = "Authorization"
	tagCache_control             = "Cache-Control"
	tagContent_length            = "Content-Length"
	tagContent_type              = "Content-Type"
	tagPragma                    = "Pragma"
	tagX_auth_authentication_uri = "X-Auth-Authentication-Uri"
	tagX_auth_user               = "X-Auth-User"

	// HTTP ヘッダ値。
	tagBearer   = "Bearer"
	tagNo_cache = "no-cache"
	tagNo_store = "no-store"

	// HTTP フォーム。
	tagClient_assertion      = "client_assertion"
	tagClient_assertion_type = "client_assertion_type"
	tagClient_id             = "client_id"
	tagCode                  = "code"
	tagGrant_type            = "grant_type"
	tagNonce                 = "nonce"
	tagRedirect_uri          = "redirect_uri"
	tagResponse_type         = "response_type"
	tagScope                 = "scope"
	tagState                 = "state"

	tagStatus  = "status"
	tagMessage = "message"
	tagDetail  = "detail"

	// パラメータ値。
	tagAuthorization_code = "authorization_code"
	//tagCode = "code"
	tagId_token = "id_token"
	tagOpenid   = "openid"

	// JWT ヘッダ。
	tagAlg = "alg"

	// JWT ヘッダ値。
	tagNone = "none"

	// JWT クレーム。
	tagAmr     = "amr"
	tagAt_hash = "at_hash"
	tagAud     = "aud"
	tagAzp     = "azp"
	tagC_hash  = "c_hash"
	tagExp     = "exp"
	tagIat     = "iat"
	tagIss     = "iss"
	tagJti     = "jti"
	tagNbf     = "nbf"
	//tagNonce = "nonce"
	tagSub = "sub"
	tagTyp = "typ"

	tagAt_exp = "at_exp"
	tagAt_tag = "at_tag"

	// ID プロバイダ属性。
	tagAuthorization_endpoint = "authorization_endpoint"

	// セッション。
	tagAuth_user = "Auth-User"
)

const (
	cliAssTypeJwt_bearer = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)
