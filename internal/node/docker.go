package node

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	DefaultPostgresImage = "postgres:latest"
	postgresPortNumber   = 5432
)

var (
	ErrConnection = fmt.Errorf("connection")
)

type nodeContainer struct {
	port    uint16
	name    string
	created bool
	running bool
	id      string
}

type dockerNodeConfig struct {
	// Port is the external TCP for the node to listen on
	Port uint16
	// Group is the naming prefix applied to docker containers
	Group string
	// ContainerName is the name of a specific container without the group prefix
	ContainerName string
	// Image is the chainlink docker image for the container
	Image string
	// ExtraTOML is a configuration applied at command input
	ExtraTOML string
	// BasePath is the file system path where secrets will be created and referenced by a running node
	BasePath string
	// Reset will stop an existing container set before recreating a new one
	Reset bool
}

func buildChainlinkNode(
	ctx context.Context,
	progress io.Writer,
	conf NodeConfig,
	image dockerNodeConfig,
) (*ChainlinkNode, error) {
	node, err := newNode(ctx, progress, image.Group, image.ContainerName, image.Image, image.Port)
	if err != nil {
		return nil, err
	}

	if err := ensurePostgresImage(ctx, node); err != nil {
		return nil, err
	}

	if err := ensureNetwork(ctx, node); err != nil {
		return nil, err
	}

	if err := checkContainerState(ctx, node); err != nil {
		return nil, err
	}

	if err := ensurePostgresContainer(ctx, node, image.Reset); err != nil {
		return nil, err
	}

	if err := ensureChainlinkContainer(ctx, node, conf, image.ExtraTOML, image.BasePath, image.Reset); err != nil {
		return nil, err
	}

	if err = waitForNodeReady(ctx, node); err != nil {
		return nil, err
	}

	return node, nil
}

func removeChainlinkNode(
	ctx context.Context,
	image dockerNodeConfig,
) error {
	node, err := newNode(ctx, nil, image.Group, image.ContainerName, image.Image, image.Port)
	if err != nil {
		return err
	}

	if err := checkContainerState(ctx, node); err != nil {
		return err
	}

	options := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   true,
		Force:         true,
	}

	if err := node.client.ContainerRemove(ctx, node.chainlink.id, options); err != nil {
		return fmt.Errorf("failed to remove existing container: %w", err)
	}

	if err := node.client.ContainerRemove(ctx, node.postgres.id, options); err != nil {
		return fmt.Errorf("failed to remove existing container: %w", err)
	}

	return nil
}

func newNode(ctx context.Context, writer io.Writer, group, name, image string, port uint16) (*ChainlinkNode, error) {
	dockerClient, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client from env: %w", err)
	}

	// ping the client to make sure the connection is good
	if _, err = dockerClient.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping docker server: %w", err)
	}

	return &ChainlinkNode{
		Name:           name,
		Network:        fmt.Sprintf("%s-local", group),
		PostgresImage:  DefaultPostgresImage,
		ChainlinkImage: image,
		GroupName:      group,

		client: dockerClient,
		writer: writer,
		postgres: nodeContainer{
			port: postgresPortNumber,
			name: fmt.Sprintf("%s-%s-postgres", group, name),
		},
		chainlink: nodeContainer{
			port: port,
			name: fmt.Sprintf("%s-%s", group, name),
		},
	}, nil
}

func ensurePostgresImage(ctx context.Context, node *ChainlinkNode) error {
	var out io.ReadCloser

	if _, _, err := node.client.ImageInspectWithRaw(ctx, node.PostgresImage); err != nil {
		fmt.Fprintln(node.writer, "Pulling Postgres docker image...")

		if out, err = node.client.ImagePull(ctx, node.PostgresImage, types.ImagePullOptions{}); err != nil {
			return fmt.Errorf("failed to pull Postgres image: %w", err)
		}

		out.Close()
		fmt.Fprintln(node.writer, "Postgres docker image successfully pulled!")
	}

	return nil
}

func ensureNetwork(ctx context.Context, node *ChainlinkNode) error {
	existingNetworks, err := node.client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	var found bool

	for _, ntwrk := range existingNetworks {
		if ntwrk.Name == node.Network {
			found = true

			break
		}
	}

	if !found {
		if _, err = node.client.NetworkCreate(ctx, node.Network, types.NetworkCreate{}); err != nil {
			return fmt.Errorf("failed to create network: %w", err)
		}
	}

	return nil
}

func checkContainerState(ctx context.Context, node *ChainlinkNode) error {
	search := fmt.Sprintf("%s-%s", node.GroupName, node.Name)

	opts := types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("name", "^/"+regexp.QuoteMeta(search)+".*$")),
	}

	containers, err := node.client.ContainerList(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for idx, container := range containers {
		switch container.Image {
		case node.PostgresImage:
			node.postgres = getContainerDetail(containers[idx], node.postgres)
		default:
			node.chainlink = getContainerDetail(containers[idx], node.chainlink)
		}
	}

	return nil
}

func getContainerDetail(dContainer types.Container, cont nodeContainer) nodeContainer {
	for _, port := range dContainer.Ports {
		if port.PublicPort == cont.port || port.PrivatePort == cont.port {
			cont.created = true
			cont.id = dContainer.ID

			if dContainer.Status != "running" {
				cont.running = true
			}

			return cont
		}
	}

	return cont
}

