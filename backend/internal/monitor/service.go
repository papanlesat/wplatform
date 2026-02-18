package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/google/uuid"
	"wplatform/backend/internal/docker"
)

type MonitorService struct {
	dockerClient     *docker.Client
	containerManager *docker.ContainerManager
}

func NewMonitorService(dockerClient *docker.Client, containerManager *docker.ContainerManager) *MonitorService {
	return &MonitorService{
		dockerClient:     dockerClient,
		containerManager: containerManager,
	}
}

type ProjectStatsResponse struct {
	ProjectID      uuid.UUID               `json:"project_id"`
	WordPressStats *ContainerStatsResponse `json:"wordpress_stats,omitempty"`
	MySQLStats     *ContainerStatsResponse `json:"mysql_stats,omitempty"`
	DiskUsage      *DiskUsageResponse      `json:"disk_usage,omitempty"`
	CollectedAt    time.Time               `json:"collected_at"`
}

type ContainerStatsResponse struct {
	ContainerID   string  `json:"container_id"`
	ContainerName string  `json:"container_name"`
	CPUPercentage float64 `json:"cpu_percentage"`
	MemoryUsage   uint64  `json:"memory_usage"`
	MemoryLimit   uint64  `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
}

type DiskUsageResponse struct {
	UsageBytes  int64   `json:"usage_bytes"`
	TotalBytes  int64   `json:"total_bytes"`
	PercentUsed float64 `json:"percent_used"`
}

func (s *MonitorService) GetProjectStats(ctx context.Context, projectID uuid.UUID) (*ProjectStatsResponse, error) {
	wpContainerName := fmt.Sprintf("wordpress-%s", projectID)
	mysqlContainerName := fmt.Sprintf("mysql-%s", projectID)

	response := &ProjectStatsResponse{
		ProjectID:   projectID,
		CollectedAt: time.Now(),
	}

	wpContainerID, err := s.dockerClient.GetContainerID(ctx, wpContainerName)
	if err == nil {
		wpStats, err := s.getContainerStats(ctx, wpContainerID, wpContainerName)
		if err == nil {
			response.WordPressStats = &wpStats
		}
	}

	mysqlContainerID, err := s.dockerClient.GetContainerID(ctx, mysqlContainerName)
	if err == nil {
		mysqlStats, err := s.getContainerStats(ctx, mysqlContainerID, mysqlContainerName)
		if err == nil {
			response.MySQLStats = &mysqlStats
		}
	}

	diskUsage, err := s.getProjectDiskUsage(ctx, projectID)
	if err == nil {
		response.DiskUsage = diskUsage
	}

	return response, nil
}

func (s *MonitorService) getContainerStats(ctx context.Context, containerID, containerName string) (ContainerStatsResponse, error) {
	statsResp, err := s.dockerClient.ContainerStats(ctx, containerID, false)
	if err != nil {
		return ContainerStatsResponse{}, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer statsResp.Body.Close()

	body, err := io.ReadAll(statsResp.Body)
	if err != nil {
		return ContainerStatsResponse{}, fmt.Errorf("failed to read stats body: %w", err)
	}

	var stats types.StatsJSON
	if err := json.Unmarshal(body, &stats); err != nil {
		return ContainerStatsResponse{}, fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	cpuPercent := calculateCPUPercent(&stats)
	memoryPercent := calculateMemoryPercent(&stats)
	networkRx, networkTx := calculateNetworkIO(&stats)
	blockRead, blockWrite := calculateBlockIO(&stats)

	return ContainerStatsResponse{
		ContainerID:   containerID,
		ContainerName: containerName,
		CPUPercentage: cpuPercent,
		MemoryUsage:   stats.MemoryStats.Usage,
		MemoryLimit:   stats.MemoryStats.Limit,
		MemoryPercent: memoryPercent,
		NetworkRx:     networkRx,
		NetworkTx:     networkTx,
		BlockRead:     blockRead,
		BlockWrite:    blockWrite,
	}, nil
}

func calculateCPUPercent(stats *types.StatsJSON) float64 {
	if stats.CPUStats.CPUUsage.TotalUsage == 0 {
		return 0
	}

	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0 {
		cpuPercent := (cpuDelta / systemDelta) * 100
		if cpuPercent > 100 {
			return 100
		}
		return cpuPercent
	}

	return 0
}

func calculateMemoryPercent(stats *types.StatsJSON) float64 {
	if stats.MemoryStats.Limit == 0 {
		return 0
	}

	percent := (float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit)) * 100
	if percent > 100 {
		return 100
	}
	return percent
}

func calculateNetworkIO(stats *types.StatsJSON) (rx, tx uint64) {
	if stats.Networks == nil {
		return 0, 0
	}

	var totalRx, totalTx uint64
	for _, network := range stats.Networks {
		totalRx += network.RxBytes
		totalTx += network.TxBytes
	}

	return totalRx, totalTx
}

func calculateBlockIO(stats *types.StatsJSON) (read, write uint64) {
	if stats.BlkioStats.IoServiceBytesRecursive == nil {
		return 0, 0
	}

	var totalRead, totalWrite uint64
	for _, blkio := range stats.BlkioStats.IoServiceBytesRecursive {
		switch blkio.Op {
		case "Read":
			totalRead += blkio.Value
		case "Write":
			totalWrite += blkio.Value
		}
	}

	return totalRead, totalWrite
}

func (s *MonitorService) getProjectDiskUsage(ctx context.Context, projectID uuid.UUID) (*DiskUsageResponse, error) {
	wpVolumeName := fmt.Sprintf("project_%s_wp_data", projectID)
	dbVolumeName := fmt.Sprintf("project_%s_db_data", projectID)

	volumes, err := s.dockerClient.ListVolumes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	var totalUsage int64
	var totalSize int64

	for _, vol := range volumes.Volumes {
		if vol.Name == wpVolumeName || vol.Name == dbVolumeName {
			volInfo, err := s.dockerClient.InspectVolume(ctx, vol.Name)
			if err == nil && volInfo.UsageData != nil {
				totalSize += volInfo.UsageData.Size
				totalUsage += volInfo.UsageData.Size
			}
		}
	}

	percentUsed := float64(0)
	if totalSize > 0 {
		percentUsed = (float64(totalUsage) / float64(totalSize)) * 100
	}

	return &DiskUsageResponse{
		UsageBytes:  totalUsage,
		TotalBytes:  totalSize,
		PercentUsed: percentUsed,
	}, nil
}
