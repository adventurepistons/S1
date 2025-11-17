import { Injectable, Logger } from '@nestjs/common';
import { Builder, By, WebDriver, until } from 'selenium-webdriver';
import { Options as ChromeOptions } from 'selenium-webdriver/chrome';
import { Options as FirefoxOptions } from 'selenium-webdriver/firefox';
import * as path from 'path';
import * as fs from 'fs';
import { S3Service } from '../../storage/s3.service';
import { TestRunnerService } from './test-runner.service';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export interface TestExecutionConfig {
  executionId: string;
  projectPath: string;
  tests: string[];
  browser: string;
  headless: boolean;
  timeout: number;
  retryFailed: number;
}

export interface TestResult {
  testName: string;
  className: string;
  status: 'passed' | 'failed' | 'skipped';
  duration: number;
  errorMessage?: string;
  stackTrace?: string;
  screenshotUrl?: string;
  videoUrl?: string;
  logs?: string;
}

@Injectable()
export class SeleniumExecutorService {
  private readonly logger = new Logger(SeleniumExecutorService.name);

  constructor(
    private s3Service: S3Service,
    private testRunner: TestRunnerService,
  ) {}

  /**
   * Execute Java/Selenium tests using Maven or Gradle
   */
  async executeJavaTests(config: TestExecutionConfig): Promise<TestResult[]> {
    this.logger.log(`Executing Java tests for execution: ${config.executionId}`);

    const results: TestResult[] = [];

    try {
      // Detect build tool (Maven or Gradle)
      const buildTool = await this.detectBuildTool(config.projectPath);

      // Build test command
      const testCommand = this.buildTestCommand(buildTool, config);

      this.logger.log(`Running command: ${testCommand}`);

      // Emit log
      await this.testRunner.emitLog({
        type: 'log',
        executionId: config.executionId,
        level: 'info',
        message: `Executing: ${testCommand}`,
        timestamp: new Date().toISOString(),
      });

      // Execute tests
      const { stdout, stderr } = await execAsync(testCommand, {
        cwd: config.projectPath,
        timeout: config.timeout * 1000,
        maxBuffer: 10 * 1024 * 1024, // 10MB
      });

      // Parse test results from output
      const parsedResults = this.parseTestOutput(stdout, stderr, buildTool);

      // Process each result
      for (const result of parsedResults) {
        // Capture screenshot for failed tests
        if (result.status === 'failed' && result.screenshotPath) {
          const screenshotUrl = await this.uploadScreenshot(
            config.executionId,
            result.testName,
            result.screenshotPath,
          );
          result.screenshotUrl = screenshotUrl;
        }

        results.push(result);

        // Emit test completion
        await this.testRunner.emitLog({
          type: 'test_completed',
          executionId: config.executionId,
          testName: result.testName,
          status: result.status,
          duration: result.duration,
          timestamp: new Date().toISOString(),
        });
      }

      this.logger.log(`Tests completed. Total: ${results.length}`);
    } catch (error) {
      this.logger.error(`Test execution failed: ${error.message}`);

      // Emit error
      await this.testRunner.emitLog({
        type: 'log',
        executionId: config.executionId,
        level: 'error',
        message: `Execution failed: ${error.message}`,
        timestamp: new Date().toISOString(),
      });

      throw error;
    }

    return results;
  }

  /**
   * Execute Playwright tests (TypeScript/JavaScript)
   */
  async executePlaywrightTests(config: TestExecutionConfig): Promise<TestResult[]> {
    this.logger.log(`Executing Playwright tests for execution: ${config.executionId}`);

    const results: TestResult[] = [];

    try {
      const testCommand = `npx playwright test ${config.tests.join(' ')} --reporter=json`;

      this.logger.log(`Running command: ${testCommand}`);

      await this.testRunner.emitLog({
        type: 'log',
        executionId: config.executionId,
        level: 'info',
        message: `Executing Playwright tests`,
        timestamp: new Date().toISOString(),
      });

      const { stdout, stderr } = await execAsync(testCommand, {
        cwd: config.projectPath,
        timeout: config.timeout * 1000,
        env: {
          ...process.env,
          PWTEST_SKIP_TEST_OUTPUT: '1',
        },
      });

      // Parse Playwright JSON output
      const parsedResults = this.parsePlaywrightOutput(stdout);

      for (const result of parsedResults) {
        results.push(result);

        await this.testRunner.emitLog({
          type: 'test_completed',
          executionId: config.executionId,
          testName: result.testName,
          status: result.status,
          duration: result.duration,
          timestamp: new Date().toISOString(),
        });
      }
    } catch (error) {
      this.logger.error(`Playwright execution failed: ${error.message}`);
      throw error;
    }

    return results;
  }

