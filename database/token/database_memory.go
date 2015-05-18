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
	"sync"
	"time"
)

// メモリ上のアクセストークン情報の格納庫。
type memoryDb struct {
	lock      sync.Mutex
	tagToElem map[string]*Element
	tagToExp  map[string]time.Time
}

func NewMemoryDb() Db {
	return &memoryDb{
		tagToElem: map[string]*Element{},
		tagToExp:  map[string]time.Time{},
	}
}

// 取得。
func (this *memoryDb) GetByTag(tag string) (*Element, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	elem := this.tagToElem[tag]
	if elem == nil {
		return nil, nil
	} else if time.Now().After(this.tagToExp[tag]) {
		delete(this.tagToElem, tag)
		delete(this.tagToExp, tag)
		return nil, nil
	}

	// Replace で使う更新日時が変わらないように防御的コピー。
	e := *elem
	return &e, nil
}

// 保存。
func (this *memoryDb) Save(elem *Element, exp time.Time) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.tagToElem[elem.Tag()] = elem
	this.tagToExp[elem.Tag()] = exp
	return nil
}
