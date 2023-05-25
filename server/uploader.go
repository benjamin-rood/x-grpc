package main

import (
	"io"
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
	wc io.WriteCloser
}

// Check interface conformity
var _ uploadpb.UploaderServer = &Uploader{}

func (u *Uploader) UploadFile(stream uploadpb.Uploader_UploadFileServer) error {
	// implement handling of stream upload from a client in the following way:
	// - use the os package to open a file pointer so we can write bytes to disk at
	//   file path specified in the `p` constant
	// - for each received stream part:
	//   - unless EOF, read the bytes chunk from the UploadRequest message and write DIRECTLY to disk
	// - when stream is finished, safely close and return nil
	const p = "./received_file"
	var err error
	if u.wc == nil {
		// fallback to using os.OpenFile
		u.wc, err = os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to open file: %v", err)
		}
	}
	defer u.wc.Close()

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

		if _, err := u.wc.Write(req.GetChunk()); err != nil {
			return status.Errorf(codes.Internal, "failed to write chunk to file: %v", err)
		}
		size += uint32(len(req.GetChunk()))
	}
}
