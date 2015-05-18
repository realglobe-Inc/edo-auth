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
	"bytes"
	"github.com/realglobe-Inc/edo-lib/base64url"
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
	idp, _ := base.Claim(tagIss).(string)
	if idp == "" {
		return nil, erro.New("no idp")
	}
	nonc, _ := base.Claim(tagNonce).(string)
	if idp == "" {
		return nil, erro.New("no nonce")
	}

	return &idToken{
		base: base,
		alg:  alg,
		idp:  idp,
		nonc: nonc,
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

func (this *idToken) codeHash() ([]byte, error) {
	if this.cHash == nil {
		raw, _ := this.base.Claim(tagC_hash).(string)
		if raw == "" {
			return nil, nil
		}
		cHash, err := base64url.DecodeString(raw)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		this.cHash = cHash
	}
	return this.cHash, nil
}

func (this *idToken) verifyCodeHash(cod string) (err error) {
	cHash, err := this.codeHash()
	if err != nil {
		return erro.Wrap(err)
	} else if cHash == nil {
		return nil
	}

	hGen, err := jwt.HashFunction(this.algorithm())
	if err != nil {
		return erro.Wrap(err)
	}
	h := hGen.New()
	h.Write([]byte(cod))
	hash := h.Sum(nil)
	if !bytes.Equal(cHash, hash[:len(hash)/2]) {
		return erro.New("verification failed")
	}
	return nil
}

func (this *idToken) tokenHash() ([]byte, error) {
	if this.atHash == nil {
		raw, _ := this.base.Claim(tagAt_hash).(string)
		if raw == "" {
			return nil, nil
		}
		atHash, err := base64url.DecodeString(raw)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		this.atHash = atHash
	}
	return this.atHash, nil
}

func (this *idToken) verifyTokenHash(tok string) (err error) {
	atHash, err := this.tokenHash()
	if err != nil {
		return erro.Wrap(err)
	} else if atHash == nil {
		return nil
	}

	hGen, err := jwt.HashFunction(this.algorithm())
	if err != nil {
		return erro.Wrap(err)
	}
	h := hGen.New()
	h.Write([]byte(tok))
	hash := h.Sum(nil)
	if !bytes.Equal(atHash, hash[:len(hash)/2]) {
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
