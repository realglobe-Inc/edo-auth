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
	"github.com/realglobe-Inc/edo-id-provider/request"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// パニックとエラーの処理をまとめる。
func panicErrorWrapper(s *server.Stopper, f server.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Stop()
		defer s.Unstop()

		// panic時にプロセス終了しないようにrecoverする
		defer func() {
			if rcv := recover(); rcv != nil {
				responseError(w, erro.New(rcv), request.Parse(r, ""))
				return
			}
		}()

		//////////////////////////////
		server.LogRequest(level.DEBUG, r, true)
		//////////////////////////////

		if err := f(w, r); err != nil {
			responseError(w, erro.Wrap(err), request.Parse(r, ""))
			return
		}
	}
}

// ユーザーエージェント向けにエラーを返す。
func responseError(w http.ResponseWriter, origErr error, sender *request.Request) {
	e := server.ErrorFrom(origErr)
	log.Err(sender, ": "+e.Message())
	log.Debug(sender, ": ", origErr)

	msg := `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Error</title></head><body><h1>`
	msg += strconv.Itoa(e.Status())
	msg += " "
	msg += http.StatusText(e.Status())
	msg += `</h1><p><font size="+1"><b>`
	msg += strings.Replace(html.EscapeString(e.Message()), "\n", "<br/>", -1)
	msg += `:</b></font></p><p>`
	msg += strings.Replace(html.EscapeString(e.Error()), "\n", "<br/>", -1)
	msg += `</p></body></html>`
	buff := []byte(msg)

	w.Header().Set(tagContent_type, server.ContentTypeHtml)
	w.Header().Set(tagContent_length, strconv.Itoa(len(buff)))
	w.WriteHeader(e.Status())
	if _, err := w.Write(buff); err != nil {
		log.Err(sender, ": ", erro.Wrap(err))
	}
}

// エラーページにリダイレクトさせる。
func redirectError(w http.ResponseWriter, r *http.Request, origErr error, sender *request.Request, uri *url.URL) error {
	e := server.ErrorFrom(origErr)
	log.Err(sender, ": "+e.Message())
	log.Debug(sender, ": ", origErr)

	queries := uri.Query()
	queries.Set(tagStatus, strconv.Itoa(e.Status()))
	queries.Set(tagMessage, e.Message())
	uri.RawQuery = queries.Encode()

	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)
	http.Redirect(w, r, uri.String(), http.StatusFound)
	return nil
}
