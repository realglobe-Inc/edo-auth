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
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestAuthRequest(t *testing.T) {
	q := url.Values{}
	q.Set("client_id", "https://ta.example.org")
	q.Set("redirect_uri", "https://ta.example.org/callback")
	authUri, err := url.Parse("https://idp.example.org/auth?" + q.Encode())
	path := "/a/b/c?a=b&c=d"
	r, err := http.NewRequest("GET", "https://ta.example.org"+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("X-Auth-Uri", authUri.String())

	req, err := parseAuthRequest(r, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(req.authUri(), authUri) {
		t.Error(req.authUri())
		t.Fatal(authUri)
	} else if req.path() != path {
		t.Error(req.path())
		t.Fatal(path)
	}
}
