package main

import (
	"fmt"
	"io"
	"log"
	"os"

	uploadpb "github.com/benjamin-rood/x-grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type filepath string

/**
 * Embed `UnsafeUploaderServer` instead of `UnimplementedUploaderServer`
 * to ensure we get compilation errors unless `Uploader` service is defined.
 * Good practice, even if it's a single rpc server.
 */
type Uploader struct {
	uploadpb.UnsafeUploaderServer
	w OpenWriteCloser
}

// Check interface conformity
var _ uploadpb.UploaderServer = &Uploader{}

type OpenWriteCloser interface {
	Open(string) error
	io.Writer
	io.Closer
}

// just a glorified wrapper around a call to `os.Create(...)`
type diskWriter struct {
	f *os.File
}

// Check interface conformity
var _ OpenWriteCloser = &diskWriter{}

func (dw *diskWriter) Open(filename string) error {
	var err error
	log.Printf("creating file '%s'\n", filename)
	p := "./received_files/" + filename
	dw.f, err = os.Create(p)
	return err
}

func (dw *diskWriter) Write(p []byte) (int, error) {
	return dw.f.Write(p)
}

func (dw *diskWriter) Close() error {
	return dw.f.Close()
}

func NewCustomUploader(writer OpenWriteCloser) *Uploader {
	return &Uploader{w: writer}
}

func NewUploader() *Uploader {
	return &Uploader{w: &diskWriter{}}
}

func (u *Uploader) UploadFile(stream uploadpb.Uploader_UploadFileServer) error {
	// implement handling of stream upload from a client in the following way:
	// - use the os package to open a file pointer so we can write bytes to disk at
	//   file path specified in the `p` constant
	// - for each received stream part:
	//   - unless EOF, read the bytes chunk from the UploadRequest message and write DIRECTLY to disk
	// - when stream is finished, safely close and return nil
	const p = "./received_file"
	if err := u.w.Open(p); err != nil {
		return status.Errorf(codes.Internal, "failed to open file: %v", err)
	}
	defer func() {
		log.Println("closing file")
		u.w.Close()
		log.Println("closed")
	}()

	var size uint32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&uploadpb.UploadResponse{
				FileName: p,
				Size:     size,
			})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive chunk: %v", err)
		}

		if _, err := u.w.Write(req.GetChunk()); err != nil {
			return status.Errorf(codes.Internal, "failed to write chunk to file: %v", err)
		}
		size += uint32(len(req.GetChunk()))
	}
}

/*
*

	since the saved file has no guaranteed file size limit* (hypothetically it could
	be greater than the availability of the available memory), the only way to prevent
	crashing by running out of memory is to either assert an upper file size limit
	beneath the currently availble memory on the system, or, we must do an on-disk
	byte traversal
*/
func modifyJSON() error {
	return fmt.Errorf("not implemented")
}
