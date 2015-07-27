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
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/test"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
}

var monPool, _ = mgo.DialWithTimeout(test_monAddr, time.Minute)

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

// 正常系。
// アカウント認証代行部分の連携がうまくできていることの検査。
func TestServer(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	if monPool == nil {
		t.SkipNow()
	}

	// ID プロバイダの準備。
	idpServ, err := newTestIdProvider([]jwk.Key{test_idpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer idpServ.close()
	idp := idpServ.info()

	// 環境設定。
	param, err := newTestParameter()
	if err != nil {
		t.Fatal(err)
	}
	param.socPort, err = test.FreePort()
	if err != nil {
		t.Fatal(err)
	}
	// 鍵。
	keyDbPath, err := setupKeyDb(param.keyDbPath, []jwk.Key{test_key})
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(keyDbPath)
	param.keyDbPath = keyDbPath
	// ID プロバイダ DB の準備。
	param.idpDbTag = param.idpDbTag + strconv.FormatInt(time.Now().UnixNano(), 16)
	if err := setupIdpDb(param.idpDbTag, param.idpDbTag2, []idpdb.Element{idp}); err != nil {
		t.Fatal(err)
	}
	defer clearIdpDb(param.idpDbTag, param.idpDbTag2)

	testServer(t, param, idpServ)
}

// データベースをちゃんと使った正常系。
func TestServerWithDb(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	if monPool == nil {
		t.SkipNow()
	}
	red, err := test.NewRedisServer()
	if err != nil {
		t.Fatal(err)
	} else if red == nil {
		t.SkipNow()
	}
	defer red.Close()

	// ID プロバイダの準備。
	idpServ, err := newTestIdProvider([]jwk.Key{test_idpKey})
	if err != nil {
		t.Fatal(err)
	}
	defer idpServ.close()
	idp := idpServ.info()

	// 環境設定。
	param, err := newTestParameter()
	if err != nil {
		t.Fatal(err)
	}
	param.socPort, err = test.FreePort()
	if err != nil {
		t.Fatal(err)
	}
	// 鍵。
	param.keyDbType = "redis"
	param.keyDbAddr = red.Address()
	keyDbPath, err := setupKeyDb(param.keyDbPath, []jwk.Key{test_key})
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(keyDbPath)
	param.keyDbPath = keyDbPath
	// web DB の準備。
	param.webDbType = "redis"
	param.webDbAddr = red.Address()
	// ID プロバイダ DB の準備。
	param.idpDbTag = param.idpDbTag + strconv.FormatInt(time.Now().UnixNano(), 16)
	if err := setupIdpDb(param.idpDbTag, param.idpDbTag2, []idpdb.Element{idp}); err != nil {
		t.Fatal(err)
	}
	defer clearIdpDb(param.idpDbTag, param.idpDbTag2)
	// セッション DB の準備。
	param.asessDbType = "redis"
	param.asessDbAddr = red.Address()
	// アクセストークン DB の準備。
	param.tokDbType = "redis"
	param.tokDbAddr = red.Address()

	testServer(t, param, idpServ)
}

func testServer(t *testing.T, param *parameters, idpServ *testIdProvider) {
	idp := idpServ.info()

	errCh := make(chan error, 1)
	go func() {
		errCh <- serve(param)
	}()
	defer func() { param.shutCh <- struct{}{} }()

	selfUri, err := url.Parse("http://localhost:" + strconv.Itoa(param.socPort))
	if err != nil {
		t.Fatal(err)
	} else if err := waitServer(selfUri.String()+param.pathOk, errCh, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	noRedi := func(req *http.Request, via []*http.Request) error { return errors.New("redirect flag") }

	// 認証前リクエスト。
	reqPath := strings.TrimLeft(param.pathAuth, "/") + "/a/b/c"
	authReq, err := newAuthRequest(selfUri.String()+reqPath, idp.AuthUri())
	if err != nil {
		t.Fatal(err)
	}
	authResp, err := (&http.Client{Jar: cookJar, CheckRedirect: noRedi}).Do(authReq)
	if err != nil {
		if e, ok := err.(*url.Error); !ok || e.Err.Error() != "redirect flag" {
			t.Fatal(err)
		}
	}
	defer authResp.Body.Close()
	server.LogResponse(level.DEBUG, authResp, false)

	if authResp.StatusCode != http.StatusFound {
		t.Error(authResp.StatusCode)
		t.Fatal(http.StatusFound)
	}
	authUri, err := url.Parse(authResp.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	} else if authUri := authUri.Scheme + "://" + authUri.Host + authUri.Path; authUri != idp.AuthUri() {
		t.Error(authUri)
		t.Fatal(idp.AuthUri())
	}
	sessExist := false
	for _, cook := range cookJar.Cookies(selfUri) {
		if cook.Name == param.asessLabel {
			sessExist = true
			break
		}
	}
	if !sessExist {
		t.Fatal("no new session")
	}
	authQ := authUri.Query()
	if len(authQ) == 0 {
		t.Fatal("no query")
	} else if respType := request.FormValueSet(authQ.Get("response_type")); len(respType) != 1 || !respType["code"] {
		t.Error(respType)
		t.Fatal("code")
	} else if scop := request.FormValueSet(authQ.Get("scope")); len(scop) != 1 || !scop["openid"] {
		t.Error(scop)
		t.Fatal("openid")
	} else if taId := authQ.Get("client_id"); taId != param.selfId {
		t.Error(taId)
		t.Fatal(param.selfId)
	} else if rediUri := authQ.Get("redirect_uri"); rediUri != param.rediUri {
		t.Error(rediUri)
		t.Fatal(param.rediUri)
	} else if authQ.Get("state") == "" {
		t.Fatal("no state")
	} else if authQ.Get("nonce") == "" {
		t.Fatal("no nonce")
	}

	// 認証後リクエスト。
	{
		s, h, b, err := newTestTokenResponse(param.selfId, authQ.Get("nonce"), idp)
		if err != nil {
			t.Fatal(err)
		}
		idpServ.addResponse(s, h, b)
	}
	{
		s, h, b, err := newTestAccountResponse(idp)
		if err != nil {
			t.Fatal(err)
		}
		idpServ.addResponse(s, h, b)
	}

	cbReq, err := newCallbackRequest(selfUri.String()+param.pathCb, authQ.Get("state"))
	if err != nil {
		t.Fatal(err)
	}
	cbResp, err := (&http.Client{Jar: cookJar, CheckRedirect: noRedi}).Do(cbReq)
	if err != nil {
		if e, ok := err.(*url.Error); !ok || e.Err.Error() != "redirect flag" {
			t.Fatal(err)
		}
	}
	defer cbResp.Body.Close()
	server.LogResponse(level.DEBUG, cbResp, false)

	if cbResp.StatusCode != http.StatusFound {
		t.Error(cbResp.StatusCode)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(cbResp.Header.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != reqPath {
		t.Error(uri.Path)
		t.Fatal(reqPath)
	}
	var buff struct {
		Iss    string
		Sub    string
		At_tag string
		At_exp int
		Email  string
	}
	if jt, err := jwt.Parse([]byte(cbResp.Header.Get("X-Auth-User"))); err != nil {
		t.Fatal(err)
	} else if err := json.Unmarshal(jt.RawBody(), &buff); err != nil {
		t.Fatal(err)
	} else if buff.Iss != idp.Id() {
		t.Error(buff.Iss)
		t.Fatal(idp.Id())
	} else if buff.Sub != test_acntId {
		t.Error(buff.Sub)
		t.Fatal(test_acntId)
	} else if len(buff.At_tag) != param.tokTagLen {
		t.Error(len(buff.At_tag), buff.At_tag)
		t.Fatal(param.tokTagLen)
	} else if buff.At_exp == 0 {
		t.Fatal("no at_exp")
	} else if buff.Email != test_acntEmail {
		t.Error(buff.Email)
		t.Fatal(test_acntEmail)
	}
}

func setupKeyDb(path string, keys []jwk.Key) (dir string, err error) {
	dir, err = ioutil.TempDir(filepath.Dir(path), filepath.Base(path))
	for i, key := range keys {
		if data, err := json.Marshal(key.ToMap()); err != nil {
			return "", erro.Wrap(err)
		} else if ioutil.WriteFile(filepath.Join(dir, strconv.Itoa(i)+".json"), data, 0644); err != nil {
			return "", erro.Wrap(err)
		}
	}
	return dir, nil
}

func setupIdpDb(db, coll string, idps []idpdb.Element) error {
	conn := monPool.New()
	defer conn.Close()

	for _, idp := range idps {
		keys := []map[string]interface{}{}
		for _, key := range idp.Keys() {
			keys = append(keys, key.ToMap())
		}
		m := bson.M{
			"issuer":                    idp.Id(),
			"authorization_endpoint":    idp.AuthUri(),
			"token_endpoint":            idp.TokenUri(),
			"userinfo_endpoint":         idp.AccountUri(),
			"cooperation_from_endpoint": idp.CoopFromUri(),
			"cooperation_to_endpoint":   idp.CoopToUri(),
			"jwks": keys,
		}
		for k, v := range idp.Names() {
			if k == "" {
				m["issuer_name"] = v
			} else {
				m["issuer_name#"+k] = v
			}
		}
		if err := conn.DB(db).C(coll).Insert(m); err != nil {
			return erro.Wrap(err)
		}
	}
	return nil
}

func clearIdpDb(db, coll string) error {
	conn := monPool.New()
	defer conn.Close()

	return conn.DB(db).DropDatabase()
}

func waitServer(uri string, errCh chan error, exp time.Time) error {
	for {
		if time.Now().After(exp) {
			return erro.New("timeout")
		}
		r, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			return erro.Wrap(err)
		}
		r.Header.Set("Connection", "close")
		if _, err := http.DefaultClient.Do(r); err == nil {
			break
		}

		select {
		case err := <-errCh:
			return erro.Wrap(err)
		default:
		}
		time.Sleep(time.Millisecond)
	}
	return nil
}
