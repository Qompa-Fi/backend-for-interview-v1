package main

import (
	"bytes"
	"sync"
)

type SafeBuffer struct {
	mu sync.Mutex
	b  *bytes.Buffer
}

func NewSafeBuffer() SafeBuffer {
	return SafeBuffer{b: new(bytes.Buffer)}
}

func (sb *SafeBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return sb.b.Write(p)
}

func (sb *SafeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return sb.b.String()
}

func (sb *SafeBuffer) WriteString(s string) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return sb.b.WriteString(s)
}
