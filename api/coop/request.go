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
	"net/http"

	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/go-lib/erro"
)

type request struct {
	codToks [][]byte
}

func parseRequest(r *http.Request) (*request, error) {
	rawCodToks := r.Header[tagX_edo_code_tokens]
	if rawCodToks == nil {
		rawCodToks = requtil.FormValues(r.FormValue(tagCode_tokens))
	}
	if len(rawCodToks) == 0 {
		return nil, erro.New("no cooperation codes")
	}
	codToks := [][]byte{}
	for _, rawCodTok := range rawCodToks {
		codToks = append(codToks, []byte(rawCodTok))
	}
	return &request{codToks}, nil
}

func (this *request) codeTokens() [][]byte {
	return this.codToks
}
