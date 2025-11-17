import { Injectable, Logger, NotFoundException } from '@nestjs/common';
import { InjectQueue } from '@nestjs/bullmq';
import { Queue } from 'bullmq';
import { DatabaseService } from '../../shared/database.service';
import { CacheService } from '../../shared/cache.service';
import { ExecuteTestsDto } from '../dto/execute-tests.dto';
import { ExecutionStatus, TestStatus } from '@prisma/client';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export interface TestExecutionResult {
  executionId: string;
  status: ExecutionStatus;
  queuePosition?: number;
  estimatedStartTime?: Date;
  websocketUrl: string;
}

export interface TestLog {
  type: 'status' | 'log' | 'test_started' | 'test_completed' | 'screenshot' | 'execution_completed';
  executionId: string;
  testName?: string;
  status?: string;
  message?: string;
  level?: 'info' | 'warn' | 'error' | 'debug';
  duration?: number;
  url?: string;
  timestamp: string;
  summary?: any;
}

@Injectable()
export class TestRunnerService {
  private readonly logger = new Logger(TestRunnerService.name);

  constructor(
    private db: DatabaseService,
    private cacheService: CacheService,
    @InjectQueue('test-execution') private testQueue: Queue,
  ) {}

  /**
   * Execute tests - creates execution record and adds to queue
   */
  async executeTests(
    userId: string,
    dto: ExecuteTestsDto,
  ): Promise<TestExecutionResult> {
    // Verify project exists and user has access
    const project = await this.db.project.findFirst({
      where: {
        id: dto.projectId,
        userId,
      },
    });

    if (!project) {
      throw new NotFoundException('Project not found');
    }

    // Create execution record
    const execution = await this.db.testExecution.create({
      data: {
        projectId: dto.projectId,
        userId,
        status: ExecutionStatus.QUEUED,
        environment: dto.environment,
        browser: dto.browser,
        headless: dto.headless,
        parallel: dto.parallel,
        maxWorkers: dto.maxWorkers,
        retryFailed: dto.retryFailedTests,
      },
    });

    // Add to queue
    await this.testQueue.add(
      'run-tests',
      {
        executionId: execution.id,
        projectId: dto.projectId,
        tests: dto.tests || ['*'],
        environment: dto.environment,
        browser: dto.browser,
        headless: dto.headless,
        parallel: dto.parallel,
        maxWorkers: dto.maxWorkers,
        retryFailed: dto.retryFailedTests,
        timeout: dto.timeout,
        tags: dto.tags,
      },
      {
        attempts: 3,
        backoff: {
          type: 'exponential',
          delay: 2000,
        },
      },
    );

    this.logger.log(`Test execution queued: ${execution.id}`);

    // Get queue position
    const queuePosition = await this.getQueuePosition(execution.id);

    return {
      executionId: execution.id,
      status: ExecutionStatus.QUEUED,
      queuePosition,
      websocketUrl: `ws://localhost:3000/tests/stream/${execution.id}`,
    };
  }

  /**
   * Get execution status
   */
  async getExecutionStatus(executionId: string, userId: string): Promise<any> {
    const execution = await this.db.testExecution.findFirst({
      where: {
        id: executionId,
        userId,
      },
      include: {
        results: {
          select: {
            id: true,
            testName: true,
            status: true,
            duration: true,
            errorMessage: true,
            screenshotUrl: true,
          },
        },
      },
    });

    if (!execution) {
      throw new NotFoundException('Execution not found');
    }

    return {
      executionId: execution.id,
      status: execution.status,
      startedAt: execution.startedAt,
      completedAt: execution.completedAt,
      duration: execution.duration,
      summary: {
        total: execution.totalTests || 0,
        passed: execution.passedTests || 0,
        failed: execution.failedTests || 0,
        skipped: execution.skippedTests || 0,
      },
      tests: execution.results,
    };
  }

  /**
   * Cancel execution
   */
  async cancelExecution(executionId: string, userId: string): Promise<void> {
    const execution = await this.db.testExecution.findFirst({
      where: {
        id: executionId,
        userId,
      },
    });

    if (!execution) {
      throw new NotFoundException('Execution not found');
    }

    // Update status
    await this.db.testExecution.update({
      where: { id: executionId },
      data: { status: ExecutionStatus.CANCELLED },
    });

    // Remove from queue if still queued
    const jobs = await this.testQueue.getJobs(['waiting', 'active']);
    for (const job of jobs) {
      if (job.data.executionId === executionId) {
        await job.remove();
      }
    }

    this.logger.log(`Test execution cancelled: ${executionId}`);
  }

