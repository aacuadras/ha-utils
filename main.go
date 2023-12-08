package main

import (
	"context"
	"time"

	"github.com/aacuadras/ha-utils/lib/docker"
	"github.com/docker/docker/client"
)

func main() {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	defer client.Close()

	containerSettings := &docker.Settings{
		ImageName:     "homeassistant/home-assistant",
		ContainerName: "homeassistant",
		EnvVars: []string{
			"TZ=America/Chicago",
		},
	}

	_, err = docker.StartContainer(client, containerSettings, context.Background())
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 30)

	docker.StopContainer(client, containerSettings)
}
