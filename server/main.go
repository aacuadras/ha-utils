package main

import (
	"context"
	"log"
	"net"

	"github.com/aacuadras/ha-utils/lib/docker"
	pb "github.com/aacuadras/ha-utils/server/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedDockerUtilsServer
}

func (s *server) StartContainer(ctx context.Context, in *pb.ContainerRequest) (*pb.ContainerResponse, error) {
	containerSettings := &docker.Settings{
		ImageName:     "homeassistant/home-assistant",
		ContainerName: in.ContainerName,
		EnvVars: []string{
			"TZ=America/Chicago",
		},
	}

	id, err := docker.StartContainer(containerSettings, ctx)

	if err != nil {
		return nil, err
	}

	containerInfo, err := docker.GetContainer(ctx, id)

	if err != nil {
		return nil, err
	}

	return &pb.ContainerResponse{Status: containerInfo.State.Status, ContainerId: id}, nil
}

func (s *server) StopContainer(ctx context.Context, in *pb.ContainerRequest) (*pb.ContainerResponse, error) {
	// The only important part in the options is the container name
	containerSettings := &docker.Settings{
		ContainerName: in.ContainerName,
	}

	if err := docker.StopContainer(containerSettings); err != nil {
		return nil, err
	}

	containerStatus, err := docker.GetContainer(ctx, in.ContainerName)

	if err != nil {
		return nil, err
	}

	return &pb.ContainerResponse{ContainerId: "", Status: containerStatus.State.Status}, nil
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	s := grpc.NewServer(opts...)
	pb.RegisterDockerUtilsServer(s, &server{})
	reflection.Register(s)
	s.Serve(listener)
}
