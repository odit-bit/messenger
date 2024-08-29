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
		b := make(B, 0, 1024)
		return (*B)(&b)
	},
}

func New() *B {
	return pool.Get().(*B)
}

func (b *B) WriteByte(byt byte) error {
	*b = append(*b, byt)
	return nil
}

func (b *B) Write(bytes []byte) (int, error) {
	*b = append(*b, bytes...)
	return len(bytes), nil
}

func (b *B) WriteString(s string) (int, error) {
	*b = append(*b, s...)
	return len(s), nil
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

func (b *B) Len() int {
	return len(*b)
}

func (b *B) Reset() {
	b.SetLen(0)
}
