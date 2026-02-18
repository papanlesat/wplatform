package docker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type VolumeManager struct {
	cli *Client
}

func NewVolumeManager(cli *Client) *VolumeManager {
	return &VolumeManager{cli: cli}
}

func (vm *VolumeManager) CreateProjectVolumes(ctx context.Context, projectID uuid.UUID) (wpVolumeID, dbVolumeID string, err error) {
	wpVolumeName := fmt.Sprintf("project_%s_wp_data", projectID.String())
	dbVolumeName := fmt.Sprintf("project_%s_db_data", projectID.String())

	wpVol, err := vm.cli.CreateVolume(ctx, wpVolumeName, "local")
	if err != nil {
		return "", "", fmt.Errorf("failed to create WordPress volume %s: %w", wpVolumeName, err)
	}

	dbVol, err := vm.cli.CreateVolume(ctx, dbVolumeName, "local")
	if err != nil {
		vm.cli.RemoveVolume(ctx, wpVol.Name, false)
		return "", "", fmt.Errorf("failed to create database volume %s: %w", dbVolumeName, err)
	}

	return wpVol.Name, dbVol.Name, nil
}

func (vm *VolumeManager) RemoveProjectVolumes(ctx context.Context, projectID uuid.UUID) error {
	wpVolumeName := fmt.Sprintf("project_%s_wp_data", projectID.String())
	dbVolumeName := fmt.Sprintf("project_%s_db_data", projectID.String())

	if err := vm.cli.RemoveVolume(ctx, wpVolumeName, true); err != nil {
		return fmt.Errorf("failed to remove WordPress volume %s: %w", wpVolumeName, err)
	}

	if err := vm.cli.RemoveVolume(ctx, dbVolumeName, true); err != nil {
		return fmt.Errorf("failed to remove database volume %s: %w", dbVolumeName, err)
	}

	return nil
}

func (vm *VolumeManager) VolumeExists(ctx context.Context, volumeName string) (bool, error) {
	_, err := vm.cli.InspectVolume(ctx, volumeName)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (vm *VolumeManager) GetVolumeSize(ctx context.Context, volumeName string) (int64, error) {
	vol, err := vm.cli.InspectVolume(ctx, volumeName)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect volume %s: %w", volumeName, err)
	}

	if vol.UsageData == nil {
		return 0, nil
	}

	return vol.UsageData.Size, nil
}
