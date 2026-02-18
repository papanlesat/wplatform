export interface User {
  id: string;
  email: string;
  role: 'admin' | 'user';
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Project {
  id: string;
  user_id: string;
  name: string;
  domain: string;
  status: 'creating' | 'running' | 'stopped' | 'error';
  cpu_limit: number;
  memory_limit_mb: number;
  wordpress_container_name: string;
  mysql_container_name: string;
  db_name: string;
  db_user: string;
  created_at: string;
}

export interface CreateProjectRequest {
  name: string;
  domain: string;
  cpu_limit?: number;
  memory_limit_mb?: number;
}

export interface Backup {
  id: string;
  project_id: string;
  file_path: string;
  size_bytes: number;
  type: 'manual' | 'scheduled';
  created_at: string;
}

export interface BackupSchedule {
  id: string;
  project_id: string;
  enabled: boolean;
  schedule_cron: string;
  retention_days: number;
  created_at: string;
  updated_at: string;
}

export interface ProjectStats {
  project_id: string;
  wordpress_stats?: ContainerStats;
  mysql_stats?: ContainerStats;
  disk_usage?: DiskUsage;
  collected_at: string;
}

export interface ContainerStats {
  container_id: string;
  container_name: string;
  cpu_percentage: number;
  memory_usage: number;
  memory_limit: number;
  memory_percent: number;
  network_rx: number;
  network_tx: number;
  block_read: number;
  block_write: number;
}

export interface DiskUsage {
  usage_bytes: number;
  total_bytes: number;
  percent_used: number;
}

export interface EnvironmentVariable {
  id: string;
  project_id: string;
  key: string;
  value: string;
  created_at: string;
  updated_at: string;
}

export interface Domain {
  id: string;
  project_id: string;
  domain_name: string;
  ssl_status: 'pending' | 'active' | 'failed';
  cloudflare_api_token?: string;
  cloudflare_email?: string;
  dns_challenge_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface ApiError {
  error: string;
  message?: string;
}
