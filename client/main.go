// client.go

package main

import (
	"context"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	uploadpb "github.com/benjamin-rood/x-grpc/proto"
	"google.golang.org/grpc"
)

const (
	chunkSize = 128 // Upload chunks of 128 bytes.
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(":50080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client instance.
	client := uploadpb.NewUploaderClient(conn)

	// Open the file to be uploaded.
	file, err := os.Open("./small_test.json")
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
		log.Println(n)
		log.Println(string(buf[:n]))
		if err != nil && err != io.EOF {
			log.Fatalf("failed to read file: %v", err)
		}
		if n == 0 {
			break
		}
		if err := stream.Send(&uploadpb.UploadRequest{Chunk: buf[:n]}); err != nil {
			log.Fatalf("failed to send chunk: %v", err)
		}

		// Simulate connection issues by randomly sleeping between bursts.
		sleepTime := rand.Intn(450) + 50 // Sleep for 50-500ms.
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}

	// Close the stream and wait for the server to respond.
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed: %s", err)
	}
	log.Printf("uploaded file: %v (%v bytes)", resp.FileName, resp.Size)
}
