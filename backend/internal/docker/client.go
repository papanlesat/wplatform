package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	system "github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

func (c *Client) Close() error {
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}

func (c *Client) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	return c.cli.ContainerList(ctx, container.ListOptions{All: all})
}

func (c *Client) CreateContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, name string) (container.CreateResponse, error) {
	return c.cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, name)
}

func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (c *Client) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	timeoutSeconds := int(timeout.Seconds())
	return c.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
}

func (c *Client) RestartContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	timeoutSeconds := int(timeout.Seconds())
	return c.cli.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
}

func (c *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return c.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: force})
}

func (c *Client) GetContainerLogs(ctx context.Context, containerID string, tail string) (io.ReadCloser, error) {
	return c.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
		Follow:     false,
		Timestamps: true,
	})
}

func (c *Client) StreamContainerLogs(ctx context.Context, containerID string, tail string) (io.ReadCloser, error) {
	return c.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
		Follow:     true,
		Timestamps: true,
	})
}

func (c *Client) InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return c.cli.ContainerInspect(ctx, containerID)
}

func (c *Client) GetContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error) {
	return c.cli.ContainerStats(ctx, containerID, stream)
}

func (c *Client) WaitContainer(ctx context.Context, containerID string) (int64, error) {
	statusCh, errCh := c.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case status := <-statusCh:
		return status.StatusCode, nil
	case err := <-errCh:
		return 0, err
	}
}

func (c *Client) ListNetworks(ctx context.Context) ([]types.NetworkResource, error) {
	return c.cli.NetworkList(ctx, types.NetworkListOptions{})
}

func (c *Client) CreateNetwork(ctx context.Context, name string, driver string) (types.NetworkCreateResponse, error) {
	return c.cli.NetworkCreate(ctx, name, types.NetworkCreate{
		Driver: driver,
	})
}

func (c *Client) ConnectNetwork(ctx context.Context, networkID, containerID string) error {
	return c.cli.NetworkConnect(ctx, networkID, containerID, &network.EndpointSettings{})
}

func (c *Client) DisconnectNetwork(ctx context.Context, networkID, containerID string, force bool) error {
	return c.cli.NetworkDisconnect(ctx, networkID, containerID, force)
}

func (c *Client) RemoveNetwork(ctx context.Context, networkID string) error {
	return c.cli.NetworkRemove(ctx, networkID)
}

func (c *Client) InspectNetwork(ctx context.Context, networkID string) (types.NetworkResource, error) {
	return c.cli.NetworkInspect(ctx, networkID, types.NetworkInspectOptions{})
}

func (c *Client) ListVolumes(ctx context.Context) (volume.ListResponse, error) {
	return c.cli.VolumeList(ctx, volume.ListOptions{})
}

func (c *Client) CreateVolume(ctx context.Context, name, driver string) (volume.Volume, error) {
	resp, err := c.cli.VolumeCreate(ctx, volume.CreateOptions{
		Name:   name,
		Driver: driver,
	})
	if err != nil {
		return volume.Volume{}, err
	}
	return resp, nil
}

func (c *Client) RemoveVolume(ctx context.Context, volumeID string, force bool) error {
	return c.cli.VolumeRemove(ctx, volumeID, force)
}

func (c *Client) InspectVolume(ctx context.Context, volumeID string) (volume.Volume, error) {
	return c.cli.VolumeInspect(ctx, volumeID)
}

func (c *Client) PullImage(ctx context.Context, ref string) (io.ReadCloser, error) {
	return c.cli.ImagePull(ctx, ref, image.PullOptions{})
}

func (c *Client) CreateImageFromContainer(ctx context.Context, containerID string, options container.CommitOptions) (types.IDResponse, error) {
	return c.cli.ContainerCommit(ctx, containerID, options)
}

func (c *Client) ExecCreate(ctx context.Context, containerID string, config types.ExecConfig) (types.IDResponse, error) {
	return c.cli.ContainerExecCreate(ctx, containerID, config)
}

func (c *Client) ExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error {
	return c.cli.ContainerExecStart(ctx, execID, config)
}

func (c *Client) ExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	return c.cli.ContainerExecInspect(ctx, execID)
}

func (c *Client) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
	return c.cli.ContainerExecAttach(ctx, execID, config)
}

func (c *Client) ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error) {
	return c.cli.ContainerStats(ctx, containerID, stream)
}

func (c *Client) SystemInfo(ctx context.Context) (system.Info, error) {
	return c.cli.Info(ctx)
}

func (c *Client) SystemDiskUsage(ctx context.Context) (types.DiskUsage, error) {
	return c.cli.DiskUsage(ctx, types.DiskUsageOptions{})
}

func (c *Client) ListImages(ctx context.Context, options types.ImageListOptions) ([]image.Summary, error) {
	return c.cli.ImageList(ctx, options)
}

func (c *Client) RemoveImage(ctx context.Context, imageID string, force, pruneChildren bool) ([]image.DeleteResponse, error) {
	return c.cli.ImageRemove(ctx, imageID, image.RemoveOptions{
		Force:         force,
		PruneChildren: pruneChildren,
	})
}

func (c *Client) GetNetworkID(ctx context.Context, name string) (string, error) {
	networks, err := c.ListNetworks(ctx)
	if err != nil {
		return "", err
	}

	for _, net := range networks {
		if net.Name == name {
			return net.ID, nil
		}
	}

	return "", fmt.Errorf("network %s not found", name)
}

func (c *Client) GetContainerID(ctx context.Context, name string) (string, error) {
	containers, err := c.ListContainers(ctx, true)
	if err != nil {
		return "", err
	}

	for _, cont := range containers {
		for _, cname := range cont.Names {
			if cname == "/"+name || cname == name {
				return cont.ID, nil
			}
		}
	}

	return "", fmt.Errorf("container %s not found", name)
}
