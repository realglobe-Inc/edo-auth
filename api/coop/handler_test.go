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
	"github.com/realglobe-Inc/edo-auth/database/token"
	"github.com/realglobe-Inc/edo-id-provider/claims"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/test"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

func newTestHandler(keys []jwk.Key, idps []idpdb.Element) *handler {
	return New(
		server.NewStopper(),
		"https://to.example.org",
		test_toTaSigAlg,
		"",
		"Edo-Cooperation",
		10,
		time.Hour,
		10,
		time.Minute,
		keydb.NewMemoryDb(keys),
		idpdb.NewMemoryDb(idps),
		token.NewMemoryDb(),
		rand.New(time.Minute),
		nil,
		true,
	).(*handler)
}

type testIdProvider struct {
	base *test.HttpServer
	keys []jwk.Key
}

func newTestIdProvider(keys []jwk.Key) (*testIdProvider, error) {
	base, err := test.NewHttpServer(time.Minute)
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

// 仲介コードが 1 つの場合の正常系。
// レスポンスが X-Auth-User, X-Auth-User-Tag, X-Auth-Users, X-Auth-From-Id を含むことの検査。
// 主体情報が iss, sub, at_tag, at_exp クレームを含むことの検査。
// 主体でないアカウント情報が iss, sub クレームを含むことの検査。
func TestSingleNormal(t *testing.T) {
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
	hndl := newTestHandler([]jwk.Key{test_toTaKey}, []idpdb.Element{idp})

	r, err := newTestSingleRequest(hndl, idp)
	if err != nil {
		t.Fatal(err)
	}

	s, h, b, err := newTestSingleIdpResponse(hndl, idp)
	if err != nil {
		t.Fatal(err)
	}
	reqCh := idpServ.addResponse(s, h, b)

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	select {
	case req := <-reqCh:
		if contType, contType2 := "application/json", req.Header.Get("Content-Type"); contType2 != contType {
			t.Error(contType)
			t.Fatal(contType2)
		}
		var buff struct {
			Grant_type  string
			Code        string
			Claims      *claims.Request
			User_claims claims.Claims
		}
		if err := json.NewDecoder(req.Body).Decode(&buff); err != nil {
			t.Fatal(err)
		} else if grntType := "cooperation_code"; buff.Grant_type != grntType {
			t.Error(buff.Grant_type)
			t.Fatal(grntType)
		} else if buff.Code != test_cod {
			t.Error(buff.Code)
			t.Fatal(test_cod)
		}
	case <-time.After(time.Minute):
		t.Fatal("no request")
	}

	if w.Code != http.StatusOK {
		t.Error(w.Code)
		t.Fatal(http.StatusOK)
	}
	acntJt, err := jwt.Parse([]byte(w.HeaderMap.Get("X-Auth-User")))
	if err != nil {
		t.Fatal(err)
	}
	var acntBuff struct {
		Iss    string
		Sub    string
		At_tag string
		At_exp int64
		Email  string
	}
	if err := json.Unmarshal(acntJt.RawBody(), &acntBuff); err != nil {
		t.Fatal(err)
	} else if acntBuff.Iss != idp.Id() {
		t.Error(acntBuff.Iss)
		t.Fatal(idp.Id())
	} else if acntBuff.Sub != test_acntId {
		t.Error(acntBuff.Sub)
		t.Fatal(test_acntId)
	} else if acntBuff.At_tag == "" {
		t.Fatal("no token tag")
	} else if now, exp := time.Now(), time.Unix(acntBuff.At_exp, 0); exp.Before(now) {
		t.Error("expired")
		t.Error(now)
		t.Fatal(exp)
	} else if acntBuff.Email != test_acntEmail {
		t.Error(acntBuff.Email)
		t.Fatal(test_acntEmail)
	}
	acntTag := w.HeaderMap.Get("X-Auth-User-Tag")
	if acntTag != test_acntTag {
		t.Error(acntTag)
		t.Fatal(test_acntTag)
	}
	acntsJt, err := jwt.Parse([]byte(w.HeaderMap.Get("X-Auth-Users")))
	if err != nil {
		t.Fatal(err)
	}
	type account struct {
		Iss   string
		Sub   string
		Email string
	}
	var acntsBuff map[string]*account
	if err := json.Unmarshal(acntsJt.RawBody(), &acntsBuff); err != nil {
		t.Fatal(err)
	} else if len(acntsBuff) == 0 {
		t.Fatal("no accounts")
	} else if subAcnt := acntsBuff[test_subAcnt1Tag]; subAcnt == nil {
		t.Fatal("no sub account")
	} else if subAcnt.Iss != idp.Id() {
		t.Error(subAcnt.Iss)
		t.Fatal(idp.Id())
	} else if subAcnt.Sub != test_subAcnt1Id {
		t.Error(subAcnt.Sub)
		t.Fatal(test_subAcnt1Id)
	} else if subAcnt.Email != test_subAcnt1Email {
		t.Error(subAcnt.Email)
		t.Fatal(test_subAcnt1Email)
	}
}

// ID プロバイダからエラーが返されたら、エラーを返せることの検査。
func TestDenyIdProviderError(t *testing.T) {
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
	hndl := newTestHandler([]jwk.Key{test_toTaKey}, []idpdb.Element{idp})

	r, err := newTestSingleRequest(hndl, idp)
	if err != nil {
		t.Fatal(err)
	}

	s := http.StatusForbidden
	b := []byte(`{"error":"access_denied","error_description":"you are alone"}`)
	reqCh := idpServ.addResponse(s, http.Header{"Content-Type": {"application/json"}}, b)

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	select {
	case <-reqCh:
	case <-time.After(time.Minute):
		t.Fatal("no request")
	}

	if w.Code != s {
		t.Error(w.Code)
		t.Fatal(s)
	}

	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "access_denied"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 仲介コードが 1 つ以上の場合の正常系。
// レスポンスが X-Auth-User, X-Auth-User-Tag, X-Auth-Users, X-Auth-From-Id を含むことの検査。
// 主体情報が iss, sub, at_tag, at_exp クレームを含むことの検査。
// 主体でないアカウント情報が iss, sub クレームを含むことの検査。
func TestMultiNormal(t *testing.T) {
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
	subIdpServ, err := newTestIdProvider([]jwk.Key{test_subIdpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer subIdpServ.close()
	subIdp := subIdpServ.info()
	hndl := newTestHandler([]jwk.Key{test_toTaKey}, []idpdb.Element{idp, subIdp})

	r, err := newTestRequest(hndl, idp, subIdp)
	if err != nil {
		t.Fatal(err)
	}

	var reqCh <-chan *http.Request
	{
		s, h, b, err := newTestMainIdpResponse(hndl, idp)
		if err != nil {
			t.Fatal(err)
		}
		reqCh = idpServ.addResponse(s, h, b)
	}
	var subReqCh <-chan *http.Request
	{
		s, h, b, err := newTestSubIdpResponse(hndl, subIdp)
		if err != nil {
			t.Fatal(err)
		}
		subReqCh = subIdpServ.addResponse(s, h, b)
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	select {
	case req := <-reqCh:
		if contType, contType2 := "application/json", req.Header.Get("Content-Type"); contType2 != contType {
			t.Error(contType)
			t.Fatal(contType2)
		}
		var buff struct {
			Grant_type  string
			Code        string
			Claims      *claims.Request
			User_claims claims.Claims
		}
		if err := json.NewDecoder(req.Body).Decode(&buff); err != nil {
			t.Fatal(err)
		} else if grntType := "cooperation_code"; buff.Grant_type != grntType {
			t.Error(buff.Grant_type)
			t.Fatal(grntType)
		} else if buff.Code != test_cod {
			t.Error(buff.Code)
			t.Fatal(test_cod)
		}
	case <-time.After(time.Minute):
		t.Fatal("no request")
	}
	select {
	case req := <-subReqCh:
		if contType, contType2 := "application/json", req.Header.Get("Content-Type"); contType2 != contType {
			t.Error(contType)
			t.Fatal(contType2)
		}
		var buff struct {
			Grant_type  string
			Code        string
			Claims      *claims.Request
			User_claims claims.Claims
		}
		if err := json.NewDecoder(req.Body).Decode(&buff); err != nil {
			t.Fatal(err)
		} else if grntType := "cooperation_code"; buff.Grant_type != grntType {
			t.Error(buff.Grant_type)
			t.Fatal(grntType)
		} else if buff.Code != test_subCod {
			t.Error(buff.Code)
			t.Fatal(test_subCod)
		}
	case <-time.After(time.Minute):
		t.Fatal("no request")
	}

	if w.Code != http.StatusOK {
		t.Error(w.Code)
		t.Fatal(http.StatusOK)
	}
	acntJt, err := jwt.Parse([]byte(w.HeaderMap.Get("X-Auth-User")))
	if err != nil {
		t.Fatal(err)
	}
	var acntBuff struct {
		Iss    string
		Sub    string
		At_tag string
		At_exp int64
		Email  string
	}
	if err := json.Unmarshal(acntJt.RawBody(), &acntBuff); err != nil {
		t.Fatal(err)
	} else if acntBuff.Iss != idp.Id() {
		t.Error(acntBuff.Iss)
		t.Fatal(idp.Id())
	} else if acntBuff.Sub != test_acntId {
		t.Error(acntBuff.Sub)
		t.Fatal(test_acntId)
	} else if acntBuff.At_tag == "" {
		t.Fatal("no token tag")
	} else if now, exp := time.Now(), time.Unix(acntBuff.At_exp, 0); exp.Before(now) {
		t.Error("expired")
		t.Error(now)
		t.Fatal(exp)
	} else if acntBuff.Email != test_acntEmail {
		t.Error(acntBuff.Email)
		t.Fatal(test_acntEmail)
	}
	acntTag := w.HeaderMap.Get("X-Auth-User-Tag")
	if acntTag != test_acntTag {
		t.Error(acntTag)
		t.Fatal(test_acntTag)
	}
	acntsJt, err := jwt.Parse([]byte(w.HeaderMap.Get("X-Auth-Users")))
	if err != nil {
		t.Fatal(err)
	}
	type account struct {
		Iss   string
		Sub   string
		Email string
	}
	var acntsBuff map[string]*account
	if err := json.Unmarshal(acntsJt.RawBody(), &acntsBuff); err != nil {
		t.Fatal(err)
	} else if len(acntsBuff) == 0 {
		t.Fatal("no accounts")
	} else if subAcnt1 := acntsBuff[test_subAcnt1Tag]; subAcnt1 == nil {
		t.Fatal("no sub account")
	} else if subAcnt1.Iss != idp.Id() {
		t.Error(subAcnt1.Iss)
		t.Fatal(idp.Id())
	} else if subAcnt1.Sub != test_subAcnt1Id {
		t.Error(subAcnt1.Sub)
		t.Fatal(test_subAcnt1Id)
	} else if subAcnt1.Email != test_subAcnt1Email {
		t.Error(subAcnt1.Email)
		t.Fatal(test_subAcnt1Email)
	} else if subAcnt2 := acntsBuff[test_subAcnt2Tag]; subAcnt2 == nil {
		t.Fatal("no sub account")
	} else if subAcnt2.Iss != subIdp.Id() {
		t.Error(subAcnt2.Iss)
		t.Fatal(subIdp.Id())
	} else if subAcnt2.Sub != test_subAcnt2Id {
		t.Error(subAcnt2.Sub)
		t.Fatal(test_subAcnt2Id)
	} else if subAcnt2.Email != test_subAcnt2Email {
		t.Error(subAcnt2.Email)
		t.Fatal(test_subAcnt2Email)
	}
}
