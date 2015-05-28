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
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
)

type codeToken struct {
	idp      string
	cod      string
	aud      map[string]bool
	acntTag  string
	acntTags map[string]bool
	refHash  string
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

func (this *codeToken) referralHash() string {
	return this.refHash
}

func (this *codeToken) accountTag() string {
	return this.acntTag
}

func (this *codeToken) accountTags() map[string]bool {
	return this.acntTags
}

func (this *codeToken) UnmarshalJSON(data []byte) error {
	var buff struct {
		Idp      string            `json:"iss"`
		Cod      string            `json:"sub"`
		Aud      audience.Audience `json:"aud"`
		AcntTag  string            `json:"user_tag"`
		AcntTags strset.Set        `json:"user_tags"`
		RefHash  string            `json:"ref_hash"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.idp = buff.Idp
	this.cod = buff.Cod
	this.aud = buff.Aud
	this.acntTag = buff.AcntTag
	this.acntTags = buff.AcntTags
	this.refHash = buff.RefHash
	return nil
}
