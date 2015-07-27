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
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
)

func TestCoopResponse(t *testing.T) {
	tok := "McnSl40-QRtAxBoBOmj9GJfALyNdJy"
	tokType := "Bearer"
	expIn := 2817 * time.Second
	scop := strsetutil.New("openid")
	idsTok := []byte("eyJhbGciOiJSUzI1NiJ9.eyJhdWQiOiJodHRwOi8vbG9jYWxob3N0OjcwMDAiLCJleHAiOjE0MzI4MjI1MzMsImlhdCI6MTQzMjgwMDkzMywiaWRzIjp7Im1haW4iOnsic3ViIjoiWi0wWnRGZWlyaU1kWkZIYlNDR3ZDbHplclBYWkU5SGJKNWlaREx2X3hVWSJ9fSwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDoxNjA0Iiwic3ViIjoiaHR0cDovL2xvY2FsaG9zdDo3MDAwIn0.JRCRQRJkE5UTwniZqYf6Y-DncD58Qs0GQe5SF7A_K66nymZR9mUrXWj0rAG6NcXEsH5IpoFfcYUVLylx_XxyrKNK8ynGgiGwOblNW41rb2FmLmtFdf3Y03z9ID5hWJsEJmyup071s5WUeURAa3xv-h-t-cSBqXIv_R-TPWGyWSKVSe6lKudfOh74rrG_5IHheFs3hKiyw9viVJRV9sOekcrV93ppdUweguxx_AgSjFhdEh6atmUy5ft3Oz3doqfSQHQ2xMR_V9cFryzXrfJDyual4KBFMQHJ3LU2kyuoebKpHjwT5Iv3Bn-QY7U_SIO4vvcBuOcvjNyWyymla6KzPA")
	body, err := json.Marshal(map[string]interface{}{
		"access_token": tok,
		"token_type":   tokType,
		"expires_in":   int(expIn / time.Second),
		"scope":        requtil.ValueSetForm(scop),
		"ids_token":    string(idsTok),
	})
	if err != nil {
		t.Fatal(err)
	}

	r, err := http.ReadResponse(bufio.NewReader(io.MultiReader(
		strings.NewReader("HTTP/1.1 200 OK\r\n"+"Content-Type: application/json\r\n"+"\r\n"),
		bytes.NewReader(body),
	)), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := parseCoopResponse(r)
	if err != nil {
		t.Fatal(err)
	} else if resp.token() != tok {
		t.Error(resp.token())
		t.Fatal(tok)
	} else if resp.tokenType() != tokType {
		t.Error(resp.tokenType())
		t.Fatal(tokType)
	} else if resp.expiresIn() != expIn {
		t.Error(resp.expiresIn())
		t.Fatal(expIn)
	} else if !reflect.DeepEqual(resp.scope(), scop) {
		t.Error(resp.scope())
		t.Fatal(scop)
	} else if !bytes.Equal(resp.idsToken(), idsTok) {
		t.Error(resp.idsToken())
		t.Fatal(idsTok)
	}
}
