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
	"net/http"
	"reflect"
	"testing"
)

func TestRequest(t *testing.T) {
	raw := []byte("eyJhbGciOiJSUzI1NiJ9.eyJhdWQiOiJodHRwOi8vbG9jYWxob3N0OjcwMDAiLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjE2MDQiLCJzdWIiOiJ4LVdsaFM4V2pnUU5CQW5EMk5ZWUpTZzFrTHRSN3ciLCJ1c2VyX3RhZyI6Im1haW4ifQ.Mq8jgri0xN3Bx2nCLfXcp2MG3-YOpus3zaDGcw9e-7uC_rqOcqwpZkKSWHPk7fY9BAja_DHRNutMxavMUgvstR9TO2BSYyNTS6PUwchv9OIydZh-eAT2NF63LGXG5whqfrP6L-p4RCK_MKYn-cxPxJBCkdHkxevdBRyagOF-JCAwu74A7N1Z9LGc7xkTPJONjt_Hqr4kQs5-Lsk3_jkk8omLO448dgE5iRP6QFnVRBO5ES975eQmCql8gE68nNI5KzAcyEuubVt6YYoORdPYJxPXFN5dglXvTSs29GpDc8wJhAaS3eY184nsdt9_MSGS4emeKXAs3L6HE1qDGhnW_g")

	r, err := http.NewRequest("GET", "https://ta.example.org/api/invite", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("X-Edo-Code-Tokens", string(raw))

	req, err := parseRequest(r)
	if err != nil {
		t.Fatal(err)
	} else if codToks := [][]byte{raw}; !reflect.DeepEqual(req.codeTokens(), codToks) {
		t.Error(req.codeTokens())
		t.Fatal(codToks)
	}
}
