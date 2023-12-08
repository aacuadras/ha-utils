package docker

import (
	"context"
	"io"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Settings struct {
	ImageName     string
	ContainerName string
	EnvVars       []string
}

func pullImage(client *client.Client, imageName string) error {
	reader, err := client.ImagePull(context.Background(), imageName, types.ImagePullOptions{})

	if err != nil {
		return err
	}

	defer reader.Close()
	io.Copy(log.Default().Writer(), reader)

	return nil
}

func setContainerSettings() (*container.HostConfig, *network.NetworkingConfig, map[nat.Port]struct{}, error) {
	portNumber := "8123"
	port, err := nat.NewPort("tcp", portNumber)
	if err != nil {
		return nil, nil, nil, err
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
		port: {},
	}

	return hostConfig, networkConfig, exposedPorts, nil
}

func StopContainer(client *client.Client, settings *Settings) error {
	context := context.Background()
	log.Printf("Stopping container %s...", settings.ContainerName)

	err := client.ContainerStop(context, settings.ContainerName, container.StopOptions{})

	if err != nil {
		log.Panic("Unable to stop home assistant container")
		return err
	}

	err = client.ContainerRemove(context, settings.ContainerName, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})

	if err != nil {
		log.Panic("Unable to remove container")
		return err
	}

	log.Printf("Successfully stopped container %s!", settings.ContainerName)
	return nil
}

func StartContainer(client *client.Client, settings *Settings, ctx context.Context) (string, error) {
	hostConfig, networkConfig, exposedPorts, err := setContainerSettings()

	if err != nil {
		return "", err
	}

	config := &container.Config{
		Image:        settings.ImageName,
		Env:          settings.EnvVars,
		ExposedPorts: exposedPorts,
		Hostname:     settings.ContainerName,
	}

	err = pullImage(client, settings.ImageName)
	if err != nil {
		return "", err
	}

	log.Printf("Creating %s Container...", settings.ContainerName)
	cont, err := client.ContainerCreate(
		ctx,
		config,
		hostConfig,
		networkConfig,
		&v1.Platform{},
		settings.ContainerName,
	)

	if err != nil {
		return "", err
	}

	client.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	log.Printf("Container %s Started! (%s)", settings.ContainerName, cont.ID)

	return cont.ID, nil
}

func ListContainerIDs(client *client.Client, ctx context.Context) ([]string, error) {
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		return []string{}, err
	}

	var containerIds []string
	for _, container := range containers {
		containerIds = append(containerIds, container.ID)
	}

	return containerIds, nil
}
