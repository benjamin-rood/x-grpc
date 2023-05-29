package main

import (
	"bytes"
	"log"
)

// quick bootstrap of bytes.Buffer to use for testing
type bufwc struct {
	buffer  *bytes.Buffer
	m       map[string][]byte // quasi in-memory database of "files"
	current string
}

func NewBufferWriter() *bufwc {
	buf := bufwc{}
	buf.buffer = bytes.NewBuffer([]byte{})
	buf.m = make(map[string][]byte)
	return &buf
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

func (b *bufwc) Load(key string) ([]byte, error) {
	// return copy of data value of the matching key in database, perform modifications
	value, ok := b.m[key]
	if !ok {
		// being lazy
		log.Fatalf("no such entry '%s' to load", key)
	}
	data := make([]byte, len(value))
	n := copy(data, value)
	if n == 0 {
		panic("no bytes copied")
	}
	return data, nil
}

func (b *bufwc) String() string {
	return b.buffer.String()
}
