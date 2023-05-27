// client.go
// this is just a simple validator of the server implementation
// for me to play around with uploading different files

package main

import (
	"context"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"

	uploadpb "github.com/benjamin-rood/x-grpc/proto"
	"google.golang.org/grpc"
)

const (
	// play around with different chunk sizes
	kb        = 1024
	mb        = kb * kb
	chunkSize = 100 * kb // Upload chunks of 100KB
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("File path argument is missing.")
	}

	filePath := os.Args[1]

	// Open the file to be uploaded.
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	defer file.Close()

	// Get the file name and extension of the file
	fileName := filepath.Base(filePath)
	fileExt := filepath.Ext(fileName)
	mimeType := mime.TypeByExtension(fileExt)
	if mimeType == "" {
		log.Fatalln("can't detect mime-type of file", mimeType)
	}

	// Set up a connection to the server (using insecure because this is not real)
	conn, err := grpc.Dial(":59999", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client instance.
	client := uploadpb.NewUploaderClient(conn)

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
		if err := stream.Send(&uploadpb.UploadRequest{
			FileName: fileName,
			Chunk:    chunk,
			MimeType: mimeType,
		}); err != nil {
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
