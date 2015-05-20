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
	"github.com/realglobe-Inc/edo-auth/database/token"
	"github.com/realglobe-Inc/edo-auth/database/usession"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/driver"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"html/template"
	"net/http"
	"os"
	"strings"
)

func main() {
	var exitCode = 0
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()
	defer rglog.Flush()

	logutil.InitConsole("github.com/realglobe-Inc")

	param, err := parseParameters(os.Args...)
	if err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(erro.Wrap(err))
		exitCode = 1
		return
	}

	logutil.SetupConsole("github.com/realglobe-Inc", param.consLv)
	if err := logutil.Setup("github.com/realglobe-Inc", param.logType, param.logLv, param); err != nil {
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
		keyDb = keydb.NewRedisCache(keydb.NewFileDb(param.keyDbPath), redPools.Get(param.keyDbAddr), param.keyDbTag+"."+param.taId, param.keyDbExpIn)
		log.Info("Use keys in directory " + param.keyDbPath + " with redis " + param.keyDbAddr + ": " + param.keyDbTag + "." + param.taId)
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
	var usessDb usession.Db
	switch param.usessDbType {
	case "memory":
		usessDb = usession.NewMemoryDb()
		log.Info("Save user sessions in memory")
	case "redis":
		usessDb = usession.NewRedisDb(redPools.Get(param.usessDbAddr), param.usessDbTag)
		log.Info("Save user sessions in redis " + param.usessDbAddr + ": " + param.usessDbTag)
	default:
		return erro.New("invalid user session DB type " + param.usessDbType)
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

	sys := &system{
		param.taId,
		param.rediUri,
		param.sigAlg,
		param.sigKid,

		errTmpl,

		param.usessLabel,
		param.usessLen,
		param.authExpIn,
		param.usessExpIn,
		param.usessDbExpIn,
		param.statLen,
		param.noncLen,
		param.tokTagLen,
		param.tokDbExpIn,
		param.jtiLen,
		param.jtiExpIn,

		keyDb,
		webDb,
		idpDb,
		usessDb,
		tokDb,

		param.cookPath,
		param.cookSec,
	}

	// バックエンドの準備完了。

	s := server.NewStopper()
	defer func() {
		// 処理の終了待ち。
		s.Lock()
		defer s.Unlock()
		for s.Stopped() {
			s.Wait()
		}
	}()

	mux := http.NewServeMux()
	routes := map[string]bool{}
	mux.HandleFunc(param.pathOk, pagePanicErrorWrapper(s, errTmpl, func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}))
	routes[param.pathOk] = true
	mux.HandleFunc(param.pathAuth, pagePanicErrorWrapper(s, errTmpl, sys.authPage))
	routes[param.pathAuth] = true
	mux.HandleFunc(param.pathCb, pagePanicErrorWrapper(s, errTmpl, sys.callbackPage))
	routes[param.pathCb] = true
	if param.uiDir != "" {
		// ファイル配信も自前でやる。
		pathUi := strings.TrimRight(param.pathUi, "/") + "/"
		mux.Handle(pathUi, http.StripPrefix(pathUi, http.FileServer(http.Dir(param.uiDir))))
		routes[param.pathUi] = true
	}

	if !routes["/"] {
		mux.HandleFunc("/", pagePanicErrorWrapper(s, errTmpl, func(w http.ResponseWriter, r *http.Request) error {
			return erro.Wrap(server.NewError(http.StatusNotFound, "invalid endpoint", nil))
		}))
	}

	return server.Serve(param, mux)
}

func pagePanicErrorWrapper(s *server.Stopper, errTmpl *template.Template, f server.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Stop()
		defer s.Unstop()

		// panic時にプロセス終了しないようにrecoverする
		defer func() {
			if rcv := recover(); rcv != nil {
				server.RespondPageError(w, r, erro.New(rcv), errTmpl, request.Parse(r, "").String()+": ")
				return
			}
		}()

		//////////////////////////////
		server.LogRequest(level.DEBUG, r, true)
		//////////////////////////////

		if err := f(w, r); err != nil {
			server.RespondPageError(w, r, erro.Wrap(err), errTmpl, request.Parse(r, "").String()+": ")
			return
		}
	}
}

func apiPanicErrorWrapper(s *server.Stopper, f server.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Stop()
		defer s.Unstop()

		// panic時にプロセス終了しないようにrecoverする
		defer func() {
			if rcv := recover(); rcv != nil {
				idperr.RespondApiError(w, r, erro.New(rcv), request.Parse(r, ""))
				return
			}
		}()

		//////////////////////////////
		server.LogRequest(level.DEBUG, r, true)
		//////////////////////////////

		if err := f(w, r); err != nil {
			idperr.RespondApiError(w, r, erro.Wrap(err), request.Parse(r, ""))
			return
		}
	}
}
