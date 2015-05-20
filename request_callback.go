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
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
)

type callbackRequest struct {
	*request.Request

	cod   string
	stat  string
	idTok []byte
}

func parseCallbackRequest(r *http.Request, base *request.Request) (*callbackRequest, error) {
	cod := r.FormValue(tagCode)
	if cod == "" {
		return nil, erro.New("no code")
	}
	stat := r.FormValue(tagState)
	if stat == "" {
		return nil, erro.New("no state")
	}
	var idTok []byte
	if rawIdTok := r.FormValue(tagId_token); rawIdTok != "" {
		idTok = []byte(rawIdTok)
	}
	return &callbackRequest{base, cod, stat, idTok}, nil
}

// 認可コードを返す。
func (this *callbackRequest) code() string {
	return this.cod
}

// state を返す。
func (this *callbackRequest) state() string {
	return this.stat
}

// ID トークンを返す。
func (this *callbackRequest) idToken() []byte {
	return this.idTok
}
