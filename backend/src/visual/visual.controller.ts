import {
  Controller,
  Post,
  Get,
  Put,
  Body,
  Param,
  Query,
  UseGuards,
  UploadedFile,
  UseInterceptors,
  HttpCode,
  HttpStatus,
} from '@nestjs/common';
import { FileInterceptor } from '@nestjs/platform-express';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { VisualTestingService } from './visual-testing.service';

@Controller('visual')
@UseGuards(JwtAuthGuard)
export class VisualController {
  constructor(private visualService: VisualTestingService) {}

  /**
   * Capture baseline screenshot
   * POST /api/v1/visual/baselines
   */
  @Post('baselines')
  @UseInterceptors(FileInterceptor('screenshot'))
  @HttpCode(HttpStatus.CREATED)
  async captureBaseline(
    @CurrentUser('id') userId: string,
    @UploadedFile() file: Express.Multer.File,
    @Body()
    body: {
      projectId: string;
      testName: string;
      viewportWidth: number;
      viewportHeight: number;
      browser: string;
    },
  ) {
    return this.visualService.captureBaseline(
      body.projectId,
      body.testName,
      file.buffer,
      parseInt(body.viewportWidth.toString()),
      parseInt(body.viewportHeight.toString()),
      body.browser,
    );
  }

  /**
   * Compare screenshot with baseline
   * POST /api/v1/visual/compare
   */
  @Post('compare')
  @UseInterceptors(FileInterceptor('screenshot'))
  @HttpCode(HttpStatus.OK)
  async compareWithBaseline(
    @CurrentUser('id') userId: string,
    @UploadedFile() file: Express.Multer.File,
    @Body()
    body: {
      projectId: string;
      testName: string;
      executionId: string;
      viewportWidth: number;
      viewportHeight: number;
      browser: string;
      threshold?: number;
    },
  ) {
    return this.visualService.compareWithBaseline(
      body.projectId,
      body.testName,
      file.buffer,
      body.executionId,
      parseInt(body.viewportWidth.toString()),
      parseInt(body.viewportHeight.toString()),
      body.browser,
      body.threshold ? parseFloat(body.threshold.toString()) : 0.05,
    );
  }

  /**
   * Get all baselines
   * GET /api/v1/visual/baselines/:projectId
   */
  @Get('baselines/:projectId')
  async getBaselines(
    @CurrentUser('id') userId: string,
    @Param('projectId') projectId: string,
  ) {
    return this.visualService.getBaselines(projectId, userId);
  }

  /**
   * Update baseline
   * PUT /api/v1/visual/baselines/:baselineId
   */
  @Put('baselines/:baselineId')
  @UseInterceptors(FileInterceptor('screenshot'))
  async updateBaseline(
    @CurrentUser('id') userId: string,
    @Param('baselineId') baselineId: string,
    @UploadedFile() file: Express.Multer.File,
  ) {
    return this.visualService.updateBaseline(baselineId, userId, file.buffer);
  }

  /**
   * Get comparison history
   * GET /api/v1/visual/comparisons/:projectId?testName=LoginTest
   */
  @Get('comparisons/:projectId')
  async getComparisons(
    @CurrentUser('id') userId: string,
    @Param('projectId') projectId: string,
    @Query('testName') testName?: string,
  ) {
    return this.visualService.getComparisons(projectId, userId, testName);
  }
}
