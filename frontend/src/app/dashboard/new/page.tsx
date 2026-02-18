'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { projectApi } from '@/lib/api';
import { useAuthStore } from '@/lib/auth';
import { CreateProjectRequest } from '@/types';

export default function NewProjectPage() {
  const router = useRouter();
  const { token, isLoading: authLoading, checkAuth } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [formData, setFormData] = useState<CreateProjectRequest>({
    name: '',
    domain: '',
    cpu_limit: 0.5,
    memory_limit_mb: 512,
  });

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  useEffect(() => {
    if (!authLoading && !token) {
      router.push('/login');
    }
  }, [authLoading, token, router]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: name === 'cpu_limit' ? parseFloat(value) : name === 'memory_limit_mb' ? parseInt(value) : value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      const project = await projectApi.create(formData);
      router.push(`/dashboard/${project.id}`);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create project');
      setIsLoading(false);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  if (!token) {
    return null;
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

      <main className="max-w-3xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">Create New Project</h2>

          {error && (
            <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded mb-6">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="bg-white shadow rounded-lg p-6 space-y-6">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                Project Name
              </label>
              <input
                type="text"
                id="name"
                name="name"
                required
                value={formData.name}
                onChange={handleChange}
                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                placeholder="My WordPress Site"
              />
            </div>

            <div>
              <label htmlFor="domain" className="block text-sm font-medium text-gray-700">
                Domain
              </label>
              <input
                type="text"
                id="domain"
                name="domain"
                required
                value={formData.domain}
                onChange={handleChange}
                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                placeholder="example.com"
              />
              <p className="mt-1 text-sm text-gray-500">
                Make sure the domain DNS A record points to this server&apos;s IP address.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-6">
              <div>
                <label htmlFor="cpu_limit" className="block text-sm font-medium text-gray-700">
                  CPU Limit (cores)
                </label>
                <input
                  type="number"
                  id="cpu_limit"
                  name="cpu_limit"
                  min="0.1"
                  max="4"
                  step="0.1"
                  required
                  value={formData.cpu_limit}
                  onChange={handleChange}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                />
              </div>

              <div>
                <label htmlFor="memory_limit_mb" className="block text-sm font-medium text-gray-700">
                  Memory Limit (MB)
                </label>
                <input
                  type="number"
                  id="memory_limit_mb"
                  name="memory_limit_mb"
                  min="128"
                  max="4096"
                  step="128"
                  required
                  value={formData.memory_limit_mb}
                  onChange={handleChange}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                />
              </div>
            </div>

            <div className="flex justify-end space-x-3 pt-4">
              <Link
                href="/dashboard"
                className="px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
              >
                Cancel
              </Link>
              <button
                type="submit"
                disabled={isLoading}
                className="px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Creating...' : 'Create Project'}
              </button>
            </div>
          </form>
        </div>
      </main>
    </div>
  );
}
