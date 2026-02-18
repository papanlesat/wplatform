package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         UserRole  `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type ProjectStatus string

const (
	StatusCreating ProjectStatus = "creating"
	StatusRunning  ProjectStatus = "running"
	StatusStopped  ProjectStatus = "stopped"
	StatusError    ProjectStatus = "error"
)

type Project struct {
	ID                     uuid.UUID     `json:"id" db:"id"`
	UserID                 uuid.UUID     `json:"user_id" db:"user_id"`
	Name                   string        `json:"name" db:"name"`
	Domain                 string        `json:"domain" db:"domain"`
	Status                 ProjectStatus `json:"status" db:"status"`
	CPULimit               float64       `json:"cpu_limit" db:"cpu_limit"`
	MemoryLimitMB          int           `json:"memory_limit_mb" db:"memory_limit_mb"`
	WordPressContainerName string        `json:"wordpress_container_name,omitempty" db:"wordpress_container_name"`
	MySQLContainerName     string        `json:"mysql_container_name,omitempty" db:"mysql_container_name"`
	DBName                 string        `json:"-" db:"db_name"`
	DBUser                 string        `json:"-" db:"db_user"`
	DBPassword             string        `json:"-" db:"db_password"`
	CreatedAt              time.Time     `json:"created_at" db:"created_at"`
}

type BackupType string

const (
	BackupTypeManual    BackupType = "manual"
	BackupTypeScheduled BackupType = "scheduled"
)

type Backup struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ProjectID uuid.UUID  `json:"project_id" db:"project_id"`
	FilePath  string     `json:"file_path" db:"file_path"`
	SizeBytes int64      `json:"size_bytes" db:"size_bytes"`
	Type      BackupType `json:"type" db:"type"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type SSLStatus string

const (
	SSLStatusPending SSLStatus = "pending"
	SSLStatusActive  SSLStatus = "active"
	SSLStatusFailed  SSLStatus = "failed"
)

type Domain struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	ProjectID           uuid.UUID `json:"project_id" db:"project_id"`
	DomainName          string    `json:"domain_name" db:"domain_name"`
	SSLStatus           SSLStatus `json:"ssl_status" db:"ssl_status"`
	CloudflareAPIToken  string    `json:"-" db:"cloudflare_api_token"`
	CloudflareEmail     string    `json:"-" db:"cloudflare_email"`
	DNSChallengeEnabled bool      `json:"dns_challenge_enabled" db:"dns_challenge_enabled"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

type EnvironmentVariable struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type BackupSchedule struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ProjectID     uuid.UUID `json:"project_id" db:"project_id"`
	Enabled       bool      `json:"enabled" db:"enabled"`
	ScheduleCron  string    `json:"schedule_cron" db:"schedule_cron"`
	RetentionDays int       `json:"retention_days" db:"retention_days"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type ContainerStats struct {
	ContainerID   string  `json:"container_id"`
	ContainerName string  `json:"container_name"`
	CPUPercentage float64 `json:"cpu_percentage"`
	MemoryUsage   int64   `json:"memory_usage"`
	MemoryLimit   int64   `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     int64   `json:"network_rx"`
	NetworkTx     int64   `json:"network_tx"`
	BlockRead     int64   `json:"block_read"`
	BlockWrite    int64   `json:"block_write"`
}

type ProjectStats struct {
	ProjectID      uuid.UUID       `json:"project_id"`
	WordPressStats *ContainerStats `json:"wordpress_stats,omitempty"`
	MySQLStats     *ContainerStats `json:"mysql_stats,omitempty"`
	DiskUsage      int64           `json:"disk_usage_bytes"`
}
