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

package asession

import (
	"testing"
	"time"
)

const (
	test_id      = "EBBR9STJ-hUL9PzniRasrc-qeFxJ9m"
	test_path    = "/ui/index.html"
	test_idp     = "https://idp.example.org"
	test_ta      = "https://ta.example.org"
	test_rediUri = "https://ta.example.org/callback"
	test_stat    = "YJgUit_Wx5"
	test_nonc    = "Wjj1_YUOlR"
)

func TestElement(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	elem := New(test_id, exp, test_path, test_idp, test_ta, test_rediUri, test_stat, test_nonc)

	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Fatal(test_id)
	} else if elem.Invalid() {
		t.Fatal("invalid")
	} else if !elem.Expires().Equal(exp) {
		t.Error(elem.Expires())
		t.Fatal(exp)
	} else if elem.Path() != test_path {
		t.Error(elem.Path())
		t.Fatal(test_path)
	} else if elem.IdProvider() != test_idp {
		t.Error(elem.IdProvider())
		t.Fatal(test_idp)
	} else if elem.Ta() != test_ta {
		t.Error(elem.Ta())
		t.Fatal(test_ta)
	} else if elem.RedirectUri() != test_rediUri {
		t.Error(elem.RedirectUri())
		t.Fatal(test_rediUri)
	} else if elem.State() != test_stat {
		t.Error(elem.State())
		t.Fatal(test_stat)
	} else if elem.Nonce() != test_nonc {
		t.Error(elem.Nonce())
		t.Fatal(test_nonc)
	}

	date := elem.Date()
	elem.Invalidate()
	if !elem.Invalid() {
		t.Fatal("valid after invalidate")
	} else if elem.Date().Before(date) {
		t.Error(elem.Date())
		t.Fatal(date)
	}
}
