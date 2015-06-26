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

import ()

const (
	// アンダースコア。
	tagAlg                   = "alg"
	tagAt_exp                = "at_exp"
	tagAt_tag                = "at_tag"
	tagAud                   = "aud"
	tagClient_assertion      = "client_assertion"
	tagClient_assertion_type = "client_assertion_type"
	tagCode                  = "code"
	tagCode_tokens           = "code_tokens"
	tagCooperation_code      = "cooperation_code"
	tagExp                   = "exp"
	tagGrant_type            = "grant_type"
	tagIat                   = "iat"
	tagIss                   = "iss"
	tagJti                   = "jti"
	tagKid                   = "kid"
	tagNone                  = "none"
	tagSub                   = "sub"

	// 頭大文字、ハイフン。
	tagContent_type            = "Content-Type"
	tagX_auth_from_id          = "X-Auth-From-Id"
	tagX_auth_user             = "X-Auth-User"
	tagX_auth_user_tag         = "X-Auth-User-Tag"
	tagX_auth_users            = "X-Auth-Users"
	tagX_edo_code_tokens       = "X-Edo-Code-Tokens"
	tagX_edo_cooperation_error = "X-Edo-Cooperation-Error"
)

const (
	cliAssTypeJwt_bearer = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"

	contTypeJson = "application/json"
)
