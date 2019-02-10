package main

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"fmt"
	"os"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"strings"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		panic(err)
	}

	cli.NegotiateAPIVersion(ctx)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})

	if err != nil {
		panic(err)
	}

	networks := make([]string, 0)
	containerNames := make([]string, 0)
	webContainers := make([]types.Container, 0)

	for _, activeContainer := range containers  {
		if ! isWebServer(activeContainer) {
			continue
		}

		name := activeContainer.Names[0]
		containerNames = append(containerNames, name)
		webContainers = append(webContainers, activeContainer)
		networkSettings := activeContainer.NetworkSettings
		for name := range networkSettings.Networks {
			if ! contains(networks, name) {
				networks = append(networks, name)
			}
		}
	}

	args := filters.NewArgs()
	args.Add("name", "hangar")

	hangars, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: args})

	if len(hangars) > 0 {
		cli.ContainerRemove(ctx, hangars[0].ID, types.ContainerRemoveOptions{})
	}

	networkConfig := map[string]*network.EndpointSettings {
	}
	for _, networkName := range networks {
		networkConfig[networkName] = &network.EndpointSettings{}
	}

	// write caddyfile
	f, err := os.Create("/tmp/hangarconf")
	check(err)
	defer f.Close()

	for _, webContainer := range webContainers  {
		availableNetworks := make([]string, 0)
		for networkName := range webContainer.NetworkSettings.Networks {
			availableNetworks = append(availableNetworks, networkName)
		}

		domain := strings.Replace(webContainer.Names[0][1:], "_", "-", -1)

		res := fmt.Sprintf("%s.test:443 {\n    tls self_signed\n    proxy / %s\n}\n", domain, webContainer.NetworkSettings.Networks[availableNetworks[0]].IPAddress)
		f.WriteString(res)
	}
	volumeMounts := []string{
		"/tmp/hangarconf:/etc/Caddyfile",
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "abiosoft/caddy",
		ExposedPorts: nat.PortSet{
			"80/tcp": struct{}{},
			"443/tcp": struct{}{},
		},
	}, &container.HostConfig{Binds: volumeMounts, PortBindings: nat.PortMap{
		"80/tcp": []nat.PortBinding{
			{
				HostIP: "0.0.0.0",
				HostPort: "80",
			},
		},
		"443/tcp": []nat.PortBinding{
			{
				HostIP: "0.0.0.0",
				HostPort: "443",
			},
		},
		},
	}, &network.NetworkingConfig{
		EndpointsConfig: networkConfig,
	}, "hangar")

	if err != nil {
		panic(err)
	}


	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})

	if err != nil {
		panic(err)
	}

	for _, name := range networks {
		fmt.Fprintln(os.Stdout, name)
	}
}

func isWebServer(container types.Container) bool {
	for _, port := range container.Ports {
		if port.PrivatePort == 80 {
			return true
		}
	}

	return false
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}


func check(e error) {
	if e != nil {
		panic(e)
	}
}