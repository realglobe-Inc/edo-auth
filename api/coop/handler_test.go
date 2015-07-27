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
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

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
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
}

func newTestHandler(keys []jwk.Key, idps []idpdb.Element) *handler {
	return New(
		server.NewStopper(),
		"https://to.example.org",
		test_toTaSigAlg,
		"",
		"Edo-Cooperation",
		30,
		10,
		time.Hour,
		10,
		time.Minute,
		keydb.NewMemoryDb(keys),
		idpdb.NewMemoryDb(idps),
		token.NewMemoryDb(),
		rand.New(time.Minute),
		http.DefaultClient,
		"/",
		false,
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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

// 仲介コードが 2 つ以上の場合の正常系。
// レスポンスが X-Auth-User, X-Auth-User-Tag, X-Auth-Users, X-Auth-From-Id を含むことの検査。
// 主体情報が iss, sub, at_tag, at_exp クレームを含むことの検査。
// 主体でないアカウント情報が iss, sub クレームを含むことの検査。
func TestMultiNormal(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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

// 仲介データの JWT の署名がおかしかったらエラーを返すことの検査。
func TestDenyInvalidSign(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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
	r.Header.Set("X-Edo-Code-Tokens", regexp.MustCompile("\\.[^.]+$").ReplaceAllString(r.Header.Get("X-Edo-Code-Tokens"), ".AAAA"))

	s, h, b, err := newTestSingleIdpResponse(hndl, idp)
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s, h, b)

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 仲介データに iss が無かったらエラーを返すことの検査。
func TestDenyNoIss(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSingleDenyNoSomething(t, "iss")
}

// 仲介データに sub が無かったらエラーを返すことの検査。
func TestDenyNoSub(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSingleDenyNoSomething(t, "sub")
}

// 仲介データに aud が無かったらエラーを返すことの検査。
func TestDenyNoAud(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSingleDenyNoSomething(t, "aud")
}

// 仲介コードが 1 つの場合に、仲介データに from_client が無かったらエラーを返すことの検査。
func TestSingleDenyNoFromClient(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSingleDenyNoSomething(t, "from_client")
}

// 仲介コードが 1 つの場合に、仲介データに user_tag が無かったらエラーを返すことの検査。
func TestSingleDenyNoUserTag(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSingleDenyNoSomething(t, "user_tag")
}

func testSingleDenyNoSomething(t *testing.T, something string) {
	idpServ, err := newTestIdProvider([]jwk.Key{test_idpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer idpServ.close()
	idp := idpServ.info()
	hndl := newTestHandler([]jwk.Key{test_toTaKey}, []idpdb.Element{idp})

	r, err := newTestSingleRequestWithParams(hndl, idp, map[string]interface{}{something: nil})
	if err != nil {
		t.Fatal(err)
	}

	s, h, b, err := newTestSingleIdpResponse(hndl, idp)
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s, h, b)

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 仲介コードが 2 つ以上の場合に、仲介データに ref_hash が無かったらエラーを返すことの検査。
func TestMultiDenyNoRefHash(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testMultiDenyNoSomething(t, "ref_hash", "")
	testMultiDenyNoSomething(t, "", "ref_hash")
}

// 仲介コードが 2 つ以上の場合に、仲介データに user_tags が無かったらエラーを返すことの検査。
func TestMultiDenyNoUserTags(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testMultiDenyNoSomething(t, "", "user_tags")
}

func testMultiDenyNoSomething(t *testing.T, something1, something2 string) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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

	r, err := newTestRequestWithParams(hndl, idp, subIdp, map[string]interface{}{something1: nil}, map[string]interface{}{something2: nil})
	if err != nil {
		t.Fatal(err)
	}

	{
		s, h, b, err := newTestMainIdpResponse(hndl, idp)
		if err != nil {
			t.Fatal(err)
		}
		idpServ.addResponse(s, h, b)
	}
	{
		s, h, b, err := newTestSubIdpResponse(hndl, subIdp)
		if err != nil {
			t.Fatal(err)
		}
		subIdpServ.addResponse(s, h, b)
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// アカウントタグが重複したらエラーを返すことの検査。
func TestMultiDenyAccountTagOverlap(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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

	r, err := newTestRequestWithParams(hndl, idp, subIdp, nil, map[string]interface{}{"user_tags": []string{test_subAcnt1Tag}})
	if err != nil {
		t.Fatal(err)
	}

	{
		s, h, b, err := newTestMainIdpResponse(hndl, idp)
		if err != nil {
			t.Fatal(err)
		}
		idpServ.addResponse(s, h, b)
	}
	{
		s, h, b, err := newTestSubIdpResponse(hndl, subIdp)
		if err != nil {
			t.Fatal(err)
		}
		subIdpServ.addResponse(s, h, b)
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// セッションを発行できることの検査。
func TestSession(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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

	{
		s, h, b, err := newTestMainIdpResponseWithParams(hndl, idp, nil, map[string]interface{}{
			"ids": map[string]map[string]interface{}{
				test_acntTag: {
					"sub": test_acntId,
				},
				test_subAcnt1Tag: {
					"sub": test_subAcnt1Id,
				},
			}})
		if err != nil {
			t.Fatal(err)
		}
		idpServ.addResponse(s, h, b)
	}
	{
		s, h, b, err := newTestSubIdpResponseWithParams(hndl, subIdp, nil, map[string]interface{}{
			"ids": map[string]map[string]interface{}{
				test_subAcnt2Tag: {
					"sub": test_subAcnt2Id,
				},
			}})
		if err != nil {
			t.Fatal(err)
		}
		subIdpServ.addResponse(s, h, b)
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Error(w.Code)
		t.Fatal(http.StatusOK)
	} else if cook := w.HeaderMap.Get("Set-Cookie"); cook == "" {
		t.Fatal("no session")
	}
}

// ID プロバイダからアクセストークンが帰って来なかったら拒否できること検査。
func TestDenyNoAccessTokenFromIdProvider(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testDenyNoSomethingFromIdProvider(t, "access_token", "")
}

// ID プロバイダから ids_token が帰って来なかったら拒否できること検査。
func TestDenyNoIdsTokenFromIdProvider(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testDenyNoSomethingFromIdProvider(t, "ids_token", "")
}

// ID プロバイダから ids_token の iss が帰って来なかったら拒否できること検査。
func TestDenyNoIdsTokenIssFromIdProvider(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testDenyNoSomethingFromIdProvider(t, "", "iss")
}

// ID プロバイダから ids_token の sub が帰って来なかったら拒否できること検査。
func TestDenyNoIdsTokenSubFromIdProvider(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testDenyNoSomethingFromIdProvider(t, "", "sub")
}

// ID プロバイダから ids_token の aud が帰って来なかったら拒否できること検査。
func TestDenyNoIdsTokenAudFromIdProvider(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testDenyNoSomethingFromIdProvider(t, "", "aud")
}

// ID プロバイダから ids_token の exp が帰って来なかったら拒否できること検査。
func TestDenyNoIdsTokenExpFromIdProvider(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testDenyNoSomethingFromIdProvider(t, "", "exp")
}

// ID プロバイダから ids_token の iat が帰って来なかったら拒否できること検査。
func TestDenyNoIdsTokenIatFromIdProvider(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testDenyNoSomethingFromIdProvider(t, "", "iat")
}

// ID プロバイダから ids_token の ids が帰って来なかったら拒否できること検査。
func TestDenyNoIdsTokenIdsFromIdProvider(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testDenyNoSomethingFromIdProvider(t, "", "ids")
}

func testDenyNoSomethingFromIdProvider(t *testing.T, something1, something2 string) {
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

	s, h, b, err := newTestSingleIdpResponseWithParams(hndl, idp, map[string]interface{}{something1: nil}, map[string]interface{}{something2: nil})
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s, h, b)

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Error(w.Code)
		t.Fatal(http.StatusForbidden)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "access_denied"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// ID プロバイダから要求した主体の情報が帰って来なかったら拒否できること検査。
func TestDenyNoMainAccountFromIdProvider(t *testing.T) {
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

	s, h, b, err := newTestSingleIdpResponseWithParams(hndl, idp, nil, map[string]interface{}{
		"ids": map[string]map[string]interface{}{
			test_subAcnt1Tag: {
				"sub":   test_subAcnt1Id,
				"email": test_subAcnt1Email,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s, h, b)

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Error(w.Code)
		t.Fatal(http.StatusForbidden)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "access_denied"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// ID プロバイダから要求した主体でないアカウント情報が帰って来なかったら拒否できること検査。
func TestDenyNoSubAccountFromIdProvider(t *testing.T) {
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

	s, h, b, err := newTestSingleIdpResponseWithParams(hndl, idp, nil, map[string]interface{}{
		"ids": map[string]map[string]interface{}{
			test_acntTag: {
				"sub":   test_acntId,
				"email": test_acntEmail,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	idpServ.addResponse(s, h, b)

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Error(w.Code)
		t.Fatal(http.StatusForbidden)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "access_denied"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}
