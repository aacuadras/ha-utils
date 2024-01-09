package server

import (
	"context"

	"github.com/aacuadras/ha-utils/lib/docker"
	pb "github.com/aacuadras/ha-utils/server/pb"
)

type server struct {
	pb.UnimplementedDockerUtilsServer
}

func NewServer() pb.DockerUtilsServer {
	return &server{}
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

	return &pb.ContainerResponse{ContainerId: "", Status: "stopped"}, nil
}

func (s *server) GetContainer(ctx context.Context, in *pb.ContainerRequest) (*pb.ContainerResponse, error) {
	status, err := docker.GetContainer(ctx, in.ContainerName)
	if err != nil {
		return &pb.ContainerResponse{}, nil
	}

	return &pb.ContainerResponse{
		ContainerId: status.ID,
		Status:      status.State.Status,
	}, nil
}
