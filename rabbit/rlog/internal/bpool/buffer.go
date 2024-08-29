// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package buffer provides a pool-allocated byte buffer.
package bpool

import (
	"sync"
)

// adapted from go/src/fmt/print.go

type B []byte

var pool = sync.Pool{
	New: func() any {
		b := make(B, 1024)
		return &b
	},
}

func New() *B {
	return pool.Get().(*B)
}

func (b *B) Free() {
	const maxBufferSize = 16 << 10
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		pool.Put(b)
	}
}

func (b *B) SetLen(n int) {
	*b = (*b)[:n]
}

func (b *B) Reset() {
	b.SetLen(0)
}
