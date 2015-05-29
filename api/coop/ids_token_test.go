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
	"reflect"
	"testing"
)

func TestIdsToken(t *testing.T) {
	raw := []byte("eyJhbGciOiJSUzI1NiJ9.eyJhdWQiOiJodHRwOi8vbG9jYWxob3N0OjcwMDAiLCJleHAiOjE0MzI4MjI1MzMsImlhdCI6MTQzMjgwMDkzMywiaWRzIjp7Im1haW4iOnsic3ViIjoiWi0wWnRGZWlyaU1kWkZIYlNDR3ZDbHplclBYWkU5SGJKNWlaREx2X3hVWSJ9fSwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDoxNjA0Iiwic3ViIjoiaHR0cDovL2xvY2FsaG9zdDo3MDAwIn0.JRCRQRJkE5UTwniZqYf6Y-DncD58Qs0GQe5SF7A_K66nymZR9mUrXWj0rAG6NcXEsH5IpoFfcYUVLylx_XxyrKNK8ynGgiGwOblNW41rb2FmLmtFdf3Y03z9ID5hWJsEJmyup071s5WUeURAa3xv-h-t-cSBqXIv_R-TPWGyWSKVSe6lKudfOh74rrG_5IHheFs3hKiyw9viVJRV9sOekcrV93ppdUweguxx_AgSjFhdEh6atmUy5ft3Oz3doqfSQHQ2xMR_V9cFryzXrfJDyual4KBFMQHJ3LU2kyuoebKpHjwT5Iv3Bn-QY7U_SIO4vvcBuOcvjNyWyymla6KzPA")

	idsTok, err := parseIdsToken(raw)
	if err != nil {
		t.Fatal(err)
	} else if idp := "http://localhost:1604"; idsTok.idProvider() != idp {
		t.Error(idsTok.idProvider())
		t.Fatal(idp)
	} else if frTa := "http://localhost:7000"; idsTok.fromTa() != frTa {
		t.Error(idsTok.fromTa())
		t.Fatal(frTa)
	} else if aud := "http://localhost:7000"; !idsTok.audience()[aud] {
		t.Error(idsTok.audience())
		t.Fatal(aud)
	} else if idsTok.expires().IsZero() {
		t.Fatal("no expiration date")
	} else if idsTok.date().IsZero() {
		t.Fatal("no date")
	} else if attrs := map[string]map[string]interface{}{"main": {"sub": "Z-0ZtFeiriMdZFHbSCGvClzerPXZE9HbJ5iZDLv_xUY"}}; !reflect.DeepEqual(idsTok.attributes(), attrs) {
		t.Error(idsTok.attributes())
		t.Fatal(attrs)
	}
}
