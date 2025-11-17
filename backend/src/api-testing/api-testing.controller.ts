import {
  Controller,
  Post,
  Body,
  UseGuards,
  HttpCode,
  HttpStatus,
} from '@nestjs/common';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { QuotaGuard } from '../auth/guards/quota.guard';
import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { Tool } from '../auth/decorators/tool.decorator';
import { ApiTestingService, ApiTestDefinition } from './api-testing.service';
import { QuotaService } from '../auth/quota.service';

@Controller('api-tests')
@UseGuards(JwtAuthGuard)
export class ApiTestingController {
  constructor(
    private apiTestingService: ApiTestingService,
    private quotaService: QuotaService,
  ) {}

  /**
   * Execute API tests
   * POST /api/v1/api-tests/execute
   */
  @Post('execute')
  @UseGuards(QuotaGuard)
  @Tool('test-copilot')
  @HttpCode(HttpStatus.OK)
  async executeTests(
    @CurrentUser('id') userId: string,
    @Body() body: { tests: ApiTestDefinition[] },
  ) {
    const results = await this.apiTestingService.executeTests(body.tests);

    // Increment quota
    await this.quotaService.incrementUsage(userId, 'test-copilot', body.tests.length);

    const summary = {
      total: results.length,
      passed: results.filter((r) => r.passed).length,
      failed: results.filter((r) => r.passed === false).length,
    };

    return {
      summary,
      results,
    };
  }

  /**
   * Generate tests from OpenAPI spec
   * POST /api/v1/api-tests/generate/openapi
   */
  @Post('generate/openapi')
  @HttpCode(HttpStatus.OK)
  async generateFromOpenAPI(@Body() body: { specUrl: string }) {
    const tests = await this.apiTestingService.generateTestsFromOpenAPI(body.specUrl);

    return {
      count: tests.length,
      tests,
    };
  }

  /**
   * Generate tests from GraphQL schema
   * POST /api/v1/api-tests/generate/graphql
   */
  @Post('generate/graphql')
  @HttpCode(HttpStatus.OK)
  async generateFromGraphQL(@Body() body: { url: string }) {
    const tests = await this.apiTestingService.generateTestsFromGraphQL(body.url);

    return {
      count: tests.length,
      tests,
    };
  }
}
