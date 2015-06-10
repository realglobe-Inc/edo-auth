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

package auth

import (
	"bytes"
	"encoding/json"
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/hash"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/go-lib/erro"
)

type idToken struct {
	base *jwt.Jwt

	// 必須。
	alg  string
	idp  string
	nonc string

	// オプション。
	cHash  []byte
	atHash []byte
	attrs  map[string]interface{}
}

func parseIdToken(raw []byte) (*idToken, error) {
	base, err := jwt.Parse(raw)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	alg, _ := base.Header(tagAlg).(string)
	if alg == "" {
		return nil, erro.New("no alg")
	}
	var buff struct {
		Idp    string `json:"iss"`
		Nonc   string `json:"nonce"`
		CHash  string `json:"c_hash"`
		AtHash string `json:"at_hash"`
	}
	if err := json.Unmarshal(base.RawBody(), &buff); err != nil {
		return nil, erro.Wrap(err)
	} else if buff.Idp == "" {
		return nil, erro.New("no ID provider ID")
	} else if buff.Nonc == "" {
		return nil, erro.New("no nonce")
	}
	var cHash, atHash []byte
	if buff.CHash != "" {
		cHash, err = base64url.DecodeString(buff.CHash)
		if err != nil {
			return nil, erro.Wrap(err)
		}
	}
	if buff.AtHash != "" {
		atHash, err = base64url.DecodeString(buff.AtHash)
		if err != nil {
			return nil, erro.Wrap(err)
		}
	}

	return &idToken{
		base:   base,
		alg:    alg,
		idp:    buff.Idp,
		nonc:   buff.Nonc,
		cHash:  cHash,
		atHash: atHash,
	}, nil
}

func (this *idToken) algorithm() string {
	return this.alg
}

func (this *idToken) idProvider() string {
	return this.idp
}

func (this *idToken) nonce() string {
	return this.nonc
}

func (this *idToken) verify(keys []jwk.Key) error {
	return this.base.Verify(keys)
}

func (this *idToken) codeHash() []byte {
	return this.cHash
}

func (this *idToken) verifyCodeHash(cod string) (err error) {
	hGen, err := jwt.HashFunction(this.alg)
	if err != nil {
		return erro.Wrap(err)
	}
	h := hash.Hashing(hGen.New(), []byte(cod))
	if !bytes.Equal(this.cHash, h[:len(h)/2]) {
		return erro.New("verification failed")
	}
	return nil
}

func (this *idToken) tokenHash() []byte {
	return this.atHash
}

func (this *idToken) verifyTokenHash(tok string) (err error) {
	hGen, err := jwt.HashFunction(this.alg)
	if err != nil {
		return erro.Wrap(err)
	}
	h := hash.Hashing(hGen.New(), []byte(tok))
	if !bytes.Equal(this.atHash, h[:len(h)/2]) {
		return erro.New("verification failed")
	}
	return nil
}

func (this *idToken) attributes() map[string]interface{} {
	if this.attrs == nil {
		attrs := map[string]interface{}{}
		for _, name := range this.base.ClaimNames() {
			switch name {
			case tagAmr,
				tagAt_hash,
				tagAud,
				tagAzp,
				tagC_hash,
				tagExp,
				tagIat,
				tagJti,
				tagNbf,
				tagNonce,
				tagTyp:
				// アカウント情報ではない。
			default:
				attrs[name] = this.base.Claim(name)
			}
		}
		this.attrs = attrs
	}
	return this.attrs
}
