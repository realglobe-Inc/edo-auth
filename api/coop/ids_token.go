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
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

type idsToken struct {
	base *jwt.Jwt

	idp        string
	frTa       string
	aud        map[string]bool
	exp        time.Time
	date_      time.Time
	tagToAttrs map[string]map[string]interface{}
}

func parseIdsToken(raw []byte) (*idsToken, error) {
	base, err := jwt.Parse(raw)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	var buff struct {
		Idp        string                            `json:"iss"`
		FrTa       string                            `json:"sub"`
		Aud        audience.Audience                 `json:"aud"`
		Exp        int64                             `json:"exp"`
		Date       int64                             `json:"iat"`
		TagToAttrs map[string]map[string]interface{} `json:"ids"`
	}
	if err := json.Unmarshal(base.RawBody(), &buff); err != nil {
		return nil, erro.Wrap(err)
	} else if buff.Idp == "" {
		return nil, erro.New("no ID provider ID")
	} else if buff.FrTa == "" {
		return nil, erro.New("no from-TA ID")
	} else if buff.Aud == nil {
		return nil, erro.New("no audience")
	} else if buff.Exp == 0 {
		return nil, erro.New("no expiration date")
	} else if buff.Date == 0 {
		return nil, erro.New("no published date")
	} else if len(buff.TagToAttrs) == 0 {
		return nil, erro.New("no IDs")
	}
	return &idsToken{
		base:       base,
		idp:        buff.Idp,
		frTa:       buff.FrTa,
		aud:        buff.Aud,
		exp:        time.Unix(buff.Exp, 0),
		date_:      time.Unix(buff.Date, 0),
		tagToAttrs: buff.TagToAttrs,
	}, nil
}

func (this *idsToken) idProvider() string {
	return this.idp
}

func (this *idsToken) fromTa() string {
	return this.frTa
}

func (this *idsToken) audience() map[string]bool {
	return this.aud
}

func (this *idsToken) expires() time.Time {
	return this.exp
}

func (this *idsToken) date() time.Time {
	return this.date_
}

func (this *idsToken) attributes() map[string]map[string]interface{} {
	return this.tagToAttrs
}

func (this *idsToken) verify(keys []jwk.Key) error {
	return this.base.Verify(keys)
}
