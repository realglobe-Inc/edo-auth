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
	"net/http"
	"time"

	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/go-lib/erro"
)

type tokenResponse struct {
	tok   string
	exp   time.Time
	scop  map[string]bool
	idTok []byte
}

func parseTokenResponse(resp *http.Response) (*tokenResponse, error) {
	var buff struct {
		Access_token string
		Expires_in   int
		Scope        string
		Id_token     string
	}
	if err := json.NewDecoder(resp.Body).Decode(&buff); err != nil {
		return nil, erro.Wrap(err)
	} else if buff.Access_token == "" {
		return nil, erro.New("no access token")
	}

	var exp time.Time
	if buff.Expires_in != 0 {
		exp = time.Now().Add(time.Duration(buff.Expires_in) * time.Second)
	}
	var scop map[string]bool
	if buff.Scope != "" {
		scop = request.FormValueSet(buff.Scope)
	}
	var idTok []byte
	if buff.Id_token != "" {
		idTok = []byte(buff.Id_token)
	}

	return &tokenResponse{
		tok:   buff.Access_token,
		exp:   exp,
		scop:  scop,
		idTok: idTok,
	}, nil
}

func (this *tokenResponse) token() string {
	return this.tok
}

func (this *tokenResponse) expires() time.Time {
	return this.exp
}

func (this *tokenResponse) scope() map[string]bool {
	return this.scop
}

func (this *tokenResponse) idToken() []byte {
	return this.idTok
}
