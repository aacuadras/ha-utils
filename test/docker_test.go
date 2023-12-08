package test

import (
	"context"
	"testing"

	"github.com/aacuadras/ha-utils/lib/docker"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestCreateDestroyDockerContainer(t *testing.T) {
	client, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	defer client.Close()

	containerSettings := &docker.Settings{
		ImageName:     "alpine",
		ContainerName: "testContainer",
		EnvVars: []string{
			"TZ=America/Chicago",
		},
	}

	id, _ := docker.StartContainer(client, containerSettings, context.Background())

	containerIds, err := docker.ListContainerIDs(client, context.Background())

	assert.Nil(t, err)
	assert.Contains(t, containerIds, id)

	docker.StopContainer(client, containerSettings)
	containerIds, _ = docker.ListContainerIDs(client, context.Background())

	assert.NotContains(t, containerIds, id)
}
