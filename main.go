package main

import (
	"github.com/aacuadras/ha-utils/lib/docker"
	"github.com/docker/docker/client"
)

func main() {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	defer client.Close()

	err = docker.StartContainer(client)
	if err != nil {
		panic(err)
	}
}
