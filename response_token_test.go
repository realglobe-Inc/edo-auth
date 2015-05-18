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
	"bufio"
	"bytes"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestTokenResponse(t *testing.T) {
	tok := "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	tokType := "Bearer"
	expIn := 3600
	idTok := []byte("eyJhbGciOiJSUzI1NiIsImtpZCI6IjFlOWdkazcifQ.ewogImlzcyI6ICJodHRwOi8vc2VydmVyLmV4YW1wbGUuY29tIiwKICJzdWIiOiAiMjQ4Mjg5NzYxMDAxIiwKICJhdWQiOiAiczZCaGRSa3F0MyIsCiAibm9uY2UiOiAibi0wUzZfV3pBMk1qIiwKICJleHAiOiAxMzExMjgxOTcwLAogImlhdCI6IDEzMTEyODA5NzAKfQ.ggW8hZ1EuVLuxNuuIJKX_V8a_OMXzR0EHR9R6jgdqrOOF4daGU96Sr_P6qJp6IcmD3HP99Obi1PRs-cwh3LO-p146waJ8IhehcwL7F09JdijmBqkvPeB2T9CJNqeGpe-gccMg4vfKjkM8FcGvnzZUN4_KSP0aAp1tOJ1zZwgjxqGByKHiOtX7TpdQyHE5lcMiKPXfEIQILVq0pc_E2DzL7emopWoaoZTF_m0_N0YzFC6g6EJbOEoRoSK5hoDalrcvRYLSrQAZZKflyuVCyixEoV9GfNQC3_osjzw2PAithfubEEBLuVVk4XUVrWOLrLl0nx7RkKU8NXNHq-rvKMzqg")
	raw := strings.Replace(`HTTP/1.1 200 OK
Content-Type: application/json

{"access_token":"`+tok+
		`","token_type":"`+tokType+
		`","expires_in":`+strconv.Itoa(expIn)+
		`,"scope":"openid email","id_token": "`+string(idTok)+`"}`,
		"\n", "\r\n", -1)
	r, err := http.ReadResponse(bufio.NewReader(strings.NewReader(raw)), nil)
	if err != nil {
		t.Fatal(err)
	}

	exp := time.Now().Add(time.Duration(expIn) * time.Second)
	resp, err := parseTokenResponse(r)
	if err != nil {
		t.Fatal(err)
	}

	if resp.token() != tok {
		t.Error(resp.token())
		t.Fatal(tok)
	} else if resp.expires().Before(exp) {
		t.Error(resp.expires())
		t.Fatal(exp)
	} else if scop := map[string]bool{"openid": true, "email": true}; !reflect.DeepEqual(resp.scope(), scop) {
		t.Error(resp.scope())
		t.Fatal(scop)
	} else if !bytes.Equal(resp.idToken(), idTok) {
		t.Error(string(resp.idToken()))
		t.Fatal(string(idTok))
	}
}
