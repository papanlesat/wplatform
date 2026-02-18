'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import Link from 'next/link';
import { PieChart, Pie, Cell, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, Legend } from 'recharts';
import { projectApi, backupApi, monitorApi, envVarApi } from '@/lib/api';
import { useAuthStore } from '@/lib/auth';
import { Project, Backup, ProjectStats, EnvironmentVariable } from '@/types';

export default function ProjectDetailPage() {
  const router = useRouter();
  const params = useParams();
  const projectId = params.id as string;
  const { token, isLoading: authLoading, checkAuth } = useAuthStore();

  const [project, setProject] = useState<Project | null>(null);
  const [backups, setBackups] = useState<Backup[]>([]);
  const [stats, setStats] = useState<ProjectStats | null>(null);
  const [logs, setLogs] = useState('');
  const [envVars, setEnvVars] = useState<EnvironmentVariable[]>([]);
  const [activeTab, setActiveTab] = useState('overview');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [newEnvKey, setNewEnvKey] = useState('');
  const [newEnvValue, setNewEnvValue] = useState('');
  const [editingEnv, setEditingEnv] = useState<EnvironmentVariable | null>(null);
  const [showAddEnv, setShowAddEnv] = useState(false);

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
      const [projectData, backupsData, statsData, envVarsData] = await Promise.all([
        projectApi.get(projectId),
        backupApi.list(projectId),
        monitorApi.getStats(projectId).catch(() => null),
        envVarApi.list(projectId).catch(() => []),
      ]);
      setProject(projectData);
      setBackups(backupsData);
      setStats(statsData);
      setEnvVars(envVarsData);
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

  const loadEnvVars = async () => {
    try {
      const data = await envVarApi.list(projectId);
      setEnvVars(data);
    } catch (err: any) {
      console.error('Failed to load env vars:', err);
    }
  };

  const handleAddEnvVar = async () => {
    if (!newEnvKey.trim() || !newEnvValue.trim()) {
      alert('Please enter both key and value');
      return;
    }
    try {
      await envVarApi.create(projectId, newEnvKey, newEnvValue);
      setNewEnvKey('');
      setNewEnvValue('');
      setShowAddEnv(false);
      loadEnvVars();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to add environment variable');
    }
  };

  const handleUpdateEnvVar = async () => {
    if (!editingEnv) return;
    try {
      await envVarApi.update(editingEnv.id, editingEnv.key, editingEnv.value);
      setEditingEnv(null);
      loadEnvVars();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to update environment variable');
    }
  };

  const handleDeleteEnvVar = async (id: string) => {
    if (!confirm('Are you sure you want to delete this environment variable?')) return;
    try {
      await envVarApi.delete(id);
      loadEnvVars();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to delete environment variable');
    }
  };

  const handleGenerateSalts = async () => {
    if (!confirm('This will regenerate all WordPress salts and may log out all users. Continue?')) return;
    try {
      await envVarApi.generateSalts(projectId);
      loadEnvVars();
      alert('WordPress salts regenerated successfully');
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to generate salts');
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
            {['overview', 'logs', 'backups', 'stats', 'settings'].map((tab) => (
              <button
                key={tab}
                onClick={() => {
                  setActiveTab(tab);
                  if (tab === 'logs') loadLogs();
                  if (tab === 'settings') loadEnvVars();
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
                
                <div className="bg-gray-50 p-4 rounded">
                  <h4 className="font-medium mb-4">CPU Usage</h4>
                  <div className="h-48">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={[
                        { name: 'WordPress', cpu: stats.wordpress_stats?.cpu_percentage || 0 },
                        { name: 'MySQL', cpu: stats.mysql_stats?.cpu_percentage || 0 },
                      ]}>
                        <XAxis dataKey="name" />
                        <YAxis unit="%" domain={[0, 100]} />
                        <Tooltip formatter={(value: number) => `${value.toFixed(2)}%`} />
                        <Bar dataKey="cpu" fill="#3b82f6" name="CPU %" />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-6">
                  {stats.wordpress_stats && (
                    <div className="bg-gray-50 p-4 rounded">
                      <h4 className="font-medium mb-2">WordPress Memory</h4>
                      <div className="h-40">
                        <ResponsiveContainer width="100%" height="100%">
                          <PieChart>
                            <Pie
                              data={[
                                { name: 'Used', value: stats.wordpress_stats.memory_usage },
                                { name: 'Free', value: stats.wordpress_stats.memory_limit - stats.wordpress_stats.memory_usage },
                              ]}
                              cx="50%"
                              cy="50%"
                              innerRadius={40}
                              outerRadius={60}
                              dataKey="value"
                              label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                            >
                              <Cell fill="#ef4444" />
                              <Cell fill="#e5e7eb" />
                            </Pie>
                            <Tooltip formatter={(value: number) => formatBytes(value)} />
                          </PieChart>
                        </ResponsiveContainer>
                      </div>
                      <p className="text-sm text-center text-gray-500 mt-2">
                        {formatBytes(stats.wordpress_stats.memory_usage)} / {formatBytes(stats.wordpress_stats.memory_limit)}
                      </p>
                    </div>
                  )}
                  {stats.mysql_stats && (
                    <div className="bg-gray-50 p-4 rounded">
                      <h4 className="font-medium mb-2">MySQL Memory</h4>
                      <div className="h-40">
                        <ResponsiveContainer width="100%" height="100%">
                          <PieChart>
                            <Pie
                              data={[
                                { name: 'Used', value: stats.mysql_stats.memory_usage },
                                { name: 'Free', value: stats.mysql_stats.memory_limit - stats.mysql_stats.memory_usage },
                              ]}
                              cx="50%"
                              cy="50%"
                              innerRadius={40}
                              outerRadius={60}
                              dataKey="value"
                              label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                            >
                              <Cell fill="#f59e0b" />
                              <Cell fill="#e5e7eb" />
                            </Pie>
                            <Tooltip formatter={(value: number) => formatBytes(value)} />
                          </PieChart>
                        </ResponsiveContainer>
                      </div>
                      <p className="text-sm text-center text-gray-500 mt-2">
                        {formatBytes(stats.mysql_stats.memory_usage)} / {formatBytes(stats.mysql_stats.memory_limit)}
                      </p>
                    </div>
                  )}
                </div>

                <div className="bg-gray-50 p-4 rounded">
                  <h4 className="font-medium mb-4">Network I/O</h4>
                  <div className="h-48">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={[
                        { 
                          name: 'WordPress', 
                          'Download': stats.wordpress_stats?.network_rx || 0,
                          'Upload': stats.wordpress_stats?.network_tx || 0,
                        },
                        { 
                          name: 'MySQL', 
                          'Download': stats.mysql_stats?.network_rx || 0,
                          'Upload': stats.mysql_stats?.network_tx || 0,
                        },
                      ]}>
                        <XAxis dataKey="name" />
                        <YAxis tickFormatter={(value) => formatBytes(value)} />
                        <Tooltip formatter={(value: number) => formatBytes(value)} />
                        <Legend />
                        <Bar dataKey="Download" fill="#22c55e" />
                        <Bar dataKey="Upload" fill="#3b82f6" />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </div>

                {stats.disk_usage && (
                  <div className="bg-gray-50 p-4 rounded">
                    <h4 className="font-medium mb-2">Disk Usage</h4>
                    <div className="h-40">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={[
                              { name: 'Used', value: stats.disk_usage.usage_bytes },
                              { name: 'Free', value: stats.disk_usage.total_bytes - stats.disk_usage.usage_bytes },
                            ]}
                            cx="50%"
                            cy="50%"
                            innerRadius={50}
                            outerRadius={70}
                            dataKey="value"
                            label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                          >
                            <Cell fill="#8b5cf6" />
                            <Cell fill="#e5e7eb" />
                          </Pie>
                          <Tooltip formatter={(value: number) => formatBytes(value)} />
                        </PieChart>
                      </ResponsiveContainer>
                    </div>
                    <p className="text-sm text-center text-gray-500 mt-2">
                      {formatBytes(stats.disk_usage.usage_bytes)} / {formatBytes(stats.disk_usage.total_bytes)}
                    </p>
                  </div>
                )}
              </div>
            )}

            {activeTab === 'stats' && !stats && (
              <div className="text-center py-8 text-gray-500">
                No stats available. Make sure the project is running.
              </div>
            )}

            {activeTab === 'settings' && (
              <div className="space-y-6">
                <div className="flex justify-between items-center">
                  <h3 className="font-medium">Environment Variables</h3>
                  <div className="space-x-2">
                    <button onClick={handleGenerateSalts} className="px-3 py-1 bg-purple-600 text-white rounded text-sm hover:bg-purple-700">
                      Generate WP Salts
                    </button>
                    <button onClick={() => setShowAddEnv(true)} className="px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700">
                      Add Variable
                    </button>
                  </div>
                </div>

                {showAddEnv && (
                  <div className="bg-gray-50 p-4 rounded space-y-3">
                    <div className="grid grid-cols-2 gap-4">
                      <input
                        type="text"
                        placeholder="Key (e.g., WP_MEMORY_LIMIT)"
                        value={newEnvKey}
                        onChange={(e) => setNewEnvKey(e.target.value)}
                        className="px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      <input
                        type="text"
                        placeholder="Value (e.g., 256M)"
                        value={newEnvValue}
                        onChange={(e) => setNewEnvValue(e.target.value)}
                        className="px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                    <div className="flex space-x-2">
                      <button onClick={handleAddEnvVar} className="px-3 py-1 bg-green-600 text-white rounded text-sm hover:bg-green-700">
                        Save
                      </button>
                      <button onClick={() => { setShowAddEnv(false); setNewEnvKey(''); setNewEnvValue(''); }} className="px-3 py-1 bg-gray-300 text-gray-700 rounded text-sm hover:bg-gray-400">
                        Cancel
                      </button>
                    </div>
                  </div>
                )}

                {envVars.length === 0 ? (
                  <p className="text-gray-500 py-4">No environment variables configured</p>
                ) : (
                  <div className="space-y-2">
                    {envVars.map((envVar) => (
                      <div key={envVar.id} className="flex justify-between items-center p-4 bg-gray-50 rounded">
                        {editingEnv?.id === envVar.id ? (
                          <div className="flex-1 grid grid-cols-2 gap-4">
                            <input
                              type="text"
                              value={editingEnv.key}
                              onChange={(e) => setEditingEnv({ ...editingEnv, key: e.target.value })}
                              className="px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <input
                              type="text"
                              value={editingEnv.value}
                              onChange={(e) => setEditingEnv({ ...editingEnv, value: e.target.value })}
                              className="px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                          </div>
                        ) : (
                          <div className="flex-1">
                            <p className="font-mono text-sm font-medium text-gray-900">{envVar.key}</p>
                            <p className="font-mono text-sm text-gray-500 truncate max-w-md">{envVar.value}</p>
                          </div>
                        )}
                        <div className="space-x-2">
                          {editingEnv?.id === envVar.id ? (
                            <>
                              <button onClick={handleUpdateEnvVar} className="px-3 py-1 bg-green-600 text-white rounded text-sm">
                                Save
                              </button>
                              <button onClick={() => setEditingEnv(null)} className="px-3 py-1 bg-gray-300 text-gray-700 rounded text-sm">
                                Cancel
                              </button>
                            </>
                          ) : (
                            <>
                              <button onClick={() => setEditingEnv(envVar)} className="px-3 py-1 bg-yellow-600 text-white rounded text-sm hover:bg-yellow-700">
                                Edit
                              </button>
                              <button onClick={() => handleDeleteEnvVar(envVar.id)} className="px-3 py-1 bg-red-600 text-white rounded text-sm hover:bg-red-700">
                                Delete
                              </button>
                            </>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
