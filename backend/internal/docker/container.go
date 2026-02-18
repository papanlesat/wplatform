package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/google/uuid"
)

const (
	MySQLImage            = "mysql:8.0"
	DefaultWordPressImage = "wordpress:6-php8.2-apache"
)

type ContainerManager struct {
	cli            *Client
	wordpressImage string
}

func NewContainerManager(cli *Client) *ContainerManager {
	return &ContainerManager{
		cli:            cli,
		wordpressImage: DefaultWordPressImage,
	}
}

func (cm *ContainerManager) SetWordPressImage(image string) {
	if image != "" {
		cm.wordpressImage = image
	}
}

func (cm *ContainerManager) GetWordPressImage() string {
	return cm.wordpressImage
}

func (cm *ContainerManager) GenerateTraefikLabels(projectID uuid.UUID, domain string) map[string]string {
	routerName := fmt.Sprintf("project-%s", projectID)

	return map[string]string{
		"traefik.enable": "true",
		"traefik.http.routers." + routerName + ".rule":                      fmt.Sprintf("Host(`%s`)", domain),
		"traefik.http.routers." + routerName + ".entrypoints":               "web,websecure",
		"traefik.http.routers." + routerName + ".tls":                       "true",
		"traefik.http.routers." + routerName + ".tls.certresolver":          "myresolver",
		"traefik.http.services." + routerName + ".loadbalancer.server.port": "80",
	}
}

func (cm *ContainerManager) CreateMySQLContainer(ctx context.Context, projectID uuid.UUID, dbName, dbUser, dbPassword, volumeName, networkID string) (string, error) {
	containerName := fmt.Sprintf("mysql-%s", projectID)

	config := &container.Config{
		Image: MySQLImage,
		Env: []string{
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", dbPassword),
			fmt.Sprintf("MYSQL_DATABASE=%s", dbName),
			fmt.Sprintf("MYSQL_USER=%s", dbUser),
			fmt.Sprintf("MYSQL_PASSWORD=%s", dbPassword),
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/var/lib/mysql", volumeName),
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"": {},
		},
	}

	resp, err := cm.cli.CreateContainer(ctx, config, hostConfig, networkingConfig, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create MySQL container: %w", err)
	}

	if err := cm.cli.ConnectNetwork(ctx, networkID, resp.ID); err != nil {
		cm.cli.RemoveContainer(ctx, resp.ID, true)
		return "", fmt.Errorf("failed to connect MySQL container to network: %w", err)
	}

	return resp.ID, nil
}

func (cm *ContainerManager) CreateWordPressContainer(ctx context.Context, projectID uuid.UUID, domain string, dbName, dbUser, dbPassword, wpVolumeName, projectNetworkID, publicNetworkID string, cpuLimit float64, memoryLimitMB int, envVars map[string]string) (string, error) {
	containerName := fmt.Sprintf("wordpress-%s", projectID)

	labels := cm.GenerateTraefikLabels(projectID, domain)
	labels["project_id"] = projectID.String()

	env := []string{
		fmt.Sprintf("WORDPRESS_DB_HOST=mysql-%s", projectID.String()),
		fmt.Sprintf("WORDPRESS_DB_NAME=%s", dbName),
		fmt.Sprintf("WORDPRESS_DB_USER=%s", dbUser),
		fmt.Sprintf("WORDPRESS_DB_PASSWORD=%s", dbPassword),
		fmt.Sprintf("WORDPRESS_TABLE_PREFIX=wp_"),
	}

	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	config := &container.Config{
		Image:  cm.wordpressImage,
		Env:    env,
		Labels: labels,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/var/www/html", wpVolumeName),
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		Resources: container.Resources{
			NanoCPUs: int64(cpuLimit * 1e9),
			Memory:   int64(memoryLimitMB * 1024 * 1024),
		},
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"": {},
		},
	}

	resp, err := cm.cli.CreateContainer(ctx, config, hostConfig, networkingConfig, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create WordPress container: %w", err)
	}

	if err := cm.cli.ConnectNetwork(ctx, projectNetworkID, resp.ID); err != nil {
		cm.cli.RemoveContainer(ctx, resp.ID, true)
		return "", fmt.Errorf("failed to connect WordPress container to project network: %w", err)
	}

	if err := cm.cli.ConnectNetwork(ctx, publicNetworkID, resp.ID); err != nil {
		cm.cli.DisconnectNetwork(ctx, projectNetworkID, resp.ID, true)
		cm.cli.RemoveContainer(ctx, resp.ID, true)
		return "", fmt.Errorf("failed to connect WordPress container to public network: %w", err)
	}

	return resp.ID, nil
}

func (cm *ContainerManager) StartContainer(ctx context.Context, containerID string) error {
	return cm.cli.StartContainer(ctx, containerID)
}

func (cm *ContainerManager) StopContainer(ctx context.Context, containerID string) error {
	timeout := 30 * time.Second
	return cm.cli.StopContainer(ctx, containerID, timeout)
}

func (cm *ContainerManager) RestartContainer(ctx context.Context, containerID string) error {
	timeout := 30 * time.Second
	return cm.cli.RestartContainer(ctx, containerID, timeout)
}

func (cm *ContainerManager) RemoveContainer(ctx context.Context, containerID string) error {
	if err := cm.cli.RemoveContainer(ctx, containerID, false); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}
	return nil
}

func (cm *ContainerManager) RemoveProjectContainers(ctx context.Context, projectID uuid.UUID) error {
	mysqlContainerName := fmt.Sprintf("mysql-%s", projectID)
	wordpressContainerName := fmt.Sprintf("wordpress-%s", projectID)

	for _, name := range []string{mysqlContainerName, wordpressContainerName} {
		containerID, err := cm.cli.GetContainerID(ctx, name)
		if err != nil {
			continue
		}

		if err := cm.StopContainer(ctx, containerID); err != nil {
			continue
		}

		if err := cm.RemoveContainer(ctx, containerID); err != nil {
			return fmt.Errorf("failed to remove container %s: %w", name, err)
		}
	}

	return nil
}

func (cm *ContainerManager) PullImage(ctx context.Context, imageName string) error {
	reader, err := cm.cli.PullImage(ctx, imageName)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()
	return nil
}

func (cm *ContainerManager) GetContainerID(ctx context.Context, containerName string) (string, error) {
	containers, err := cm.cli.ListContainers(ctx, true)
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	for _, cont := range containers {
		for _, name := range cont.Names {
			if name == "/"+containerName || name == containerName {
				return cont.ID, nil
			}
		}
	}

	return "", fmt.Errorf("container %s not found", containerName)
}

func (cm *ContainerManager) InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return cm.cli.InspectContainer(ctx, containerID)
}
