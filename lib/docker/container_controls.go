package docker

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func pullImage(client *client.Client, imageName string) error {
	reader, err := client.ImagePull(context.Background(), imageName, types.ImagePullOptions{})

	if err != nil {
		return err
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	return nil
}

func StopContainer(client *client.Client, container_name string) error {
	context := context.Background()
	log.Printf("Stopping container %s...", container_name)

	err := client.ContainerStop(context, container_name, container.StopOptions{})

	if err != nil {
		log.Panic("Unable to stop home assistant container")
		return err
	}

	err = client.ContainerRemove(context, container_name, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})

	if err != nil {
		log.Panic("Unable to remove container")
		return err
	}

	log.Printf("Successfully stopped container %s!", container_name)
	return nil
}

func StartContainer(client *client.Client) error {
	portNumber := "8123"
	port, err := nat.NewPort("tcp", portNumber)
	if err != nil {
		return err
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: portNumber,
				},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}

	endpointConfig := &network.EndpointSettings{
		Gateway: "gatewayname",
	}
	networkConfig.EndpointsConfig["bridge"] = endpointConfig

	exposedPorts := map[nat.Port]struct{}{
		port: struct{}{},
	}

	config := &container.Config{
		Image: "homeassistant/home-assistant",
		Env: []string{
			"TZ=America/Chicago",
		},
		ExposedPorts: exposedPorts,
		Hostname:     "homeassistant",
	}

	err = pullImage(client, "homeassistant/home-assistant")
	if err != nil {
		return err
	}

	log.Printf("Creating Container...")
	cont, err := client.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkConfig,
		&v1.Platform{},
		"homeassistant",
	)

	if err != nil {
		return err
	}

	client.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	log.Printf("Container Started! (%s)", cont.ID)

	return nil
}
