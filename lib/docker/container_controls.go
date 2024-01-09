package docker

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// Stops and deletes the docker container specified in the settings
func StopContainer(settings *Settings) error {
	context := context.Background()
	log.Printf("Stopping container %s...", settings.ContainerName)

	client, err := createClient()
	if err != nil {
		return err
	}

	defer client.Close()

	err = client.ContainerStop(context, settings.ContainerName, container.StopOptions{})

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

// Starts a docker contaienr with the image specified in the settings being sent in. It returns the ID of the new container
func StartContainer(settings *Settings, ctx context.Context) (string, error) {
	client, err := createClient()
	if err != nil {
		return "", err
	}

	defer client.Close()

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

// Lists the IDs of all the containers that are running in the machine, this is a helper method to test the creation
// of the containers
func ListContainerIDs(ctx context.Context) ([]string, error) {
	client, err := createClient()
	if err != nil {
		return []string{}, err
	}

	defer client.Close()

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

// Gets the container information based on its ID
func GetContainer(ctx context.Context, id string) (types.ContainerJSON, error) {
	client, err := createClient()
	if err != nil {
		return types.ContainerJSON{}, err
	}

	defer client.Close()

	container, err := client.ContainerInspect(ctx, id)

	if err != nil {
		return types.ContainerJSON{}, err
	}

	return container, nil
}