  /**
   * Execute Pytest tests (Python/Selenium)
   */
  async executePytestTests(config: TestExecutionConfig): Promise<TestResult[]> {
    this.logger.log(`Executing Pytest tests for execution: ${config.executionId}`);

    const results: TestResult[] = [];

    try {
      const testCommand = `pytest ${config.tests.join(' ')} --json-report --json-report-file=report.json`;

      await this.testRunner.emitLog({
        type: 'log',
        executionId: config.executionId,
        level: 'info',
        message: `Executing Pytest tests`,
        timestamp: new Date().toISOString(),
      });

      const { stdout, stderr } = await execAsync(testCommand, {
        cwd: config.projectPath,
        timeout: config.timeout * 1000,
      });

      // Parse pytest JSON report
      const reportPath = path.join(config.projectPath, 'report.json');
      if (fs.existsSync(reportPath)) {
        const reportData = JSON.parse(fs.readFileSync(reportPath, 'utf-8'));
        const parsedResults = this.parsePytestOutput(reportData);

        for (const result of parsedResults) {
          results.push(result);

          await this.testRunner.emitLog({
            type: 'test_completed',
            executionId: config.executionId,
            testName: result.testName,
            status: result.status,
            duration: result.duration,
            timestamp: new Date().toISOString(),
          });
        }
      }
    } catch (error) {
      this.logger.error(`Pytest execution failed: ${error.message}`);
      throw error;
    }

    return results;
  }

  /**
   * Capture screenshot
   */
  async captureScreenshot(
    driver: WebDriver,
    executionId: string,
    testName: string,
  ): Promise<string | null> {
    try {
      const screenshot = await driver.takeScreenshot();
      const buffer = Buffer.from(screenshot, 'base64');

      const { url } = await this.s3Service.uploadScreenshot(buffer, executionId, testName);

      this.logger.log(`Screenshot captured: ${url}`);

      // Emit screenshot event
      await this.testRunner.emitLog({
        type: 'screenshot',
        executionId,
        testName,
        url,
        timestamp: new Date().toISOString(),
      });

      return url;
    } catch (error) {
      this.logger.error(`Failed to capture screenshot: ${error.message}`);
      return null;
    }
  }

  /**
   * Detect build tool (Maven or Gradle)
   */
  private async detectBuildTool(projectPath: string): Promise<'maven' | 'gradle'> {
    const pomPath = path.join(projectPath, 'pom.xml');
    const gradlePath = path.join(projectPath, 'build.gradle');

    if (fs.existsSync(pomPath)) {
      return 'maven';
    } else if (fs.existsSync(gradlePath)) {
      return 'gradle';
    }

    throw new Error('No Maven or Gradle project found');
  }

  /**
   * Build test command
   */
  private buildTestCommand(buildTool: 'maven' | 'gradle', config: TestExecutionConfig): string {
    const testFilter = config.tests.includes('*') ? '' : `-Dtest=${config.tests.join(',')}`;

    if (buildTool === 'maven') {
      return `mvn test ${testFilter} -Dheadless=${config.headless}`;
    } else {
      return `gradle test ${testFilter} -Dheadless=${config.headless}`;
    }
  }

  /**
   * Parse test output from Maven/Gradle
   */
  private parseTestOutput(
    stdout: string,
    stderr: string,
    buildTool: string,
  ): TestResult[] {
    const results: TestResult[] = [];

    // Simple parsing (in production, use JUnit XML parser)
    const lines = stdout.split('\n');

    for (const line of lines) {
      // Example: [INFO] Tests run: 3, Failures: 1, Errors: 0, Skipped: 0
      const match = line.match(/Tests run: (\d+), Failures: (\d+), Errors: (\d+), Skipped: (\d+)/);

      if (match) {
        const [, total, failures, errors, skipped] = match;

        // This is simplified - in production, parse JUnit XML for detailed results
        results.push({
          testName: 'TestSuite',
          className: 'com.example.TestSuite',
          status: parseInt(failures) + parseInt(errors) > 0 ? 'failed' : 'passed',
          duration: 0,
        });
      }
    }

    return results;
  }

  /**
   * Parse Playwright JSON output
   */
  private parsePlaywrightOutput(stdout: string): TestResult[] {
    try {
      const report = JSON.parse(stdout);
      const results: TestResult[] = [];

      for (const suite of report.suites || []) {
        for (const test of suite.specs || []) {
          results.push({
            testName: test.title,
            className: suite.file,
            status: test.ok ? 'passed' : 'failed',
            duration: test.tests?.[0]?.results?.[0]?.duration || 0,
            errorMessage: test.tests?.[0]?.results?.[0]?.error?.message,
          });
        }
      }

      return results;
    } catch (error) {
      this.logger.error(`Failed to parse Playwright output: ${error.message}`);
      return [];
    }
  }

  /**
   * Parse Pytest JSON output
   */
  private parsePytestOutput(reportData: any): TestResult[] {
    const results: TestResult[] = [];

    for (const test of reportData.tests || []) {
      results.push({
        testName: test.nodeid,
        className: test.nodeid.split('::')[0],
        status: test.outcome === 'passed' ? 'passed' : 'failed',
        duration: test.duration * 1000, // Convert to ms
        errorMessage: test.call?.longrepr,
      });
    }

    return results;
  }

  /**
   * Upload screenshot to S3
   */
  private async uploadScreenshot(
    executionId: string,
    testName: string,
    screenshotPath: string,
  ): Promise<string> {
    const buffer = fs.readFileSync(screenshotPath);
    const { url } = await this.s3Service.uploadScreenshot(buffer, executionId, testName);
    return url;
  }
}