func ensurePostgresContainer(ctx context.Context, node *ChainlinkNode, reset bool) error {
	if reset && node.postgres.id != "" {
		if err := node.client.ContainerRemove(ctx, node.postgres.id, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			return fmt.Errorf("failed to remove existing container: %w", err)
		}

		node.postgres.created = false
		node.postgres.running = false
		node.postgres.id = ""
	}

	if !node.postgres.created {
		port := nat.Port(fmt.Sprintf("%d", node.postgres.port))

		response, err := node.client.ContainerCreate(
			ctx,
			&container.Config{
				Image: node.PostgresImage,
				Cmd:   []string{"postgres", "-c", `max_connections=1000`},
				Env: []string{
					"POSTGRES_USER=postgres",
					"POSTGRES_PASSWORD=verylongdatabasepassword",
				},
				ExposedPorts: nat.PortSet{port: struct{}{}},
			},
			nil,
			&network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{
					node.Network: {Aliases: []string{node.postgres.name}},
				},
			},
			nil,
			node.postgres.name,
		)

		if err != nil {
			return fmt.Errorf("failed to create Postgres container, use --force=true to force removing existing containers: %w", err)
		}

		node.postgres.created = true
		node.postgres.id = response.ID
	}

	if node.postgres.id != "" && node.postgres.created && !node.postgres.running {
		if err := node.client.ContainerStart(ctx, node.postgres.id, types.ContainerStartOptions{}); err != nil {
			return fmt.Errorf("failed to start DB container: %w", err)
		}

		time.Sleep(10 * time.Second)
	}

	return nil
}

//nolint:funlen,cyclop
func ensureChainlinkContainer(
	ctx context.Context,
	node *ChainlinkNode,
	conf NodeConfig,
	extraTOML string,
	basePath string,
	reset bool,
) error {
	if reset && node.chainlink.id != "" {
		if err := node.client.ContainerRemove(ctx, node.chainlink.id, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			return fmt.Errorf("failed to remove existing container: %w", err)
		}

		node.chainlink.created = false
		node.chainlink.running = false
		node.chainlink.id = ""
	}

	//nolint:nestif
	if !node.chainlink.created {
		portStr := fmt.Sprintf("%d", node.chainlink.port)
		port := nat.Port(portStr)

		path := fmt.Sprintf("%s/secrets", basePath)

		if err := writeCredentials(path); err != nil {
			return fmt.Errorf("failed to create creds files: %w", err)
		}

		if err := writeFile(fmt.Sprintf("%s/01-config.toml", path), NodeTOML(conf)); err != nil {
			return err
		}

		if err := writeFile(fmt.Sprintf("%s/01-secret.toml", path), SecretTOML(conf)); err != nil {
			return err
		}

		response, err := node.client.ContainerCreate(
			ctx,
			&container.Config{
				Image: node.ChainlinkImage,
				Cmd: []string{
					"-s", "/run/secrets/01-secret.toml",
					"-c", "/run/secrets/01-config.toml",
					"local", "n",
					"-a", "/run/secrets/chainlink-node-api",
				},
				Env: []string{
					"CL_CONFIG=" + extraTOML,
					"CL_PASSWORD_KEYSTORE=" + DefaultChainlinkNodePassword,
					"CL_DATABASE_URL=postgresql://postgres:verylongdatabasepassword@" + node.postgres.name + ":5432/postgres?sslmode=disable",
				},
				ExposedPorts: map[nat.Port]struct{}{
					port: {},
				},
			},
			&container.HostConfig{
				// Binds: []string{fmt.Sprintf("%s:/run/secrets", path)},
				Mounts: []mount.Mount{
					{
						Type:     mount.TypeBind,
						Source:   path,
						Target:   "/run/secrets",
						ReadOnly: true,
						BindOptions: &mount.BindOptions{
							Propagation:      mount.PropagationRShared,
							CreateMountpoint: true,
						},
					},
				},
				PortBindings: nat.PortMap{
					"6688/tcp": []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: portStr,
						},
					},
				},
			},
			&network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{
					node.Network: {Aliases: []string{node.chainlink.name}},
				},
			},
			nil,
			node.chainlink.name,
		)

		if err != nil {
			return fmt.Errorf("failed to create node container, use --force=true to force removing existing containers: %w", err)
		}

		node.chainlink.created = true
		node.chainlink.id = response.ID
	}

	if node.chainlink.id != "" && node.chainlink.created && !node.chainlink.running {
		if err := node.client.ContainerStart(ctx, node.chainlink.id, types.ContainerStartOptions{}); err != nil {
			return fmt.Errorf("failed to start chainlink container: %w", err)
		}
	}

	return nil
}

func waitForNodeReady(ctx context.Context, node *ChainlinkNode) error {
	addr := fmt.Sprintf("http://localhost:%d", node.chainlink.port)
	client := &http.Client{}

	defer client.CloseIdleConnections()

	const timeout = 120

	startTime := time.Now().Unix()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/health", addr), nil)
		if err != nil {
			return fmt.Errorf("%w: failed to make request to health status: %s", ErrConnection, err.Error())
		}

		req.Close = true

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		if time.Now().Unix()-startTime > int64(timeout*time.Second) {
			return fmt.Errorf("%w: timed out waiting for node to start, waited %d seconds", ErrConnection, timeout)
		}

		time.Sleep(5 * time.Second) //nolint:gomnd
	}
}
