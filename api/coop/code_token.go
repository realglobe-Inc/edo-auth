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
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
)

// コードトークン。
type codeToken struct {
	base *jwt.Jwt

	idp      string
	cod      string
	aud      map[string]bool
	frTa     string
	acntTag  string
	acntTags map[string]bool
	refHash  string
}

func parseCodeToken(raw []byte) (*codeToken, error) {
	base, err := jwt.Parse(raw)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	var buff struct {
		Idp      string            `json:"iss"`
		Cod      string            `json:"sub"`
		Aud      audience.Audience `json:"aud"`
		FrTa     string            `json:"from_client"`
		AcntTag  string            `json:"user_tag"`
		AcntTags strset.Set        `json:"user_tags"`
		RefHash  string            `json:"ref_hash"`
	}
	if err := json.Unmarshal(base.RawBody(), &buff); err != nil {
		return nil, erro.Wrap(err)
	} else if buff.Idp == "" {
		return nil, erro.New("no ID provider ID")
	} else if buff.Cod == "" {
		return nil, erro.New("no code")
	} else if len(buff.Aud) == 0 {
		return nil, erro.New("no audience")
	}

	return &codeToken{
		base:     base,
		idp:      buff.Idp,
		cod:      buff.Cod,
		aud:      buff.Aud,
		frTa:     buff.FrTa,
		acntTag:  buff.AcntTag,
		acntTags: buff.AcntTags,
		refHash:  buff.RefHash,
	}, nil
}

func (this *codeToken) code() string {
	return this.cod
}

func (this *codeToken) idProvider() string {
	return this.idp
}

func (this *codeToken) audience() map[string]bool {
	return this.aud
}

func (this *codeToken) fromTa() string {
	return this.frTa
}

func (this *codeToken) referralHash() string {
	return this.refHash
}

func (this *codeToken) accountTag() string {
	return this.acntTag
}

func (this *codeToken) accountTags() map[string]bool {
	return this.acntTags
}

func (this *codeToken) verify(keys []jwk.Key) error {
	return this.base.Verify(keys)
}
