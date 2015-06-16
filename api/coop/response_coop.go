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
	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

type coopResponse struct {
	tok     string
	tokType string
	expIn   time.Duration
	scop    map[string]bool
	idsTok  []byte
}

func parseCoopResponse(resp *http.Response) (*coopResponse, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, erro.New("invalid state ", resp.StatusCode)
	} else if contType := resp.Header.Get(tagContent_type); contType != contTypeJson {
		return nil, erro.New("invalid content type " + contType)
	}

	var buff struct {
		Tok     string `json:"access_token"`
		TokType string `json:"token_type"`
		ExpIn   int    `json:"expires_in"`
		Scop    string `json:"scope"`
		IdsTok  string `json:"ids_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&buff); err != nil {
		return nil, erro.Wrap(err)
	} else if buff.IdsTok == "" {
		return nil, erro.New("cannot get IDs token")
	}

	return &coopResponse{
		tok:     buff.Tok,
		tokType: buff.TokType,
		expIn:   time.Duration(buff.ExpIn) * time.Second,
		scop:    requtil.FormValueSet(buff.Scop),
		idsTok:  []byte(buff.IdsTok),
	}, nil
}

func (this *coopResponse) token() string {
	return this.tok
}

func (this *coopResponse) tokenType() string {
	return this.tokType
}

func (this *coopResponse) expiresIn() time.Duration {
	return this.expIn
}

func (this *coopResponse) scope() map[string]bool {
	return this.scop
}

func (this *coopResponse) idsToken() []byte {
	return this.idsTok
}
