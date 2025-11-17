import { Injectable, Logger, BadRequestException } from '@nestjs/common';
import { DatabaseService } from '../shared/database.service';
import { TestRunnerService } from '../tests/services/test-runner.service';
import { GithubService } from './github.service';
import { NotificationService } from './notification.service';
import * as crypto from 'crypto';

export interface WebhookPayload {
  event: string;
  action: string;
  repository: {
    full_name: string;
    clone_url: string;
  };
  pull_request?: {
    number: number;
    title: string;
    head: {
      ref: string;
      sha: string;
    };
    base: {
      ref: string;
    };
  };
  sender: {
    login: string;
  };
}

@Injectable()
export class WebhookService {
  private readonly logger = new Logger(WebhookService.name);

  constructor(
    private db: DatabaseService,
    private testRunner: TestRunnerService,
    private github: GithubService,
    private notifications: NotificationService,
  ) {}

  /**
   * Handle GitHub webhook
   */
  async handleGithubWebhook(
    payload: WebhookPayload,
    signature: string,
    secret: string,
  ): Promise<any> {
    // Verify signature
    if (!this.verifyGithubSignature(JSON.stringify(payload), signature, secret)) {
      throw new BadRequestException('Invalid signature');
    }

    this.logger.log(
      `GitHub webhook received: ${payload.event} - ${payload.action} - ${payload.repository.full_name}`,
    );

    // Handle pull request events
    if (payload.event === 'pull_request' && payload.pull_request) {
      return this.handlePullRequest(payload);
    }

    // Handle push events
    if (payload.event === 'push') {
      return this.handlePush(payload);
    }

    return { message: 'Event not handled' };
  }

  /**
   * Handle pull request webhook
   */
  private async handlePullRequest(payload: WebhookPayload): Promise<any> {
    const pr = payload.pull_request!;

    if (['opened', 'synchronize', 'reopened'].includes(payload.action)) {
      this.logger.log(`Triggering tests for PR #${pr.number}`);

      // Find project by repository
      const project = await this.db.project.findFirst({
        where: {
          repositoryUrl: payload.repository.clone_url,
        },
      });

      if (!project) {
        this.logger.warn(`No project found for repository: ${payload.repository.full_name}`);
        return { message: 'Project not found' };
      }

      // Trigger test execution
      const execution = await this.testRunner.executeTests(project.userId, {
        projectId: project.id,
        tests: ['*'],
        environment: 'staging',
        browser: 'chrome',
        headless: true,
        parallel: true,
        maxWorkers: 4,
        retryFailedTests: 1,
      });

      // Post status check to GitHub
      await this.github.createStatusCheck({
        repo: payload.repository.full_name,
        sha: pr.head.sha,
        state: 'pending',
        context: 'Test Copilot / Tests',
        description: 'Running tests...',
        target_url: `https://testcopilot.dev/executions/${execution.executionId}`,
      });

      this.logger.log(`Tests triggered for PR #${pr.number}: ${execution.executionId}`);

      return {
        message: 'Tests triggered',
        executionId: execution.executionId,
        pr: pr.number,
      };
    }

    return { message: 'Action not handled' };
  }

  /**
   * Handle push webhook
   */
  private async handlePush(payload: WebhookPayload): Promise<any> {
    // Only run tests on main/master branch
    const ref = (payload as any).ref;
    if (!ref || (!ref.includes('main') && !ref.includes('master'))) {
      return { message: 'Not main branch' };
    }

    this.logger.log(`Triggering tests for push to ${ref}`);

    // Find project
    const project = await this.db.project.findFirst({
      where: {
        repositoryUrl: payload.repository.clone_url,
      },
    });

    if (!project) {
      return { message: 'Project not found' };
    }

    // Trigger tests
    const execution = await this.testRunner.executeTests(project.userId, {
      projectId: project.id,
      tests: ['*'],
      environment: 'production',
      browser: 'chrome',
      headless: true,
      parallel: true,
      maxWorkers: 8,
      retryFailedTests: 2,
    });

    return {
      message: 'Tests triggered',
      executionId: execution.executionId,
    };
  }

  /**
   * Update GitHub status after test completion
   */
  async updateGithubStatus(executionId: string, repo: string, sha: string): Promise<void> {
    const execution = await this.db.testExecution.findUnique({
      where: { id: executionId },
    });

    if (!execution) {
      return;
    }

    let state: 'success' | 'failure' | 'error' = 'success';
    let description = `All tests passed (${execution.passedTests}/${execution.totalTests})`;

    if (execution.status === 'FAILED') {
      state = 'error';
      description = 'Execution failed';
    } else if ((execution.failedTests || 0) > 0) {
      state = 'failure';
      description = `${execution.failedTests} test(s) failed`;
    }

    await this.github.createStatusCheck({
      repo,
      sha,
      state,
      context: 'Test Copilot / Tests',
      description,
      target_url: `https://testcopilot.dev/executions/${executionId}`,
    });

    this.logger.log(`GitHub status updated: ${state}`);
  }

  /**
   * Post test results as PR comment
   */
  async postPRComment(executionId: string, repo: string, prNumber: number): Promise<void> {
    const execution = await this.db.testExecution.findUnique({
      where: { id: executionId },
      include: {
        results: {
          where: {
            status: 'FAILED',
          },
          take: 10,
        },
      },
    });

    if (!execution) {
      return;
    }

    const summary = `
## 🧪 Test Results

**Status**: ${execution.status}
**Duration**: ${execution.duration}s

| Total | Passed | Failed | Skipped |
|-------|--------|--------|---------|
| ${execution.totalTests} | ${execution.passedTests} | ${execution.failedTests} | ${execution.skippedTests} |

${
  execution.failedTests && execution.failedTests > 0
    ? `
### ❌ Failed Tests

${execution.results
  .map(
    (r) => `
**${r.testName}**
\`\`\`
${r.errorMessage}
\`\`\`
${r.screenshotUrl ? `![Screenshot](${r.screenshotUrl})` : ''}
`,
  )
  .join('\n')}
`
    : '✅ All tests passed!'
}

[View Full Report](https://testcopilot.dev/executions/${executionId})
`;

    await this.github.createComment({
      repo,
      issue_number: prNumber,
      body: summary,
    });

    this.logger.log(`PR comment posted to #${prNumber}`);
  }

  /**
   * Verify GitHub webhook signature
   */
  private verifyGithubSignature(payload: string, signature: string, secret: string): boolean {
    const hmac = crypto.createHmac('sha256', secret);
    const digest = 'sha256=' + hmac.update(payload).digest('hex');
    return crypto.timingSafeEqual(Buffer.from(signature), Buffer.from(digest));
  }
}
