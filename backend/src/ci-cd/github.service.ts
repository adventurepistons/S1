import { Injectable, Logger } from '@nestjs/common';
import axios, { AxiosInstance } from 'axios';

export interface GithubStatusCheckParams {
  repo: string; // e.g., "owner/repo"
  sha: string;
  state: 'pending' | 'success' | 'failure' | 'error';
  context: string;
  description: string;
  target_url?: string;
}

export interface GithubCommentParams {
  repo: string;
  issue_number: number;
  body: string;
}

export interface GithubCheckRunParams {
  repo: string;
  name: string;
  head_sha: string;
  status: 'queued' | 'in_progress' | 'completed';
  conclusion?: 'success' | 'failure' | 'neutral' | 'cancelled' | 'skipped' | 'timed_out';
  output?: {
    title: string;
    summary: string;
    text?: string;
  };
}

@Injectable()
export class GithubService {
  private readonly logger = new Logger(GithubService.name);
  private client: AxiosInstance;

  constructor() {
    // Initialize GitHub API client
    this.client = axios.create({
      baseURL: 'https://api.github.com',
      headers: {
        Accept: 'application/vnd.github.v3+json',
        'User-Agent': 'TestCopilot-Bot',
      },
      timeout: 30000,
    });

    // Add auth token from environment
    const token = process.env.GITHUB_TOKEN;
    if (token) {
      this.client.defaults.headers.common['Authorization'] = `token ${token}`;
    }
  }

  /**
   * Create a commit status check
   * Legacy status API - works for all GitHub apps
   */
  async createStatusCheck(params: GithubStatusCheckParams): Promise<void> {
    try {
      const [owner, repo] = params.repo.split('/');

      await this.client.post(`/repos/${owner}/${repo}/statuses/${params.sha}`, {
        state: params.state,
        target_url: params.target_url,
        description: params.description,
        context: params.context,
      });

      this.logger.log(`GitHub status created: ${params.state} for ${params.sha}`);
    } catch (error: any) {
      this.logger.error(
        `Failed to create GitHub status: ${error.response?.data?.message || error.message}`,
      );
      throw error;
    }
  }

  /**
   * Create a check run
   * Modern Checks API - requires GitHub App installation
   */
  async createCheckRun(params: GithubCheckRunParams): Promise<void> {
    try {
      const [owner, repo] = params.repo.split('/');

      await this.client.post(`/repos/${owner}/${repo}/check-runs`, {
        name: params.name,
        head_sha: params.head_sha,
        status: params.status,
        conclusion: params.conclusion,
        output: params.output,
      });

      this.logger.log(`GitHub check run created: ${params.status} for ${params.head_sha}`);
    } catch (error: any) {
      this.logger.error(
        `Failed to create check run: ${error.response?.data?.message || error.message}`,
      );
      // Don't throw - fallback to status checks
    }
  }

  /**
   * Create a comment on PR or issue
   */
  async createComment(params: GithubCommentParams): Promise<void> {
    try {
      const [owner, repo] = params.repo.split('/');

      await this.client.post(`/repos/${owner}/${repo}/issues/${params.issue_number}/comments`, {
        body: params.body,
      });

      this.logger.log(`GitHub comment created on #${params.issue_number}`);
    } catch (error: any) {
      this.logger.error(
        `Failed to create comment: ${error.response?.data?.message || error.message}`,
      );
      throw error;
    }
  }

  /**
   * Update an existing comment
   */
  async updateComment(repo: string, commentId: number, body: string): Promise<void> {
    try {
      const [owner, repoName] = repo.split('/');

      await this.client.patch(`/repos/${owner}/${repoName}/issues/comments/${commentId}`, {
        body,
      });

      this.logger.log(`GitHub comment ${commentId} updated`);
    } catch (error: any) {
      this.logger.error(
        `Failed to update comment: ${error.response?.data?.message || error.message}`,
      );
      throw error;
    }
  }

  /**
   * Get PR files changed
   */
  async getPRFiles(repo: string, prNumber: number): Promise<string[]> {
    try {
      const [owner, repoName] = repo.split('/');

      const response = await this.client.get(`/repos/${owner}/${repoName}/pulls/${prNumber}/files`);

      return response.data.map((file: any) => file.filename);
    } catch (error: any) {
      this.logger.error(`Failed to get PR files: ${error.response?.data?.message || error.message}`);
      return [];
    }
  }

  /**
   * Get repository details
   */
  async getRepository(repo: string): Promise<any> {
    try {
      const [owner, repoName] = repo.split('/');

      const response = await this.client.get(`/repos/${owner}/${repoName}`);

      return response.data;
    } catch (error: any) {
      this.logger.error(
        `Failed to get repository: ${error.response?.data?.message || error.message}`,
      );
      throw error;
    }
  }

  /**
   * Get pull request details
   */
  async getPullRequest(repo: string, prNumber: number): Promise<any> {
    try {
      const [owner, repoName] = repo.split('/');

      const response = await this.client.get(`/repos/${owner}/${repoName}/pulls/${prNumber}`);

      return response.data;
    } catch (error: any) {
      this.logger.error(`Failed to get PR: ${error.response?.data?.message || error.message}`);
      throw error;
    }
  }

  /**
   * Create a deployment status
   */
  async createDeploymentStatus(
    repo: string,
    deploymentId: number,
    state: 'error' | 'failure' | 'pending' | 'success',
    description?: string,
  ): Promise<void> {
    try {
      const [owner, repoName] = repo.split('/');

      await this.client.post(
        `/repos/${owner}/${repoName}/deployments/${deploymentId}/statuses`,
        {
          state,
          description,
        },
      );

      this.logger.log(`Deployment status created: ${state}`);
    } catch (error: any) {
      this.logger.error(
        `Failed to create deployment status: ${error.response?.data?.message || error.message}`,
      );
    }
  }
}
