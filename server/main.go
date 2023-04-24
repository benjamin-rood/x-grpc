// server/main.go

package main

import (
	"log"
	"net"

	uploadpb "github.com/benjamin-rood/x-grpc/proto"
	"google.golang.org/grpc"
)

func main() {
	// initialise TCP listener with a random port unlikely to conflict
	ln, err := net.Listen("tcp", ":50080")
	if err != nil {
		log.Fatalf("could not initialise tcp listener: %s", err)
	}
	defer ln.Close()

	uploadService := &Uploader{}
	grpcServer := grpc.NewServer()

	uploadpb.RegisterUploaderServer(grpcServer, uploadService)
	log.Fatal(grpcServer.Serve(ln))
}
