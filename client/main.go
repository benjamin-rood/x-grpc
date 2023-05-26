// client.go

package main

import (
	"context"
	"io"
	"log"
	"os"

	uploadpb "github.com/benjamin-rood/x-grpc/proto"
	"google.golang.org/grpc"
)

const (
	chunkSize = 100 * 1024 // Upload chunks of 100KB
)

func main() {
	// Set up a connection to the server (using insecure because this is not real)
	conn, err := grpc.Dial(":59999", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client instance.
	client := uploadpb.NewUploaderClient(conn)

	// Open the file to be uploaded.
	file, err := os.Open("./large_test.json")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a stream for uploading the file.
	stream, err := client.UploadFile(context.Background())
	if err != nil {
		log.Fatalf("failed to open stream: %v", err)
	}

	for {
		// Read the file in chunks and send them to the server.
		buf := make([]byte, chunkSize)
		n, err := file.Read(buf)
		if err != nil {
			log.Println(err)
		}
		if err != nil && err != io.EOF {
			log.Fatalf("failed to read file: %v", err)
		}
		if n == 0 {
			break
		}
		chunk := buf[:n]
		if err := stream.Send(&uploadpb.UploadRequest{Chunk: chunk}); err != nil {
			log.Fatalf("%s: failed to send chunk:\n<%s>", err, chunk)
		}

		// // Simulate connection issues by randomly sleeping between bursts.
		// sleepTime := rand.Intn(450) + 50 // Sleep for 50-500ms.
		// time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}

	// Close the stream and wait for the server to respond.
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed: %s", err)
	}
	log.Printf("uploaded file: %v (%v bytes)", resp.FileName, resp.Size)
}
