package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	units "github.com/docker/go-units"
	"github.com/pkg/errors"
)

type DockerRun struct {
	client                *client.Client
	ctx                   context.Context
	projectName           string
	guestRootPath         string
	guestCachePath        string
	guestProjectCachePath string
	imageTag              string
	hostRootPath          string
	hostCachePath         string
	hostGID               int
	hostUID               int
}

const (
	imageTag           = "hf-torch:latest"
	guestRootPath      = "/srv/"
	guestCachePath     = "/home/nonroot/.cache/"
	guestRootCachePath = "/root/.cache/"
)

func isCos() (bool, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			id := strings.TrimPrefix(line, "ID=")
			return id == "cos", nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("failed to scan file: %w", err)
	}

	return false, nil
}

func NewDockerRun(
	ctx context.Context,
	projectName,
	hostRootPath,
	hostCachePath string,
) *DockerRun {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	hostGID := os.Getgid()
	hostUID := os.Getuid()

	return &DockerRun{
		client:                cli,
		ctx:                   ctx,
		projectName:           projectName,
		guestRootPath:         guestRootPath,
		guestCachePath:        guestCachePath,
		guestProjectCachePath: guestCachePath + projectName,
		imageTag:              imageTag,
		hostRootPath:          hostRootPath,
		hostCachePath:         hostCachePath,
		hostGID:               hostGID,
		hostUID:               hostUID,
	}
}

func DefaultProjExpContainerName(projectName, experimentName string) string {
	return fmt.Sprintf("%s-%s", projectName, experimentName)
}

func (d *DockerRun) Kill(containerName string) error {
	options := types.ContainerListOptions{All: true, Filters: filters.NewArgs(filters.Arg("name", containerName))}

	containers, err := d.client.ContainerList(d.ctx, options)
	if err != nil {
		return errors.WithMessagef(err, "failed to list containers with name %s", containerName)
	}

	fmt.Printf("found %d containers with name %s\n", len(containers), containerName)

	for _, c := range containers {
		if c.Status == "running" {
			fmt.Printf("stopping container %s\n", c.ID)
			if err := d.client.ContainerStop(d.ctx, c.ID, container.StopOptions{Timeout: PtrTo(0)}); err != nil {
				fmt.Printf("failed to stop container %s, reason: %v", c.ID, err)
			}
		}

		fmt.Printf("removing container %s\n", c.ID)
		if err := d.client.ContainerRemove(d.ctx, c.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
			return errors.WithMessagef(err, "failed to remove container %s", c.ID)
		}
	}

	return nil
}

func (d *DockerRun) Run(
	containerName string,
	runCommand string,
	runCommandArgs []string,
	exposePort int,
) error {

	fmt.Printf("killing container %s\n", containerName)
	if err := d.Kill(containerName); err != nil {
		return errors.WithMessagef(err, "failed to kill container %s", containerName)
	}

	buildCtx, err := archive.TarWithOptions(d.hostRootPath, &archive.TarOptions{})
	if err != nil {
		panic(err)
	}
	defer buildCtx.Close()

	fmt.Printf("rebuilding image %s\n", d.imageTag)
	buildOptions := types.ImageBuildOptions{
		Tags: []string{d.imageTag},
		BuildArgs: map[string]*string{
			"GID": PtrTo(fmt.Sprintf("%d", d.hostGID)),
			"UID": PtrTo(fmt.Sprintf("%d", d.hostUID)),
		},
		Remove:      true, // Remove intermediate containers after the build
		ForceRemove: true, // Force removal of the image if it exists
	}

	buildResponse, err := d.client.ImageBuild(d.ctx, buildCtx, buildOptions)
	if err != nil {
		return errors.WithMessagef(err, "failed to build image %s", d.imageTag)
	}

	defer buildResponse.Body.Close()

	fmt.Printf("building image %s\n", d.imageTag)
	if _, err := io.Copy(os.Stdout, buildResponse.Body); err != nil {
		return errors.WithMessagef(err, "failed to build image %s", d.imageTag)
	}

	// check if host has gpu
	// if yes, add gpu to device requests
	// else, don't add gpu to device requests
	// this is a hacky way to get around the fact that docker doesn't support
	// gpu passthrough on macos
	dr := make([]container.DeviceRequest, 0, 1)

	if _, err := os.Stat("/dev/nvidia0"); err == nil {
		fmt.Printf("host has gpu, adding gpu to device requests\n")
		if isCos, _ := isCos(); isCos {
			fmt.Printf("host is cos, not adding gpu to device requests\n")
		} else {
			dr = append(dr, container.DeviceRequest{
				Count:        -1,
				Capabilities: [][]string{{"gpu"}},
			})
		}
	} else {
		fmt.Printf("host does not have gpu, not adding gpu to device requests\n")
	}

	fmt.Printf("creating container %s\n", containerName)
	createOptions := types.ContainerCreateConfig{
		Name: containerName,
		Config: &container.Config{
			Image:      d.imageTag,
			Entrypoint: append([]string{runCommand}, runCommandArgs...),
		},
		HostConfig: &container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:%s", d.hostRootPath, d.guestRootPath),
				fmt.Sprintf("%s:%s", d.hostCachePath, d.guestCachePath),
				fmt.Sprintf("%s:%s", d.hostCachePath, guestRootCachePath),
			},
			IpcMode:     container.IPCModeHost,
			PidMode:     container.PidMode("host"),
			NetworkMode: container.NetworkMode("host"),
			Resources: container.Resources{
				DeviceRequests: dr,
				Ulimits: []*units.Ulimit{
					{
						Name: "memlock",
						Soft: -1,
						Hard: -1,
					},
					{
						Name: "stack",
						Soft: 67108864,
						Hard: 67108864,
					},
				},
			},
			Privileged: true,
		},
	}

	resp, err := d.client.ContainerCreate(d.ctx, createOptions.Config, createOptions.HostConfig, nil, nil, containerName)
	if err != nil {
		return errors.WithMessagef(err, "failed to create container %s", containerName)
	}

	fmt.Printf("starting container %s\n", containerName)
	if err := d.client.ContainerStart(d.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return errors.WithMessagef(err, "failed to start container %s", containerName)
	}

	fmt.Printf("started container %s\n", containerName)

	return nil
}

func PtrTo[T any](e T) *T {
	return &e
}
