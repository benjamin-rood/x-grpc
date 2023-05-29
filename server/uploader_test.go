package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"testing"
	"time"

	uploadpb "github.com/benjamin-rood/x-grpc/proto"
	"github.com/go-test/deep"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestUploaderService_UploadFile(t *testing.T) {
	// SETUP: Server with Uploader Service
	// register server with uploader service
	buf := NewBufferWriter()
	uploadSvc := NewCustomUploader(buf)
	conn := newTestGRPCServer(t, func(srv *grpc.Server) {
		uploadpb.RegisterUploaderServer(srv, uploadSvc)
	})

	// SETUP: Test Client
	// initialise grpc service client
	// use generated client for testing
	client := uploadpb.NewUploaderClient(conn)

	//- Test, stream upload of file
	// TODO: confirm response is valid and matches expected
	const fn = "testBlob"
	_, err := sendDataInChunksToServer(t, client, jsonBlob, fn, "application/json")
	if err != nil {
		t.Fatalf("client.UploadFile: %s", err)
	}
	// confirm uploaded data matches sent exactly
	uploadedData, found := buf.m[fn]
	if !found {
		t.Fatalf("cannot find `%s` entry", fn)
	}
	if string(uploadedData) != jsonBlob {
		t.Errorf("STORED DATA ≠ SENT DATA\n%s\n≠\n%s", uploadedData, jsonBlob)
	}
	// confirm modified data is as expected
	// get resulting modified json
	modifiedData, _ := buf.loadCurrent()
	// Compare the modified JSON with the expected output
	// Unmarshal resulting json blob into a map to make comparison easier
	modifiedDataMap := map[string]any{}
	if err := json.Unmarshal(modifiedData, &modifiedDataMap); err != nil {
		t.Fatal(err)
	}
	// Unmarshal expectedOutput into a map to make comparison easier
	expectedDataMap := map[string]any{}
	if err := json.Unmarshal([]byte(expectedOutput), &expectedDataMap); err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(modifiedDataMap, expectedDataMap); diff != nil {
		t.Errorf("%v", diff)
	}
	// TODO: make set of test cases
}

func sendDataInChunksToServer(t *testing.T, client uploadpb.UploaderClient, data string, fileName string, mimeType string) (*uploadpb.UploadResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(func() {
		cancel()
	})
	// Create a stream for uploading the file.
	stream, err := client.UploadFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %s", err)
	}
	chunkSize := 10
	reader := strings.NewReader(data)
	buf := make([]byte, chunkSize)

	// Read the data in chunks and send them to the server one chunk at a time
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("Error reading string: %w", err)
		}
		if n == 0 {
			break
		}
		chunk := buf[:n]
		if err := stream.Send(&uploadpb.UploadRequest{
			FileName: fileName,
			Chunk:    chunk,
			MimeType: mimeType,
		}); err != nil {
			return nil, fmt.Errorf("%s: failed to send chunk:\n<%s>", err, chunk)
		}

		// // Simulate connection issues by randomly sleeping between bursts.
		// sleepTime := rand.Intn(450) + 50 // Sleep for 50-500ms.
		// time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}
	// Sent all chunks
	// Close the stream and wait for the server to respond.
	return stream.CloseAndRecv()
}

// handy helper stolen from github.com/MarioCarrion/grpc-microservice-example
func newTestGRPCServer(t *testing.T, register func(srv *grpc.Server)) *grpc.ClientConn {
	lis := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() {
		lis.Close()
	})

	srv := grpc.NewServer()
	t.Cleanup(func() {
		srv.Stop()
	})

	register(srv)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("srv.Serve %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(func() {
		cancel()
	})

	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	t.Cleanup(func() {
		conn.Close()
	})
	if err != nil {
		t.Fatalf("grpc.DialContext %v", err)
	}

	return conn
}
