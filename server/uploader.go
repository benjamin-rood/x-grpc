package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
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
	ProcessJSON() error
}

// default use - just a glorified wrapper around a call to `os.OpenFile(...)`
type diskWriter struct {
	f               *os.File
	writeDirPath    string
	currentFilename string
	contentType     string
}

// Check interface conformity
var _ OpenWriteCloser = &diskWriter{}

// uses the os package to open a file pointer so we can write bytes to disk
// to a file with the given filename
func (dw *diskWriter) Open(filename string) error {
	dw.currentFilename = filename // must set this prior to doing anything else, used by filePath()
	log.Printf("opening file '%s'\n", dw.filePath())
	var err error
	dw.f, err = os.OpenFile(dw.filePath(), os.O_RDWR|os.O_CREATE, 0644)
	return err
}

func (dw *diskWriter) Write(p []byte) (int, error) {
	return dw.f.Write(p)
}

func (dw *diskWriter) Close() error {
	log.Println("closing file")
	return dw.f.Close()
}

func (dw *diskWriter) filePath() string {
	if dw.currentFilename == "" {
		// if this ever happens, then there is a fault in the code/order
		// of execution and it's better to die and make sure it never happens
		log.Fatal("trying to get file r/w path when none set")
	}
	return filepath.Join(dw.writeDirPath, dw.currentFilename)
}

func (dw *diskWriter) loadFile() ([]byte, error) {
	return os.ReadFile(dw.filePath())
}

/*
*

	note: since the saved file has no guaranteed file size limit* (hypothetically it could
	be greater than the availability of the available memory), the only way to prevent
	crashing by running out of memory is to either assert an upper file size limit
	beneath the currently availble memory on the system, or, we must do an on-disk
	byte traversal of the JSON tree. This is not part of the assignment, so we will just
	error out if the JSON file to be loaded exceeds available memory.
*/
func (dw *diskWriter) ProcessJSON() error {
	// nasty, under-the-covers bullshit approach done to save time
	// if a JSON file, process it per the instructions!
	if dw.contentType == "application/json" {
		// open file again, permitted to load it all in to memory
		fileContent, err := dw.loadFile()
		if err != nil {
			return err
		}
		// make changes described in bonus requirements
		modifiedData, err := processJSON(fileContent)
		if err != nil {
			return err
		}
		// overwrite file contents with modified JSON data
		if err := dw.Open(dw.currentFilename); err != nil {
			return err
		}
		if _, err := dw.Write(modifiedData); err != nil {
			return fmt.Errorf("failed to write modified JSON data to file: %w", err)
		}
	}
	return nil
}

func NewCustomUploader(writer OpenWriteCloser) *Uploader {
	return &Uploader{w: writer}
}

const receivedFilesDir = "./received_files"

func DefaultUploader() *Uploader {
	return &Uploader{w: &diskWriter{writeDirPath: receivedFilesDir}}
}

func (u *Uploader) UploadFile(stream uploadpb.Uploader_UploadFileServer) error {
	close := func() {
		if err := u.w.Close(); err != nil {
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
	if err := u.w.Open(fn); err != nil {
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
			// finish writing
			close()
			// FIXME: Either process the json before returning the response, or send the filepath and contentType on a go channel for another process to take care of
			if err := u.w.ProcessJSON(); err != nil {
				return status.Errorf(codes.Internal, "failed to process uploaded JSON data: %s", err)
			}
			return stream.SendAndClose(&uploadpb.UploadResponse{
				FileName: fn,
				Size:     size,
			})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive chunk: %s", err)
		}

		if _, err := u.w.Write(req.GetChunk()); err != nil {
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
