package test

import (
	"context"
	"testing"

	"github.com/aacuadras/ha-utils/lib/docker"
	"github.com/stretchr/testify/assert"
)

func TestCreateDestroyDockerContainer(t *testing.T) {
	containerSettings := &docker.Settings{
		ImageName:     "alpine",
		ContainerName: "testContainer",
		EnvVars: []string{
			"TZ=America/Chicago",
		},
	}

	id, _ := docker.StartContainer(containerSettings, context.Background())

	containerIds, err := docker.ListContainerIDs(context.Background())

	assert.Nil(t, err)
	assert.Contains(t, containerIds, id)

	docker.StopContainer(containerSettings)
	containerIds, _ = docker.ListContainerIDs(context.Background())

	assert.NotContains(t, containerIds, id)
}
