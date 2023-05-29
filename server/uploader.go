package main

import (
	"io"
	"log"
	"os"
	"strings"

	uploadpb "github.com/benjamin-rood/x-grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/**
 * Embed `UnsafeUploaderServer` instead of `UnimplementedUploaderServer`
 * to ensure we get compilation errors unless `Uploader` service is defined.
 * Good practice, even if it's a single rpc server.
 */
type Uploader struct {
	uploadpb.UnsafeUploaderServer

	// embedded type that does all the other stuff
	io_thingee OpenWriteCloserLoader
	// it's an 80s-90s kiwi childhood reference... why yes I am spending too much time on this,
	// and I can't think of a sensible thing to call this and I'm going nuts, sorry
	// see: https://en.wikipedia.org/wiki/Thingee
	// and: https://www.youtube.com/watch?v=GC3LK1nx-DU
}

// Check interface conformity
var _ uploadpb.UploaderServer = &Uploader{}

// ughhhh naming is hard
type OpenWriteCloserLoader interface {
	// why define this interface?
	// 1. good practice, makes the implementation easily customisable and extendable
	// 2. makes Uploader/UploadFile implementation *testable*
	//		-- we can implement an in-memory version which writes the upload to a bytes.Buffer
	//		& confirm what UploadFile writes without needing to write to disk (avoid whenever possible)
	//		and then having to clean that up afterwards as part of a test
	Open(string) error
	io.Writer
	io.Closer
	Load(string) ([]byte, error)
}

func NewCustomUploader(writer OpenWriteCloserLoader) *Uploader {
	return &Uploader{io_thingee: writer}
}

const receivedFilesDir = "./received_files"

func DefaultUploader() *Uploader {
	// create folder where uploaded files will go
	if err := os.MkdirAll(receivedFilesDir, os.ModePerm); err != nil {
		panic(err)
	}
	return &Uploader{io_thingee: &diskWriter{writeDirPath: receivedFilesDir}}
}

func (u *Uploader) UploadFile(stream uploadpb.Uploader_UploadFileServer) error {
	close := func() {
		if err := u.io_thingee.Close(); err != nil {
			log.Fatalf("could not close file: %s", err)
		}
		log.Println("closed")
	}
	defer close()

	// grab the initial message segment to get the `file_name` & `meta_data` arguments
	req, err := stream.Recv()
	contentType := req.GetMimeType()
	log.Println("Content-Type:", contentType)
	fn := strings.TrimSpace(req.GetFileName())
	// reject if no `file_name` argument provided, make use of it
	if fn == "" {
		return status.Errorf(codes.InvalidArgument, "missing file_name arg")
	}
	if err := u.io_thingee.Open(fn); err != nil {
		return status.Errorf(codes.Internal, "failed to open file: %s", err)
	}

	// implement handling of stream upload from a client in the following way:
	// - NOTE: we have already pulled the initial stream segment!
	// - for each received stream segment:
	//   - unless EOF, read the bytes chunk from the UploadRequest message and write DIRECTLY to disk
	// 	 - get next segment
	// - once we have received all the data, try to process the data as json
	var size uint32
	for {
		if err == io.EOF {
			// finish writing received bytes
			close()
			if contentType == "application/json" {
				// load data if a json file per bonus requirements, save a modified copy
				if err := ProcessJSON(fn, u.io_thingee); err != nil {
					return status.Errorf(codes.Internal, "failed to perform modifications to uploaded JSON data: %s", err)
				}
			}
			return stream.SendAndClose(&uploadpb.UploadResponse{
				FileName: fn,
				Size:     size,
			})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive chunk: %s", err)
		}

		if _, err := u.io_thingee.Write(req.GetChunk()); err != nil {
			return status.Errorf(codes.Internal, "failed to write chunk to file: %s", err)
		}
		size += uint32(len(req.GetChunk()))
		// get the next stream segment
		req, err = stream.Recv()
	}
}
