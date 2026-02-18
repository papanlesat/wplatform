'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import Link from 'next/link';
import { projectApi, backupApi, monitorApi } from '@/lib/api';
import { useAuthStore } from '@/lib/auth';
import { Project, Backup, ProjectStats } from '@/types';

export default function ProjectDetailPage() {
  const router = useRouter();
  const params = useParams();
  const projectId = params.id as string;
  const { token, isLoading: authLoading, checkAuth } = useAuthStore();

  const [project, setProject] = useState<Project | null>(null);
  const [backups, setBackups] = useState<Backup[]>([]);
  const [stats, setStats] = useState<ProjectStats | null>(null);
  const [logs, setLogs] = useState('');
  const [activeTab, setActiveTab] = useState('overview');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  useEffect(() => {
    if (!authLoading && !token) {
      router.push('/login');
    }
  }, [authLoading, token, router]);

  useEffect(() => {
    if (token && projectId) {
      loadProject();
    }
  }, [token, projectId]);

  const loadProject = async () => {
    try {
      const [projectData, backupsData, statsData] = await Promise.all([
        projectApi.get(projectId),
        backupApi.list(projectId),
        monitorApi.getStats(projectId).catch(() => null),
      ]);
      setProject(projectData);
      setBackups(backupsData);
      setStats(statsData);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load project');
    } finally {
      setIsLoading(false);
    }
  };

  const loadLogs = async () => {
    try {
      const logsData = await projectApi.logs(projectId, 'wordpress', '100');
      setLogs(logsData);
    } catch (err: any) {
      setLogs('Failed to load logs');
    }
  };

  const handleStart = async () => {
    try {
      await projectApi.start(projectId);
      loadProject();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to start project');
    }
  };

  const handleStop = async () => {
    try {
      await projectApi.stop(projectId);
      loadProject();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to stop project');
    }
  };

  const handleRestart = async () => {
    try {
      await projectApi.restart(projectId);
      loadProject();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to restart project');
    }
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this project? This action cannot be undone.')) return;
    try {
      await projectApi.delete(projectId);
      router.push('/dashboard');
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to delete project');
    }
  };

  const handleCreateBackup = async () => {
    try {
      await backupApi.create(projectId);
      loadProject();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to create backup');
    }
  };

  const handleRestore = async (backupId: string) => {
    if (!confirm('Are you sure you want to restore from this backup?')) return;
    try {
      await backupApi.restore(backupId);
      loadProject();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to restore backup');
    }
  };

  const handleDeleteBackup = async (backupId: string) => {
    if (!confirm('Are you sure you want to delete this backup?')) return;
    try {
      await backupApi.delete(backupId);
      loadProject();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to delete backup');
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'bg-green-100 text-green-800';
      case 'stopped': return 'bg-gray-100 text-gray-800';
      case 'creating': return 'bg-yellow-100 text-yellow-800';
      case 'error': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (authLoading || isLoading) {
    return <div className="min-h-screen flex items-center justify-center">Loading...</div>;
  }

  if (!project) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-red-500">{error || 'Project not found'}</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <Link href="/dashboard" className="text-xl font-semibold text-gray-900">
                WP Platform
              </Link>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="flex justify-between items-start mb-6">
            <div>
              <h2 className="text-2xl font-bold text-gray-900">{project.name}</h2>
              <p className="text-gray-500">{project.domain}</p>
            </div>
            <div className="flex items-center space-x-2">
              <span className={`px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(project.status)}`}>
                {project.status}
              </span>
            </div>
          </div>

          <div className="flex space-x-4 mb-6 border-b">
            {['overview', 'logs', 'backups', 'stats'].map((tab) => (
              <button
                key={tab}
                onClick={() => {
                  setActiveTab(tab);
                  if (tab === 'logs') loadLogs();
                }}
                className={`px-4 py-2 text-sm font-medium ${activeTab === tab ? 'border-b-2 border-blue-500 text-blue-600' : 'text-gray-500'}`}
              >
                {tab.charAt(0).toUpperCase() + tab.slice(1)}
              </button>
            ))}
          </div>

          <div className="bg-white shadow rounded-lg p-6">
            {activeTab === 'overview' && (
              <div className="space-y-6">
                <div className="grid grid-cols-2 gap-6">
                  <div>
                    <h3 className="text-sm font-medium text-gray-500">CPU Limit</h3>
                    <p className="mt-1 text-lg">{project.cpu_limit} cores</p>
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-gray-500">Memory Limit</h3>
                    <p className="mt-1 text-lg">{project.memory_limit_mb} MB</p>
                  </div>
                </div>

                <div className="flex space-x-3 pt-4">
                  {project.status === 'running' && (
                    <>
                      <button onClick={handleRestart} className="px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-700">
                        Restart
                      </button>
                      <button onClick={handleStop} className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700">
                        Stop
                      </button>
                    </>
                  )}
                  {(project.status === 'stopped' || project.status === 'error') && (
                    <button onClick={handleStart} className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700">
                      Start
                    </button>
                  )}
                  <button onClick={handleDelete} className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700">
                    Delete Project
                  </button>
                </div>
              </div>
            )}

            {activeTab === 'logs' && (
              <div>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-auto max-h-96 text-sm">
                  {logs || 'Click the Logs tab to load logs'}
                </pre>
              </div>
            )}

            {activeTab === 'backups' && (
              <div className="space-y-4">
                <button onClick={handleCreateBackup} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
                  Create Backup
                </button>

                {backups.length === 0 ? (
                  <p className="text-gray-500 py-4">No backups yet</p>
                ) : (
                  <div className="space-y-2">
                    {backups.map((backup) => (
                      <div key={backup.id} className="flex justify-between items-center p-4 bg-gray-50 rounded">
                        <div>
                          <p className="font-medium">{new Date(backup.created_at).toLocaleString()}</p>
                          <p className="text-sm text-gray-500">{formatBytes(backup.size_bytes)} - {backup.type}</p>
                        </div>
                        <div className="space-x-2">
                          <button onClick={() => handleRestore(backup.id)} className="px-3 py-1 bg-green-600 text-white rounded text-sm">
                            Restore
                          </button>
                          <button onClick={() => handleDeleteBackup(backup.id)} className="px-3 py-1 bg-red-600 text-white rounded text-sm">
                            Delete
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {activeTab === 'stats' && stats && (
              <div className="space-y-6">
                <h3 className="font-medium">Container Stats</h3>
                <div className="grid grid-cols-2 gap-6">
                  {stats.wordpress_stats && (
                    <div className="bg-gray-50 p-4 rounded">
                      <h4 className="font-medium mb-2">WordPress</h4>
                      <p className="text-sm">CPU: {stats.wordpress_stats.cpu_percentage.toFixed(2)}%</p>
                      <p className="text-sm">Memory: {formatBytes(stats.wordpress_stats.memory_usage)} / {formatBytes(stats.wordpress_stats.memory_limit)}</p>
                      <p className="text-sm">Network: ↓ {formatBytes(stats.wordpress_stats.network_rx)} / ↑ {formatBytes(stats.wordpress_stats.network_tx)}</p>
                    </div>
                  )}
                  {stats.mysql_stats && (
                    <div className="bg-gray-50 p-4 rounded">
                      <h4 className="font-medium mb-2">MySQL</h4>
                      <p className="text-sm">CPU: {stats.mysql_stats.cpu_percentage.toFixed(2)}%</p>
                      <p className="text-sm">Memory: {formatBytes(stats.mysql_stats.memory_usage)} / {formatBytes(stats.mysql_stats.memory_limit)}</p>
                      <p className="text-sm">Network: ↓ {formatBytes(stats.mysql_stats.network_rx)} / ↑ {formatBytes(stats.mysql_stats.network_tx)}</p>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
