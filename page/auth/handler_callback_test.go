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
	"github.com/realglobe-Inc/edo-auth/database/usession"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/test"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func init() {
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

type testIdProvider struct {
	base *test.HttpServer
	keys []jwk.Key
}

func newTestIdProvider(keys []jwk.Key) (*testIdProvider, error) {
	base, err := test.NewHttpServer(time.Second)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return &testIdProvider{base, keys}, nil
}

func (this *testIdProvider) close() {
	this.base.Close()
}

func (this *testIdProvider) info() idpdb.Element {
	return idpdb.New(
		this.base.URL,
		nil,
		this.base.URL+"/auth",
		this.base.URL+"/token",
		this.base.URL+"/userinfo",
		this.base.URL+"/coop/from",
		this.base.URL+"/coop/to",
		this.keys,
	)
}

func (this *testIdProvider) selfId() string {
	return this.base.URL
}

func (this *testIdProvider) addResponse(status int, header http.Header, body []byte) <-chan *http.Request {
	return this.base.AddResponse(status, header, body)
}

// 正常系。
// 元のリクエストパスにリダイレクトさせることの検査。
// X-Auth-User ヘッダに iss, sub, at_tag, at_exp の入った JWT を入れることの検査。
// X-Auth-User ヘッダに追加属性を入れることの検査。
func TestCallback(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	idpServ, err := newTestIdProvider([]jwk.Key{test_idpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer idpServ.close()
	idp := idpServ.info()
	page := newTestPage([]jwk.Key{test_taKey}, []idpdb.Element{idp})

	now := time.Now()
	sess := usession.New(test_sessId, now.Add(page.sessExpIn), test_reqPath, idp.Id(), page.selfId, page.rediUri, test_stat, test_nonc)
	page.sessDb.Save(sess, now.Add(time.Minute))

	r, err := newCallbackRequest(page)
	if err != nil {
		t.Fatal(err)
	}

	s1, h1, b1, err := newTestTokenResponse(page, idp, nil)
	if err != nil {
		t.Fatal(err)
	}
	req1Ch := idpServ.addResponse(s1, h1, b1)
	s2, h2, b2, err := newTestAccountResponse(page, idp)
	if err != nil {
		t.Fatal(err)
	}
	req2Ch := idpServ.addResponse(s2, h2, b2)

	w := httptest.NewRecorder()
	page.HandleCallback(w, r)

	select {
	case req := <-req1Ch:
		if contType, contType2 := "application/x-www-form-urlencoded", req.Header.Get("Content-Type"); contType2 != contType {
			t.Error(contType)
			t.Fatal(contType2)
		} else if grntType, grntType2 := "authorization_code", req.FormValue("grant_type"); grntType2 != grntType {
			t.Error(grntType)
			t.Fatal(grntType2)
		} else if cod := req.FormValue("code"); cod != test_cod {
			t.Error(cod)
			t.Fatal(test_cod)
		} else if rediUri := req.FormValue("redirect_uri"); rediUri != page.rediUri {
			t.Error(rediUri)
			t.Fatal(page.rediUri)
		} else if taId := req.FormValue("client_id"); taId != page.selfId {
			t.Error(taId)
			t.Fatal(page.selfId)
		} else if assType, assType2 := "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", req.FormValue("client_assertion_type"); assType2 != assType {
			t.Error(assType2)
			t.Fatal(assType)
		}
		ass, err := jwt.Parse([]byte(req.FormValue("client_assertion")))
		if err != nil {
			t.Fatal(err)
		} else if !ass.IsSigned() {
			t.Fatal("not signed")
		} else if err := ass.Verify([]jwk.Key{test_taKey}); err != nil {
			t.Fatal(err)
		}
		var buff struct {
			Iss string
			Sub string
			Aud audience.Audience
			Jti string
			Exp int
			Iat int
		}
		if err := json.Unmarshal(ass.RawBody(), &buff); err != nil {
			t.Fatal(err)
		} else if buff.Iss != page.selfId {
			t.Error(buff.Iss)
			t.Fatal(page.selfId)
		} else if buff.Sub != page.selfId {
			t.Error(buff.Sub)
			t.Fatal(page.selfId)
		} else if !buff.Aud[idp.TokenUri()] {
			t.Error(buff.Aud)
			t.Fatal(idp.TokenUri())
		} else if len(buff.Jti) != page.jtiLen {
			t.Error(len(buff.Jti), " "+buff.Jti)
			t.Fatal(page.jtiLen)
		} else if buff.Exp == 0 {
			t.Fatal("no exp")
		} else if buff.Iat == 0 {
			t.Fatal("no iat")
		} else if !(buff.Iat < buff.Exp) {
			t.Error("exp not after iat")
			t.Error(buff.Iat)
			t.Fatal(buff.Exp)
		}
	case <-time.After(time.Second):
		t.Fatal("no request")
	}

	select {
	case req := <-req2Ch:
		if auth := strings.Fields(req.Header.Get("Authorization")); len(auth) != 2 {
			t.Error("not 2 fields")
			t.Fatal(auth)
		} else if auth[0] != "Bearer" {
			t.Error(auth[0])
			t.Fatal("Bearer")
		} else if auth[1] != test_tok {
			t.Error(auth[1])
			t.Fatal(test_tok)
		}
	case <-time.After(time.Second):
		t.Fatal("no request")
	}

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(test_reqPath); err != nil {
		t.Fatal(err)
	} else if uri2, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(uri2, uri) {
		t.Error(uri2)
		t.Fatal(uri)
	}
	var buff struct {
		Iss    string
		Sub    string
		At_tag string
		At_exp int
		Email  string
	}
	if jt, err := jwt.Parse([]byte(w.HeaderMap.Get("X-Auth-User"))); err != nil {
		t.Fatal(err)
	} else if err := json.Unmarshal(jt.RawBody(), &buff); err != nil {
		t.Fatal(err)
	} else if buff.Iss != idp.Id() {
		t.Error(buff.Iss)
		t.Fatal(idp.Id())
	} else if buff.Sub != test_acntId {
		t.Error(buff.Sub)
		t.Fatal(test_acntId)
	} else if len(buff.At_tag) != page.tokTagLen {
		t.Error(len(buff.At_tag), buff.At_tag)
		t.Fatal(page.tokTagLen)
	} else if buff.At_exp == 0 {
		t.Fatal("no at_exp")
	} else if buff.Email != test_acntEmail {
		t.Error(buff.Email)
		t.Fatal(test_acntEmail)
	}
}

// 選択サービス経由のリクエストに対応できることの検査。
func TestCallbackThroughSelector(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	idpServ, err := newTestIdProvider([]jwk.Key{test_idpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer idpServ.close()
	idp := idpServ.info()
	page := newTestPage([]jwk.Key{test_taKey}, []idpdb.Element{idp})

	now := time.Now()
	sess := usession.New(test_sessId, now.Add(page.sessExpIn), test_reqPath, "", page.selfId, page.rediUri, test_stat, test_nonc)
	page.sessDb.Save(sess, now.Add(time.Minute))

	r, err := newCallbackRequestWithIdToken(page, idp, nil)
	if err != nil {
		t.Fatal(err)
	}

	s1, h1, b1, err := newTestTokenResponse(page, idp, nil)
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s1, h1, b1)
	s2, h2, b2, err := newTestAccountResponse(page, idp)
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s2, h2, b2)

	w := httptest.NewRecorder()
	page.HandleCallback(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(test_reqPath); err != nil {
		t.Fatal(err)
	} else if uri2, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(uri2, uri) {
		t.Error(uri2)
		t.Fatal(uri)
	}
}

// state がおかしいなら拒否できることの検査。
func TestCallbackDenyInvalidState(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	idpServ, err := newTestIdProvider([]jwk.Key{test_idpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer idpServ.close()
	idp := idpServ.info()
	page := newTestPage([]jwk.Key{test_taKey}, []idpdb.Element{idp})

	now := time.Now()
	sess := usession.New(test_sessId, now.Add(page.sessExpIn), test_reqPath, idp.Id(), page.selfId, page.rediUri, test_stat, test_nonc)
	page.sessDb.Save(sess, now.Add(time.Minute))

	r, err := newCallbackRequest(page)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		q.Set("state", test_stat+"a")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleCallback(w, r)

	if w.Code != http.StatusForbidden {
		t.Error(w.Code)
		t.Fatal(http.StatusForbidden)
	}
}

// nonce がおかしいなら拒否できることの検査。
func TestCallbackDenyInvalidNonce(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	idpServ, err := newTestIdProvider([]jwk.Key{test_idpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer idpServ.close()
	idp := idpServ.info()
	page := newTestPage([]jwk.Key{test_taKey}, []idpdb.Element{idp})

	now := time.Now()
	sess := usession.New(test_sessId, now.Add(page.sessExpIn), test_reqPath, idp.Id(), page.selfId, page.rediUri, test_stat, test_nonc)
	page.sessDb.Save(sess, now.Add(time.Minute))

	r, err := newCallbackRequest(page)
	if err != nil {
		t.Fatal(err)
	}

	s1, h1, b1, err := newTestTokenResponse(page, idp, map[string]interface{}{"nonce": test_nonc + "a"})
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s1, h1, b1)
	s2, h2, b2, err := newTestAccountResponse(page, idp)
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s2, h2, b2)

	w := httptest.NewRecorder()
	page.HandleCallback(w, r)

	if w.Code != http.StatusForbidden {
		t.Error(w.Code)
		t.Fatal(http.StatusForbidden)
	}
}

// c_hash がおかしいなら拒否できることの検査。
func TestCallbackDenyInvalidCodeHash(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	idpServ, err := newTestIdProvider([]jwk.Key{test_idpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer idpServ.close()
	idp := idpServ.info()
	page := newTestPage([]jwk.Key{test_taKey}, []idpdb.Element{idp})

	now := time.Now()
	sess := usession.New(test_sessId, now.Add(page.sessExpIn), test_reqPath, "", page.selfId, page.rediUri, test_stat, test_nonc)
	page.sessDb.Save(sess, now.Add(time.Minute))

	r, err := newCallbackRequestWithIdToken(page, idp, map[string]interface{}{"c_hash": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"})
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	page.HandleCallback(w, r)

	if w.Code != http.StatusForbidden {
		t.Error(w.Code)
		t.Fatal(http.StatusForbidden)
	}
}

// at_hash がおかしいなら拒否できることの検査。
func TestCallbackDenyInvalidTokenHash(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	idpServ, err := newTestIdProvider([]jwk.Key{test_idpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer idpServ.close()
	idp := idpServ.info()
	page := newTestPage([]jwk.Key{test_taKey}, []idpdb.Element{idp})

	now := time.Now()
	sess := usession.New(test_sessId, now.Add(page.sessExpIn), test_reqPath, idp.Id(), page.selfId, page.rediUri, test_stat, test_nonc)
	page.sessDb.Save(sess, now.Add(time.Minute))

	r, err := newCallbackRequest(page)
	if err != nil {
		t.Fatal(err)
	}

	s1, h1, b1, err := newTestTokenResponse(page, idp, map[string]interface{}{"at_hash": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"})
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s1, h1, b1)

	w := httptest.NewRecorder()
	page.HandleCallback(w, r)

	if w.Code != http.StatusForbidden {
		t.Error(w.Code)
		t.Fatal(http.StatusForbidden)
	}
}
