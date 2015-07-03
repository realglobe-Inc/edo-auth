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
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwk"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"testing"
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
}

// 正常に認証 URI にリダイレクトさせることの検査。
// X-Auth-Uri が ID プロバイダなら response_type を code にすることの検査。
// scope を openid にすることの検査。
// client_id, redirect_uri, state, nonce が追加されることの検査。
// セッションが発行されることの検査。
func TestAuth(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_taKey}, []idpdb.Element{test_idp})

	r, err := newAuthRequest()
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if authUri := uri.Scheme + "://" + uri.Host + uri.Path; authUri != test_idp.AuthUri() {
		t.Error(authUri)
		t.Fatal(test_idp.AuthUri())
	} else if ok, err := regexp.MatchString(page.sessLabel+"=[0-9a-zA-Z_\\-]", w.HeaderMap.Get("Set-Cookie")); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Error("no new session")
		t.Fatal(w.HeaderMap.Get("Set-Cookie"))
	} else if q := uri.Query(); len(q) == 0 {
		t.Fatal("no query")
	} else if respType := request.FormValueSet(q.Get("response_type")); len(respType) != 1 || !respType["code"] {
		t.Error(respType)
		t.Fatal("code")
	} else if scop := request.FormValueSet(q.Get("scope")); len(scop) != 1 || !scop["openid"] {
		t.Error(scop)
		t.Fatal("openid")
	} else if taId := q.Get("client_id"); taId != page.selfId {
		t.Error(taId)
		t.Fatal(page.selfId)
	} else if rediUri := q.Get("redirect_uri"); rediUri != page.rediUri {
		t.Error(rediUri)
		t.Fatal(page.rediUri)
	} else if q.Get("state") == "" {
		t.Fatal("no state")
	} else if q.Get("nonce") == "" {
		t.Fatal("no nonce")
	}
}

// X-Auth-Uri ヘッダが無ければ拒否することの検査。
func TestAuthDenyNoAuthUri(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_taKey}, []idpdb.Element{test_idp})

	r, err := newAuthRequest()
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Del("X-Auth-Uri")

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
}

// X-Auth-Uri が ID プロバイダでないなら response_type を code id_token にすることの検査。
func TestAuthForSelector(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_taKey}, []idpdb.Element{test_idp})

	r, err := newAuthRequest()
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("X-Auth-Uri", test_idp.Id()+"a/auth")

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); len(q) == 0 {
		t.Fatal("no query")
	} else if respType, respType2 := strsetutil.New("code", "id_token"), request.FormValueSet(q.Get("response_type")); !reflect.DeepEqual(respType2, respType) {
		t.Error(respType2)
		t.Fatal(respType)
	}
}
