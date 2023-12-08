package main

import (
	"context"
	"time"

	"github.com/aacuadras/ha-utils/lib/docker"
)

func main() {
	containerSettings := &docker.Settings{
		ImageName:     "homeassistant/home-assistant",
		ContainerName: "homeassistant",
		EnvVars: []string{
			"TZ=America/Chicago",
		},
	}

	_, err := docker.StartContainer(containerSettings, context.Background())
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 30)

	docker.StopContainer(containerSettings)
}
