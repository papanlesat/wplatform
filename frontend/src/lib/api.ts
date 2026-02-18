import axios from 'axios';
import { AuthResponse, Project, CreateProjectRequest, Backup, BackupSchedule, ProjectStats, EnvironmentVariable } from '@/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export const authApi = {
  login: async (email: string, password: string): Promise<AuthResponse> => {
    const response = await api.post('/auth/login', { email, password, role: 'user' });
    return response.data;
  },

  register: async (email: string, password: string): Promise<AuthResponse> => {
    const response = await api.post('/auth/register', { email, password, role: 'user' });
    return response.data;
  },
};

export const projectApi = {
  list: async (): Promise<Project[]> => {
    const response = await api.get('/api/projects');
    return response.data || [];
  },

  get: async (id: string): Promise<Project> => {
    const response = await api.get(`/api/projects/${id}`);
    return response.data;
  },

  create: async (data: CreateProjectRequest): Promise<Project> => {
    const response = await api.post('/api/projects', data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/projects/${id}`);
  },

  start: async (id: string): Promise<void> => {
    await api.post(`/api/projects/${id}/start`);
  },

  stop: async (id: string): Promise<void> => {
    await api.post(`/api/projects/${id}/stop`);
  },

  restart: async (id: string): Promise<void> => {
    await api.post(`/api/projects/${id}/restart`);
  },

  health: async (id: string): Promise<{ status: string }> => {
    const response = await api.get(`/api/projects/${id}/health`);
    return response.data;
  },

  logs: async (id: string, container?: string, tail?: string): Promise<string> => {
    const params = new URLSearchParams();
    if (container) params.append('container', container);
    if (tail) params.append('tail', tail);
    const response = await api.get(`/api/projects/${id}/logs?${params.toString()}`, {
      responseType: 'text',
    });
    return response.data;
  },
};

export const backupApi = {
  list: async (projectId: string): Promise<Backup[]> => {
    const response = await api.get(`/api/projects/${projectId}/backups`);
    return response.data || [];
  },

  create: async (projectId: string): Promise<Backup> => {
    const response = await api.post('/api/backups', { project_id: projectId, type: 'manual' });
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/backups/${id}`);
  },

  download: async (id: string): Promise<Blob> => {
    const response = await api.get(`/api/backups/${id}/download`, {
      responseType: 'blob',
    });
    return response.data;
  },

  restore: async (id: string): Promise<void> => {
    await api.post(`/api/backups/${id}/restore`);
  },

  getSchedule: async (projectId: string): Promise<BackupSchedule> => {
    const response = await api.get(`/api/projects/${projectId}/schedule`);
    return response.data;
  },

  updateSchedule: async (projectId: string, schedule: Partial<BackupSchedule>): Promise<BackupSchedule> => {
    const response = await api.put(`/api/projects/${projectId}/schedule`, schedule);
    return response.data;
  },
};

export const monitorApi = {
  getStats: async (projectId: string): Promise<ProjectStats> => {
    const response = await api.get(`/api/projects/${projectId}/stats`);
    return response.data;
  },
};

export const dnsApi = {
  checkDNS: async (domain: string): Promise<{
    domain: string;
    server_ip: string;
    resolved_ips: string[];
    propagated: boolean;
    message: string;
  }> => {
    const response = await api.post('/api/dns/check', { domain });
    return response.data;
  },
  checkProjectDNS: async (projectId: string, domain: string): Promise<{
    domain: string;
    server_ip: string;
    resolved_ips: string[];
    propagated: boolean;
    message: string;
  }> => {
    const response = await api.get(`/api/projects/${projectId}/dns/check?domain=${encodeURIComponent(domain)}`);
    return response.data;
  },
};

export const envVarApi = {
  list: async (projectId: string): Promise<EnvironmentVariable[]> => {
    const response = await api.get(`/api/projects/${projectId}/envvars`);
    return response.data || [];
  },

  create: async (projectId: string, key: string, value: string): Promise<EnvironmentVariable> => {
    const response = await api.post(`/api/projects/${projectId}/envvars`, { key, value });
    return response.data;
  },

  update: async (id: string, key: string, value: string): Promise<EnvironmentVariable> => {
    const response = await api.put(`/api/envvars/${id}`, { key, value });
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/envvars/${id}`);
  },

  generateSalts: async (projectId: string): Promise<Record<string, string>> => {
    const response = await api.post(`/api/projects/${projectId}/envvars/generate-salts`);
    return response.data;
  },
};

export default api;
