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
	"crypto/tls"
	"github.com/realglobe-Inc/edo-auth/api/coop"
	"github.com/realglobe-Inc/edo-auth/database/asession"
	"github.com/realglobe-Inc/edo-auth/database/token"
	authpage "github.com/realglobe-Inc/edo-auth/page/auth"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	"github.com/realglobe-Inc/edo-lib/driver"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog"
	"html/template"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	var exitCode = 0
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()
	defer rglog.Flush()

	logutil.InitConsole(logRoot)

	param, err := parseParameters(os.Args...)
	if err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(erro.Wrap(err))
		exitCode = 1
		return
	}

	logutil.SetupConsole(logRoot, param.consLv)
	if err := logutil.Setup(logRoot, param.logType, param.logLv, param); err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(erro.Wrap(err))
		exitCode = 1
		return
	}

	if err := serve(param); err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(erro.Wrap(err))
		exitCode = 1
		return
	}

	log.Info("Shut down")
}

func serve(param *parameters) (err error) {

	// バックエンドの準備。

	stopper := server.NewStopper()

	redPools := driver.NewRedisPoolSet(param.redTimeout, param.redPoolSize, param.redPoolExpIn)
	defer redPools.Close()
	monPools := driver.NewMongoPoolSet(param.monTimeout)
	defer monPools.Close()

	// 鍵。
	var keyDb keydb.Db
	switch param.keyDbType {
	case "file":
		keyDb = keydb.NewFileDb(param.keyDbPath)
		log.Info("Use keys in directory " + param.keyDbPath)
	case "redis":
		keyDb = keydb.NewRedisCache(keydb.NewFileDb(param.keyDbPath), redPools.Get(param.keyDbAddr), param.keyDbTag+"."+param.selfId, param.keyDbExpIn)
		log.Info("Use keys in directory " + param.keyDbPath + " with redis " + param.keyDbAddr + ": " + param.keyDbTag + "." + param.selfId)
	default:
		return erro.New("invalid key DB type " + param.keyDbType)
	}

	// web データ。
	var webDb webdb.Db
	switch param.webDbType {
	case "direct":
		webDb = webdb.NewDirectDb()
		log.Info("Get web data directly")
	case "redis":
		webDb = webdb.NewRedisCache(webdb.NewDirectDb(), redPools.Get(param.webDbAddr), param.webDbTag, param.webDbExpIn)
		log.Info("Get web data with redis " + param.webDbAddr + ": " + param.webDbTag)
	default:
		return erro.New("invalid web data DB type " + param.webDbType)
	}

	// IdP 情報。
	var idpDb idpdb.Db
	switch param.idpDbType {
	case "mongo":
		pool, err := monPools.Get(param.idpDbAddr)
		if err != nil {
			return erro.Wrap(err)
		}
		idpDb = idpdb.NewMongoDb(pool, param.idpDbTag, param.idpDbTag2, webDb)
		log.Info("Use IdP info in mongodb " + param.idpDbAddr + ": " + param.idpDbTag + "." + param.idpDbTag2)
	default:
		return erro.New("invalid IdP DB type " + param.idpDbType)
	}

	// セッション。
	var asessDb asession.Db
	switch param.asessDbType {
	case "memory":
		asessDb = asession.NewMemoryDb()
		log.Info("Save user sessions in memory")
	case "redis":
		asessDb = asession.NewRedisDb(redPools.Get(param.asessDbAddr), param.asessDbTag)
		log.Info("Save user sessions in redis " + param.asessDbAddr + ": " + param.asessDbTag)
	default:
		return erro.New("invalid user session DB type " + param.asessDbType)
	}

	// アクセストークン。
	var tokDb token.Db
	switch param.tokDbType {
	case "memory":
		tokDb = token.NewMemoryDb()
		log.Info("Save access tokens in memory")
	case "redis":
		tokDb = token.NewRedisDb(redPools.Get(param.tokDbAddr), param.tokDbTag)
		log.Info("Save access tokens in redis " + param.tokDbAddr + ": " + param.tokDbTag)
	default:
		return erro.New("invalid access token DB type " + param.tokDbType)
	}

	var errTmpl *template.Template
	if param.tmplErr != "" {
		errTmpl, err = template.ParseFiles(param.tmplErr)
		if err != nil {
			return erro.Wrap(err)
		}
	}

	idGen := rand.New(time.Minute)

	var tr http.RoundTripper
	if param.noVeri {
		// http.DefaultTransport を参考にした。
		tr = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		}
	}

	// バックエンドの準備完了。

	if param.debug {
		server.Debug = true
	}

	authPage := authpage.New(
		stopper,
		param.selfId,
		param.rediUri,
		param.sigAlg,
		errTmpl,
		param.asessLabel,
		param.asessLen,
		param.asessExpIn,
		param.asessDbExpIn,
		param.fsessLabel,
		param.fsessLen,
		param.fsessExpIn,
		param.statLen,
		param.noncLen,
		param.tokTagLen,
		param.tokDbExpIn,
		param.jtiLen,
		param.jtiExpIn,
		keyDb,
		idpDb,
		asessDb,
		tokDb,
		idGen,
		param.cookPath,
		param.cookSec,
		param.debug,
	)

	mux := http.NewServeMux()
	routes := map[string]bool{}
	mux.HandleFunc(param.pathOk, server.WrapPage(stopper, func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}, errTmpl))
	routes[param.pathOk] = true
	mux.HandleFunc(param.pathAuth, authPage.HandleAuth)
	routes[param.pathAuth] = true
	mux.HandleFunc(param.pathCb, authPage.HandleCallback)
	routes[param.pathCb] = true
	mux.Handle(param.pathCoop, coop.New(
		stopper,
		param.selfId,
		param.sigAlg,
		param.sigKid,
		param.csessLabel,
		param.csessLen,
		param.tokTagLen,
		param.tokDbExpIn,
		param.jtiLen,
		param.jtiExpIn,
		keyDb,
		idpDb,
		tokDb,
		idGen,
		tr,
		param.cookPath,
		param.cookSec,
		param.debug,
	))
	routes[param.pathCoop] = true

	if !routes["/"] {
		mux.HandleFunc("/", server.WrapPage(stopper, func(w http.ResponseWriter, r *http.Request) error {
			return erro.Wrap(server.NewError(http.StatusNotFound, "invalid endpoint", nil))
		}, errTmpl))
	}

	// サーバー設定完了。

	defer func() {
		// 処理の終了待ち。
		stopper.Lock()
		defer stopper.Unlock()
		for stopper.Stopped() {
			stopper.Wait()
		}
	}()
	return server.Serve(mux, param.socType, param.protType, param)
}
