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

package asession

import (
	"encoding/json"
	"time"

	"github.com/realglobe-Inc/go-lib/erro"
)

// セッション。
type Element struct {
	id string
	// 無効かどうか。
	inv bool
	// 有効期限。
	exp time.Time
	// 認証後にリダイレクトさせるパス。
	path string
	// ID プロバイダの ID。
	idp string
	// TA としての ID。
	ta string
	// リダイレクト URI。
	rediUri string
	// state
	stat string
	// nonce
	nonc string

	// 更新日時。
	date time.Time
}

func New(id string, exp time.Time, path, idp, ta, rediUri, stat, nonc string) *Element {
	return &Element{
		id:      id,
		exp:     exp,
		path:    path,
		idp:     idp,
		ta:      ta,
		rediUri: rediUri,
		stat:    stat,
		nonc:    nonc,
		date:    time.Now(),
	}
}

// ID を返す。
func (this *Element) Id() string {
	return this.id
}

// 無効にされているかどうか。
func (this *Element) Invalid() bool {
	return this.inv
}

// 無効にする。
func (this *Element) Invalidate() {
	this.inv = true
	this.date = time.Now()
}

// 有効期限を返す。
func (this *Element) Expires() time.Time {
	return this.exp
}

// パスを返す。
func (this *Element) Path() string {
	return this.path
}

func (this *Element) IdProvider() string {
	return this.idp
}

func (this *Element) Ta() string {
	return this.ta
}

func (this *Element) RedirectUri() string {
	return this.rediUri
}

func (this *Element) State() string {
	return this.stat
}

func (this *Element) Nonce() string {
	return this.nonc
}

// 更新日時を返す。
func (this *Element) Date() time.Time {
	return this.date
}

//  {
//      "id": <ID>,
//      "invalid": <無効か>,
//      "expires": <有効期限>,
//      "path": <リクエストパス>,
//      "id_provider": <ID プロバイダの ID>,
//      "client_id": <TA としての ID>,
//      "redirect_uri": <リダイレクトエンドポイント>,
//      "state": <state 値>,
//      "nonce": <nonce 値>
//  }
func (this *Element) MarshalJSON() (data []byte, err error) {
	m := map[string]interface{}{
		"id":           this.id,
		"invalid":      this.inv,
		"expires":      this.exp,
		"path":         this.path,
		"client_id":    this.ta,
		"redirect_uri": this.rediUri,
		"state":        this.stat,
		"nonce":        this.nonc,
		"date":         this.date,
	}
	if this.idp != "" {
		m["id_provider"] = this.idp
	}
	return json.Marshal(m)
}

func (this *Element) UnmarshalJSON(data []byte) error {
	var buff struct {
		Id      string    `json:"id"`
		Inv     bool      `json:"invalid"`
		Exp     time.Time `json:"expires"`
		Path    string    `json:"path"`
		Idp     string    `json:"id_provider"`
		Ta      string    `json:"client_id"`
		RediUri string    `json:"redirect_uri"`
		Stat    string    `json:"state"`
		Nonc    string    `json:"nonce"`
		Date    time.Time `json:"date"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.id = buff.Id
	this.inv = buff.Inv
	this.exp = buff.Exp
	this.path = buff.Path
	this.idp = buff.Idp
	this.ta = buff.Ta
	this.rediUri = buff.RediUri
	this.stat = buff.Stat
	this.nonc = buff.Nonc
	this.date = buff.Date
	return nil
}
