package main

import (
	"log"
	"net"

	"github.com/aacuadras/ha-utils/server"
	pb "github.com/aacuadras/ha-utils/server/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Start the grpc server on port 8080
func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	s := grpc.NewServer(opts...)
	pb.RegisterDockerUtilsServer(s, server.NewServer())
	pb.RegisterFileUtilsServer(s, server.NewFileServer())
	reflection.Register(s)
	s.Serve(listener)
}
