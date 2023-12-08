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
)

type Settings struct {
	ImageName     string
	ContainerName string
	EnvVars       []string
}

func createClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
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
