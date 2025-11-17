import axios, { AxiosInstance } from 'axios';
import * as vscode from 'vscode';

let apiClient: AxiosInstance | null = null;

export function initializeApiClient(context: vscode.ExtensionContext): AxiosInstance {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:3000';

  apiClient = axios.create({
    baseURL: `${backendUrl}/api/v1`,
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Add auth token interceptor
  apiClient.interceptors.request.use(async (config) => {
    // Get token from VSCode secrets
    const token = await context.secrets.get('accessToken');

    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    return config;
  });

  // Handle errors
  apiClient.interceptors.response.use(
    (response) => response,
    async (error) => {
      if (error.response?.status === 401) {
        vscode.window.showErrorMessage('Authentication failed. Please sign in again.');
        // TODO: Trigger sign-in flow
      } else if (error.response?.status === 403) {
        vscode.window.showWarningMessage(
          error.response.data?.error?.message || 'Access forbidden. You may have reached your quota limit.',
        );
      }

      return Promise.reject(error);
    },
  );

  return apiClient;
}

export function getApiClient(): AxiosInstance {
  if (!apiClient) {
    throw new Error('API client not initialized. Call initializeApiClient first.');
  }

  return apiClient;
}
