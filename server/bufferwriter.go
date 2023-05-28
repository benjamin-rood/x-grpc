package main

import (
	"bytes"
	"fmt"
)

// quick bootstrap of bytes.Buffer to use for testing
type bufwc struct {
	buffer  *bytes.Buffer
	m       map[string][]byte // quasi in-memory database of "files"
	current string
}

// Check interface conformity
var _ OpenWriteCloserLoader = &bufwc{}

func (b *bufwc) Open(key string) error {
	b.current = key
	b.buffer = bytes.NewBuffer([]byte{})
	return nil
}

func (b *bufwc) Write(p []byte) (n int, err error) {
	return b.buffer.Write(p)
}

// will save the content of the buffer into the "current" map entry
// then clear the contents of the bytes.Buffer to be reused
func (b *bufwc) Close() error {
	// copy the contents of the buffer into a new byte slice
	data := make([]byte, b.buffer.Len())
	copy(data, b.buffer.Bytes())
	// save in the "database"
	b.m[b.current] = data
	// clear the contents of the buffer
	b.buffer.Reset()
	return nil
}

func (b *bufwc) Load(string) ([]byte, error) {
	// copy contents of current entry in database, perform modifications
	return nil, fmt.Errorf("not implemented")
}

func (b *bufwc) String() string {
	return b.buffer.String()
}
