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
	"net/url"
)

type authRequest struct {
	*request.Request

	authUri_ *url.URL
	path_    string
}

func parseAuthRequest(r *http.Request, base *request.Request) (req *authRequest, err error) {
	var authUri *url.URL
	if rawAuthUri := r.Header.Get(tagX_auth_authentication_uri); rawAuthUri == "" {
		return nil, erro.New("no authentication uri")
	} else {
		authUri, err = url.Parse(rawAuthUri)
		if err != nil {
			return nil, erro.Wrap(err)
		}
	}
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	return &authRequest{
		Request:  base,
		authUri_: authUri,
		path_:    path,
	}, nil
}

// 認証機構の URI を返す。
func (this *authRequest) authUri() *url.URL {
	return this.authUri_
}

// リクエストパスを返す。
func (this *authRequest) path() string {
	return this.path_
}
