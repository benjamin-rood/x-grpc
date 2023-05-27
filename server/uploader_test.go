package main

import (
	"bytes"
	"testing"
)

func TestUploaderService(t *testing.T) {
	// define listener
	// initialise grpc server
	// register server with uploader service
	// initialise grpc service client
	// use generated client for testing

}

// quick bootstrap of bytes.Buffer to use for testing
type bufwc struct {
	buffer *bytes.Buffer
}

func (b *bufwc) Open(string) error {
	// ignore filename
	b.buffer = bytes.NewBuffer([]byte{})
	return nil
}

func (b *bufwc) Write(p []byte) (n int, err error) {
	return b.buffer.Write(p)
}

func (b *bufwc) Close() error {
	return nil // Since it's just a buffer, closing has no effect
}

func (b *bufwc) String() string {
	return b.buffer.String()
}

// Check interface conformity
var _ OpenWriteCloser = &bufwc{}