  /**
   * Actually run the tests (called by worker)
   */
  async runTests(executionId: string, data: any): Promise<void> {
    this.logger.log(`Starting test execution: ${executionId}`);

    try {
      // Update status to RUNNING
      await this.db.testExecution.update({
        where: { id: executionId },
        data: {
          status: ExecutionStatus.RUNNING,
          startedAt: new Date(),
        },
      });

      // Emit status update
      await this.emitLog({
        type: 'status',
        executionId,
        status: 'running',
        timestamp: new Date().toISOString(),
      });

      // Get project info
      const execution = await this.db.testExecution.findUnique({
        where: { id: executionId },
        include: { project: true },
      });

      if (!execution) {
        throw new Error('Execution not found');
      }

      // TODO: This is where we'd integrate with Selenium Grid or Playwright
      // For now, we'll simulate test execution
      await this.simulateTestExecution(executionId, data);

      // Update execution with results
      const results = await this.db.testResult.findMany({
        where: { executionId },
      });

      const summary = {
        total: results.length,
        passed: results.filter((r) => r.status === TestStatus.PASSED).length,
        failed: results.filter((r) => r.status === TestStatus.FAILED).length,
        skipped: results.filter((r) => r.status === TestStatus.SKIPPED).length,
      };

      const completedAt = new Date();
      const startedAt = execution.startedAt || new Date();
      const duration = Math.floor((completedAt.getTime() - startedAt.getTime()) / 1000);

      await this.db.testExecution.update({
        where: { id: executionId },
        data: {
          status: ExecutionStatus.COMPLETED,
          completedAt,
          duration,
          totalTests: summary.total,
          passedTests: summary.passed,
          failedTests: summary.failed,
          skippedTests: summary.skipped,
        },
      });

      // Emit completion
      await this.emitLog({
        type: 'execution_completed',
        executionId,
        status: 'completed',
        summary,
        timestamp: new Date().toISOString(),
      });

      this.logger.log(`Test execution completed: ${executionId}`);
    } catch (error) {
      this.logger.error(`Test execution failed: ${executionId}`, error);

      await this.db.testExecution.update({
        where: { id: executionId },
        data: {
          status: ExecutionStatus.FAILED,
          completedAt: new Date(),
        },
      });

      await this.emitLog({
        type: 'status',
        executionId,
        status: 'failed',
        message: error.message,
        timestamp: new Date().toISOString(),
      });
    }
  }

  /**
   * Simulate test execution (replace with real Selenium/Playwright integration)
   */
  private async simulateTestExecution(executionId: string, data: any): Promise<void> {
    const testNames = data.tests.includes('*')
      ? ['LoginTest', 'CheckoutTest', 'PaymentTest']
      : data.tests;

    for (const testName of testNames) {
      // Emit test started
      await this.emitLog({
        type: 'test_started',
        executionId,
        testName,
        timestamp: new Date().toISOString(),
      });

      // Simulate test execution
      const startTime = Date.now();
      await new Promise((resolve) => setTimeout(resolve, Math.random() * 3000 + 1000));
      const duration = Date.now() - startTime;

      // Random pass/fail (90% pass rate)
      const passed = Math.random() > 0.1;
      const status = passed ? TestStatus.PASSED : TestStatus.FAILED;

      // Create test result
      await this.db.testResult.create({
        data: {
          executionId,
          testName,
          className: `com.example.${testName}`,
          status,
          duration,
          errorMessage: passed ? null : 'Element not found: #submit-btn',
          stackTrace: passed ? null : 'org.openqa.selenium.NoSuchElementException...',
        },
      });

      // Emit test completed
      await this.emitLog({
        type: 'test_completed',
        executionId,
        testName,
        status,
        duration,
        timestamp: new Date().toISOString(),
      });

      // Emit logs
      await this.emitLog({
        type: 'log',
        executionId,
        testName,
        level: 'info',
        message: `Test ${testName} ${passed ? 'passed' : 'failed'}`,
        timestamp: new Date().toISOString(),
      });
    }
  }

  /**
   * Emit log to WebSocket (via Redis pub/sub)
   */
  async emitLog(log: TestLog): Promise<void> {
    await this.cacheService.publish(`logs:${log.executionId}`, log);
  }

  /**
   * Get queue position
   */
  private async getQueuePosition(executionId: string): Promise<number> {
    const jobs = await this.testQueue.getJobs(['waiting']);
    const position = jobs.findIndex((job) => job.data.executionId === executionId);
    return position >= 0 ? position + 1 : 0;
  }
}
