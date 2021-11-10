// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package list

import (
	"testing"
)

func TestExample(t *testing.T) {
	// Create a new list and put some numbers in it.
	l := New()
	e4 := l.PushBack(4, []byte{'4'})
	e1 := l.PushFront(1, []byte{'1'})
	l.InsertBefore(3, []byte{'3'}, e4)
	l.InsertAfter(2, []byte{'2'}, e1)

	// Iterate through list and print its contents.
	for e := l.Front(); e != nil; e = e.Next() {
		t.Log(e.Value)
		t.Log(e.Score)
		t.Log(e.InsertTime)
	}

	// Output:
	// 1
	// 2
	// 3
	// 4
}
