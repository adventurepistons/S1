import {
  Controller,
  Post,
  Get,
  Delete,
  Body,
  Param,
  Query,
  UseGuards,
  HttpCode,
  HttpStatus,
} from '@nestjs/common';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { QuotaGuard } from '../auth/guards/quota.guard';
import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { Tool } from '../auth/decorators/tool.decorator';
import { TestRunnerService } from './services/test-runner.service';
import { TestReportingService } from './services/test-reporting.service';
import { QuotaService } from '../auth/quota.service';
import { ExecuteTestsDto } from './dto/execute-tests.dto';

@Controller('tests')
@UseGuards(JwtAuthGuard)
export class TestsController {
  constructor(
    private testRunner: TestRunnerService,
    private reporting: TestReportingService,
    private quotaService: QuotaService,
  ) {}

  /**
   * Execute tests
   * POST /api/v1/tests/execute
   */
  @Post('execute')
  @UseGuards(QuotaGuard)
  @Tool('test-copilot')
  @HttpCode(HttpStatus.ACCEPTED)
  async executeTests(
    @CurrentUser('id') userId: string,
    @Body() dto: ExecuteTestsDto,
  ) {
    const result = await this.testRunner.executeTests(userId, dto);

    // Increment quota
    await this.quotaService.incrementUsage(userId, 'test-copilot', 1);

    return result;
  }

  /**
   * Get execution status
   * GET /api/v1/tests/status/:executionId
   */
  @Get('status/:executionId')
  async getExecutionStatus(
    @CurrentUser('id') userId: string,
    @Param('executionId') executionId: string,
  ) {
    return this.testRunner.getExecutionStatus(executionId, userId);
  }

  /**
   * Cancel execution
   * DELETE /api/v1/tests/cancel/:executionId
   */
  @Delete('cancel/:executionId')
  @HttpCode(HttpStatus.OK)
  async cancelExecution(
    @CurrentUser('id') userId: string,
    @Param('executionId') executionId: string,
  ) {
    await this.testRunner.cancelExecution(executionId, userId);
    return { message: 'Execution cancelled' };
  }

  /**
   * Get execution report
   * GET /api/v1/tests/reports/:executionId
   */
  @Get('reports/:executionId')
  async getExecutionReport(
    @CurrentUser('id') userId: string,
    @Param('executionId') executionId: string,
  ) {
    return this.reporting.getExecutionReport(executionId, userId);
  }

  /**
   * Get execution history
   * GET /api/v1/tests/history/:projectId?days=30
   */
  @Get('history/:projectId')
  async getExecutionHistory(
    @CurrentUser('id') userId: string,
    @Param('projectId') projectId: string,
    @Query('days') days?: number,
  ) {
    return this.reporting.getExecutionHistory(projectId, userId, days ? parseInt(days.toString()) : 30);
  }

  /**
   * Get flaky tests
   * GET /api/v1/tests/flaky/:projectId
   */
  @Get('flaky/:projectId')
  async getFlakyTests(
    @CurrentUser('id') userId: string,
    @Param('projectId') projectId: string,
  ) {
    return this.reporting.getFlakyTests(projectId, userId);
  }

  /**
   * Get test coverage
   * GET /api/v1/tests/coverage/:projectId
   */
  @Get('coverage/:projectId')
  async getTestCoverage(
    @CurrentUser('id') userId: string,
    @Param('projectId') projectId: string,
  ) {
    return this.reporting.getTestCoverage(projectId, userId);
  }
}
