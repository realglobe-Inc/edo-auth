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

package token

import (
	"reflect"
	"testing"
	"time"
)

const (
	test_id  = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	test_tag = "7ST-oqqEQC"
	test_idp = "EYClXo4mQKwSgPel"
)

var (
	test_scop = map[string]bool{"openid": true, "email": true}
)

func TestElement(t *testing.T) {
	exp := time.Now().Add(time.Second)
	elem := New(test_id, test_tag, exp, test_idp, test_scop)

	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Fatal(test_id)
	} else if elem.Tag() != test_tag {
		t.Error(elem.Tag())
		t.Fatal(test_tag)
	} else if !elem.Expires().Equal(exp) {
		t.Error(elem.Expires())
		t.Fatal(exp)
	} else if elem.IdProvider() != test_idp {
		t.Error(elem.IdProvider())
		t.Fatal(test_idp)
	} else if !reflect.DeepEqual(elem.Scope(), test_scop) {
		t.Error(elem.Scope())
		t.Fatal(test_scop)
	}
}
