package test

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/aacuadras/ha-utils/server"
	"github.com/aacuadras/ha-utils/server/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func createClient(ctx context.Context) (pb.DockerUtilsClient, func()) {
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)

	s := grpc.NewServer()
	pb.RegisterDockerUtilsServer(s, server.NewServer())
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("error listening: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("error connecting to listener: %v", err)
	}

	connCloser := func() {
		err := listener.Close()
		if err != nil {
			log.Fatalf("error closing listener: %v", err)
		}

		s.Stop()
	}

	client := pb.NewDockerUtilsClient(conn)
	return client, connCloser
}

func TestCreateDestroyDockerContainerCall(t *testing.T) {
	ctx := context.Background()
	containerName := "container-test"

	client, closer := createClient(ctx)
	defer closer()

	request := pb.ContainerRequest{
		ContainerName: containerName,
	}

	out, err := client.StartContainer(ctx, &request)
	assert.Nil(t, err)
	assert.NotEmpty(t, out.ContainerId)
	assert.Equal(t, "running", out.Status)

	out, err = client.StopContainer(ctx, &request)
	assert.Nil(t, err)
	assert.Equal(t, "stopped", out.Status)
}

func TestGetContainerCall(t *testing.T) {
	ctx := context.Background()
	containerName := "container-test"

	client, closer := createClient(ctx)
	defer closer()

	request := pb.ContainerRequest{
		ContainerName: containerName,
	}

	client.StartContainer(ctx, &request)
	defer client.StopContainer(ctx, &request)
	out, err := client.GetContainer(ctx, &request)

	assert.Nil(t, err)
	assert.NotEmpty(t, out.ContainerId)
	assert.Equal(t, "running", out.Status)
}

func TestGetEmptyContainerCall(t *testing.T) {
	ctx := context.Background()
	containerName := "empty-container"

	client, closer := createClient(ctx)
	defer closer()

	request := pb.ContainerRequest{
		ContainerName: containerName,
	}

	out, err := client.GetContainer(ctx, &request)
	assert.Nil(t, err)
	assert.Empty(t, out.ContainerId)
	assert.Empty(t, out.Status)
}
