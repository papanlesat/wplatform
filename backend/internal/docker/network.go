package docker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

const (
	PublicNetworkName = "wp-public"
	NetworkDriver     = "bridge"
)

type NetworkManager struct {
	cli *Client
}

func NewNetworkManager(cli *Client) *NetworkManager {
	return &NetworkManager{cli: cli}
}

func (nm *NetworkManager) CreateProjectNetwork(ctx context.Context, projectID uuid.UUID) (string, error) {
	networkName := fmt.Sprintf("project_%s_net", projectID.String())

	resp, err := nm.cli.CreateNetwork(ctx, networkName, NetworkDriver)
	if err != nil {
		return "", fmt.Errorf("failed to create project network %s: %w", networkName, err)
	}

	return resp.ID, nil
}

func (nm *NetworkManager) AttachContainerToNetwork(ctx context.Context, containerID, networkID string) error {
	if err := nm.cli.ConnectNetwork(ctx, networkID, containerID); err != nil {
		return fmt.Errorf("failed to attach container %s to network %s: %w", containerID, networkID, err)
	}

	return nil
}

func (nm *NetworkManager) AttachContainerToPublicNetwork(ctx context.Context, containerID string) error {
	publicNetworkID, err := nm.cli.GetNetworkID(ctx, PublicNetworkName)
	if err != nil {
		return fmt.Errorf("failed to get wp-public network ID: %w", err)
	}

	return nm.AttachContainerToNetwork(ctx, containerID, publicNetworkID)
}

func (nm *NetworkManager) AttachContainerToProjectNetwork(ctx context.Context, containerID string, projectID uuid.UUID) error {
	networkID, err := nm.cli.GetNetworkID(ctx, fmt.Sprintf("project_%s_net", projectID.String()))
	if err != nil {
		return fmt.Errorf("failed to get project network ID: %w", err)
	}

	return nm.AttachContainerToNetwork(ctx, containerID, networkID)
}

func (nm *NetworkManager) RemoveProjectNetwork(ctx context.Context, projectID uuid.UUID) error {
	networkName := fmt.Sprintf("project_%s_net", projectID.String())

	networkID, err := nm.cli.GetNetworkID(ctx, networkName)
	if err != nil {
		return fmt.Errorf("failed to get project network %s ID: %w", networkName, err)
	}

	if err := nm.cli.RemoveNetwork(ctx, networkID); err != nil {
		return fmt.Errorf("failed to remove project network %s: %w", networkName, err)
	}

	return nil
}

func (nm *NetworkManager) NetworkExists(ctx context.Context, networkName string) (bool, error) {
	_, err := nm.cli.GetNetworkID(ctx, networkName)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (nm *NetworkManager) GetNetworkID(ctx context.Context, networkName string) (string, error) {
	return nm.cli.GetNetworkID(ctx, networkName)
}
