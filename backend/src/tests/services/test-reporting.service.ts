import { Injectable, Logger, NotFoundException } from '@nestjs/common';
import { DatabaseService } from '../../shared/database.service';
import { TestStatus } from '@prisma/client';

@Injectable()
export class TestReportingService {
  private readonly logger = new Logger(TestReportingService.name);

  constructor(private db: DatabaseService) {}

  /**
   * Get execution report
   */
  async getExecutionReport(executionId: string, userId: string): Promise<any> {
    const execution = await this.db.testExecution.findFirst({
      where: {
        id: executionId,
        userId,
      },
      include: {
        project: {
          select: {
            id: true,
            name: true,
            framework: true,
          },
        },
        results: {
          orderBy: {
            createdAt: 'asc',
          },
        },
      },
    });

    if (!execution) {
      throw new NotFoundException('Execution not found');
    }

    // Calculate flaky tests
    const flakyTests = await this.detectFlakyTests(execution.projectId);

    return {
      executionId: execution.id,
      projectId: execution.projectId,
      projectName: execution.project.name,
      framework: execution.project.framework,
      status: execution.status,
      environment: execution.environment,
      browser: execution.browser,
      headless: execution.headless,
      parallel: execution.parallel,
      startedAt: execution.startedAt,
      completedAt: execution.completedAt,
      duration: execution.duration,
      summary: {
        total: execution.totalTests || 0,
        passed: execution.passedTests || 0,
        failed: execution.failedTests || 0,
        skipped: execution.skippedTests || 0,
        passRate: execution.totalTests
          ? ((execution.passedTests || 0) / execution.totalTests) * 100
          : 0,
      },
      tests: execution.results.map((result) => ({
        id: result.id,
        testName: result.testName,
        className: result.className,
        status: result.status,
        duration: result.duration,
        retries: result.retries,
        errorMessage: result.errorMessage,
        stackTrace: result.stackTrace,
        screenshotUrl: result.screenshotUrl,
        videoUrl: result.videoUrl,
      })),
      flakyTests: flakyTests.filter((ft) =>
        execution.results.some((r) => r.testName === ft.testName),
      ),
    };
  }

  /**
   * Get execution history for a project
   */
  async getExecutionHistory(
    projectId: string,
    userId: string,
    days: number = 30,
  ): Promise<any> {
    const cutoffDate = new Date();
    cutoffDate.setDate(cutoffDate.getDate() - days);

    const executions = await this.db.testExecution.findMany({
      where: {
        projectId,
        userId,
        createdAt: {
          gte: cutoffDate,
        },
      },
      orderBy: {
        createdAt: 'desc',
      },
      select: {
        id: true,
        status: true,
        createdAt: true,
        startedAt: true,
        completedAt: true,
        duration: true,
        totalTests: true,
        passedTests: true,
        failedTests: true,
        skippedTests: true,
        browser: true,
        environment: true,
      },
    });

    // Calculate trends
    const totalExecutions = executions.length;
    const successfulExecutions = executions.filter(
      (e) => e.status === 'COMPLETED' && e.failedTests === 0,
    ).length;

    const averagePassRate =
      executions.reduce((sum, e) => {
        const passRate = e.totalTests ? (e.passedTests || 0) / e.totalTests : 0;
        return sum + passRate;
      }, 0) / (totalExecutions || 1);

    const averageDuration =
      executions.reduce((sum, e) => sum + (e.duration || 0), 0) / (totalExecutions || 1);

    return {
      projectId,
      days,
      executions,
      trends: {
        totalExecutions,
        successfulExecutions,
        successRate: totalExecutions ? (successfulExecutions / totalExecutions) * 100 : 0,
        averagePassRate: averagePassRate * 100,
        averageDuration,
      },
    };
  }

  /**
   * Detect flaky tests
   */
  async detectFlakyTests(projectId: string): Promise<any[]> {
    // Get all test results for this project
    const results = await this.db.testResult.findMany({
      where: {
        execution: {
          projectId,
        },
      },
      select: {
        testName: true,
        className: true,
        status: true,
        createdAt: true,
      },
      orderBy: {
        createdAt: 'desc',
      },
      take: 1000, // Last 1000 test results
    });

    // Group by test name
    const testGroups = results.reduce((acc, result) => {
      if (!acc[result.testName]) {
        acc[result.testName] = [];
      }
      acc[result.testName].push(result);
      return acc;
    }, {} as Record<string, any[]>);

    // Calculate flaky score for each test
    const flakyTests = [];

    for (const [testName, runs] of Object.entries(testGroups)) {
      if (runs.length < 3) continue; // Need at least 3 runs

      const totalRuns = runs.length;
      const passedRuns = runs.filter((r) => r.status === TestStatus.PASSED).length;
      const failedRuns = runs.filter((r) => r.status === TestStatus.FAILED).length;

      // Flaky if test has both passes and failures
      if (passedRuns > 0 && failedRuns > 0) {
        const flakyScore = failedRuns / totalRuns;

        // Only consider flaky if failure rate is between 10% and 90%
        if (flakyScore > 0.1 && flakyScore < 0.9) {
          flakyTests.push({
            testName,
            className: runs[0].className,
            flakyScore,
            totalRuns,
            passedRuns,
            failedRuns,
            lastFailedAt: runs.find((r) => r.status === TestStatus.FAILED)?.createdAt,
          });

          // Update flaky_tests table
          await this.db.flakyTest.upsert({
            where: {
              projectId_testName: {
                projectId,
                testName,
              },
            },
            create: {
              projectId,
              testName,
              className: runs[0].className,
              flakyScore,
              totalRuns,
              passedRuns,
              failedRuns,
              lastFailedAt: runs.find((r) => r.status === TestStatus.FAILED)?.createdAt,
            },
            update: {
              flakyScore,
              totalRuns,
              passedRuns,
              failedRuns,
              lastFailedAt: runs.find((r) => r.status === TestStatus.FAILED)?.createdAt,
            },
          });
        }
      }
    }

    return flakyTests.sort((a, b) => b.flakyScore - a.flakyScore);
  }

  /**
   * Get flaky tests for a project
   */
  async getFlakyTests(projectId: string, userId: string): Promise<any[]> {
    // Verify project access
    const project = await this.db.project.findFirst({
      where: {
        id: projectId,
        userId,
      },
    });

    if (!project) {
      throw new NotFoundException('Project not found');
    }

    return this.db.flakyTest.findMany({
      where: { projectId },
      orderBy: {
        flakyScore: 'desc',
      },
    });
  }

  /**
   * Get test coverage info
   */
  async getTestCoverage(projectId: string, userId: string): Promise<any> {
    // Verify project access
    const project = await this.db.project.findFirst({
      where: {
        id: projectId,
        userId,
      },
    });

    if (!project) {
      throw new NotFoundException('Project not found');
    }

    // Get unique tests that have been run
    const executedTests = await this.db.testResult.groupBy({
      by: ['testName'],
      where: {
        execution: {
          projectId,
        },
      },
      _count: {
        testName: true,
      },
    });

    // Get all indexed tests from code
    const indexedTests = await this.db.codeIndex.count({
      where: {
        projectId,
        codeType: 'test',
      },
    });

    return {
      totalTests: indexedTests,
      executedTests: executedTests.length,
      coverage: indexedTests > 0 ? (executedTests.length / indexedTests) * 100 : 0,
      tests: executedTests.map((t) => ({
        testName: t.testName,
        executions: t._count.testName,
      })),
    };
  }
}
