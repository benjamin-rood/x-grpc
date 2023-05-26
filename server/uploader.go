package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	uploadpb "github.com/benjamin-rood/x-grpc/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
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

// why define this interface?
// 1. good practice, makes the implementation easily customisable and extendable
// 2. makes Uploader/UploadFile implementation *testable* -- we can write the upload
// to an in-memory buffer and confirm what UploadFile writes without needing to write
// to disk and then having to clean that up
type OpenWriteCloser interface {
	Open(string) error
	io.Writer
	io.Closer
}

// default use - just a glorified wrapper around a call to `os.Create(...)`
type diskWriter struct {
	f *os.File
}

// Check interface conformity
var _ OpenWriteCloser = &diskWriter{}

func (dw *diskWriter) Open(filename string) error {
	// use the os package to open a file pointer so we can write bytes to disk
	// to a file with the given filename
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
	// Extract the client's IP address from the context
	clientIP, err := getClientIPFromContext(stream.Context())
	if err != nil {
		return err
	}
	// Create the random filename using the client's IP address
	tmpfn := generateTempFilename(clientIP)
	// FIXME: change `UploadRequest` message definition to include a destination filename

	// implement handling of stream upload from a client in the following way:
	// - for each received stream part:
	//   - unless EOF, read the bytes chunk from the UploadRequest message and write DIRECTLY to disk
	// - when stream is finished, safely close and return nil
	if err := u.w.Open(tmpfn); err != nil {
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
				FileName: tmpfn,
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

func generateTempFilename(clientIP net.IP) string {
	// Generate a short UUIDv1 string
	uuidV1 := generateShortUUIDv1()

	// Get the current datetime in UTC
	now := time.Now().UTC()

	// Format the datetime stamp
	datetimeStamp := now.Format("20060102-150405")

	// Combine the client's IP address, short UUIDv1, and datetime stamp to create the filename
	filenameParts := []string{
		clientIP.String(),
		uuidV1,
		datetimeStamp,
	}

	filename := fmt.Sprintf("%s", strings.Join(filenameParts, "_"))

	return filename
}

func getClientIPFromContext(ctx context.Context) (net.IP, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("failed to extract peer information from context")
	}

	// pr.Addr is a net.Addr containing the client's address information
	// You can extract the IP address from the Addr if it's a net.TCPAddr or net.UDPAddr
	// Example: clientIP := pr.Addr.(*net.TCPAddr).IP
	clientIP := pr.Addr.(*net.IPNet).IP

	return clientIP, nil
}

func generateShortUUIDv1() string {
	uuidV1 := uuid.New()
	shortUUIDv1 := uuidV1.String()[0:8]
	return shortUUIDv1
}
