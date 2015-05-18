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

package token

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

// アクセストークン情報。
type Element struct {
	id string
	// アクセストークンタグ。
	tag string
	// 有効期限。
	exp time.Time
	// 発行 IdP の ID。
	idp string
	// 許可スコープ。
	scop map[string]bool
}

func New(id, tag string, exp time.Time, idp string, scop map[string]bool) *Element {
	return &Element{
		id:   id,
		tag:  tag,
		exp:  exp,
		idp:  idp,
		scop: scop,
	}
}

// ID を返す。
func (this *Element) Id() string {
	return this.id
}

// アクセストークンタグを返す。
func (this *Element) Tag() string {
	return this.tag
}

// 有効期限を返す。
func (this *Element) Expires() time.Time {
	return this.exp
}

// IdP の ID を返す。
func (this *Element) IdProvider() string {
	return this.idp
}

// 許可スコープを返す。
func (this *Element) Scope() map[string]bool {
	return this.scop
}

//  {
//      "id": <ID>,
//      "tag": <アクセストークンタグ>,
//      "expires": <有効期限>,
//      "issuer": <IdP の ID>,
//      "scope": [
//          <許可スコープ>,
//          ...
//      ]
//  }
func (this *Element) MarshalJSON() (data []byte, err error) {
	return json.Marshal(map[string]interface{}{
		"id":      this.id,
		"tag":     this.tag,
		"expires": this.exp,
		"issuer":  this.idp,
		"scope":   strset.Set(this.scop),
	})
}

func (this *Element) UnmarshalJSON(data []byte) error {
	var buff struct {
		Id   string     `json:"id"`
		Tag  string     `json:"tag"`
		Exp  time.Time  `json:"expires"`
		Idp  string     `json:"issuer"`
		Scop strset.Set `json:"scope"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.id = buff.Id
	this.tag = buff.Tag
	this.exp = buff.Exp
	this.idp = buff.Idp
	this.scop = buff.Scop
	return nil
}
