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
	"net/http"
	"net/url"
	"testing"
)

func TestCallbackRequest(t *testing.T) {
	code := "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	state := "Bhv56FnzSs"
	idTok := []byte("eyJhbGciOiJSUzI1NiIsImtpZCI6IjFlOWdkazcifQ.ewogImlzcyI6ICJodHRwOi8vc2VydmVyLmV4YW1wbGUuY29tIiwKICJzdWIiOiAiMjQ4Mjg5NzYxMDAxIiwKICJhdWQiOiAiczZCaGRSa3F0MyIsCiAibm9uY2UiOiAibi0wUzZfV3pBMk1qIiwKICJleHAiOiAxMzExMjgxOTcwLAogImlhdCI6IDEzMTEyODA5NzAKfQ.ggW8hZ1EuVLuxNuuIJKX_V8a_OMXzR0EHR9R6jgdqrOOF4daGU96Sr_P6qJp6IcmD3HP99Obi1PRs-cwh3LO-p146waJ8IhehcwL7F09JdijmBqkvPeB2T9CJNqeGpe-gccMg4vfKjkM8FcGvnzZUN4_KSP0aAp1tOJ1zZwgjxqGByKHiOtX7TpdQyHE5lcMiKPXfEIQILVq0pc_E2DzL7emopWoaoZTF_m0_N0YzFC6g6EJbOEoRoSK5hoDalrcvRYLSrQAZZKflyuVCyixEoV9GfNQC3_osjzw2PAithfubEEBLuVVk4XUVrWOLrLl0nx7RkKU8NXNHq-rvKMzqg")
	q := url.Values{}
	q.Set("code", code)
	q.Set("state", state)
	q.Set("id_token", string(idTok))
	r, err := http.NewRequest("GET", "https://ta.example.org/callback?"+q.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}

	req, err := parseCallbackRequest(r, nil)
	if err != nil {
		t.Fatal(err)
	}

	if req.code() != code {
		t.Error(req.code())
		t.Fatal(code)
	} else if req.state() != state {
		t.Error(req.state())
		t.Fatal(state)
	} else if !bytes.Equal(req.idToken(), idTok) {
		t.Error(string(req.idToken()))
		t.Fatal(string(idTok))
	}
}
