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
	"github.com/realglobe-Inc/edo-auth/database/token"
	"github.com/realglobe-Inc/edo-auth/database/usession"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"time"
)

func newTestPage(keys []jwk.Key, idps []idpdb.Element) *Page {
	return New(
		server.NewStopper(),
		"https://ta.example.org",
		"https://ta.example.org/callback",
		test_taSigAlg,
		nil,
		"Auth-User-Backend",
		30,
		time.Minute,
		time.Hour,
		"Auth-User",
		30,
		time.Minute,
		10,
		10,
		10,
		time.Hour,
		20,
		time.Minute,
		keydb.NewMemoryDb(keys),
		idpdb.NewMemoryDb(idps),
		usession.NewMemoryDb(),
		token.NewMemoryDb(),
		rand.New(time.Second),
		"/",
		false,
		true,
	)
}
