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
	"testing"
)

func TestCodeToken(t *testing.T) {
	raw := []byte("eyJhbGciOiJSUzI1NiJ9.eyJhdWQiOiJodHRwOi8vbG9jYWxob3N0OjcwMDAiLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjE2MDQiLCJzdWIiOiJPSzR3MnE1MjdaLWs5TnNKRU1TTXBtZm5ISFBkT3EiLCJ1c2VyX3RhZyI6Im1haW4ifQ.ylDMpROhQcin2tAupuSgqGpIaL2PZ-uF5SkOE1Y36UuwED_YxwboF0e29rstIgnjlyPQ3JvzOAY5A6uATFL-JKMAQV7zPoNt2O4delieRBA-OOJ3blZKFIEKdj7WRE6cHTb94_kb47LuVk0qEb_RTqLNyd-J-mmf6EaedevzBgN3mF2wRZPnzh4gQiwMPMyeJSpm7d0OhQq9j0ZuMbxxWZV1KhN4b4Do631wXh9rdoEsLL18ICnBuwWsErTleqMQ-cOuBusqGSJiclGjQaShk1CGY5gJgMu43K96OHVWXcV0f1zXGyml2bqEpknYmAJ7w2mrjydRKC7bidMwMIZ7ZQ")

	codTok, err := parseCodeToken(raw)
	if err != nil {
		t.Fatal(err)
	} else if cod := "OK4w2q527Z-k9NsJEMSMpmfnHHPdOq"; codTok.code() != cod {
		t.Error(codTok.code())
		t.Fatal(cod)
	} else if idp := "http://localhost:1604"; codTok.idProvider() != idp {
		t.Error(codTok.idProvider())
		t.Fatal(idp)
	} else if aud := "http://localhost:7000"; !codTok.audience()[aud] {
		t.Error(codTok.audience())
		t.Fatal(aud)
	} else if acntTag := "main"; codTok.accountTag() != acntTag {
		t.Error(codTok.accountTag)
		t.Fatal(acntTag)
	}
}
