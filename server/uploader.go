package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	uploadpb "github.com/benjamin-rood/x-grpc/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
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

	// Extract the client's IP address from the context
	clientIP, err := getClientIPFromContext(stream.Context())
	if err != nil {
		return err
	}
	// Create a random string using the client's IP address & datetime stamp
	tmpfn := generateTempFilename(clientIP)
	// grab the initial message segment to get the `file_name` & `meta_data` arguments
	req, err := stream.Recv()
	contentType := req.GetMimeType()
	log.Println("Content-Type:", contentType)
	fn := strings.TrimSpace(req.GetFileName())
	// if there was a non-empty `file_name` argument provided, make use of it
	if fn != "" {
		fn = tmpfn + "_" + fn // end with the file_name as it could have a file extension suffix
	} else {
		fn = tmpfn
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

func generateTempFilename(clientIP net.IP) string {
	// Generate a short UUIDv1 string, in lieu of some request ID
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

	clientIP := pr.Addr.(*net.TCPAddr).IP

	return clientIP, nil
}

func generateShortUUIDv1() string {
	uuidV1 := uuid.New()
	shortUUIDv1 := uuidV1.String()[0:8]
	return shortUUIDv1
}
