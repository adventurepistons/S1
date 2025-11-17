import { Injectable, Logger, NotFoundException } from '@nestjs/common';
import { DatabaseService } from '../shared/database.service';
import { S3Service } from '../storage/s3.service';
import * as sharp from 'sharp';
import * as crypto from 'crypto';

export interface VisualComparisonResult {
  comparisonId: string;
  passed: boolean;
  pixelDiff: number;
  percentageDiff: number;
  diffImageUrl?: string;
  threshold: number;
}

@Injectable()
export class VisualTestingService {
  private readonly logger = new Logger(VisualTestingService.name);

  constructor(
    private db: DatabaseService,
    private s3: S3Service,
  ) {}

  /**
   * Capture and store baseline screenshot
   */
  async captureBaseline(
    projectId: string,
    testName: string,
    screenshotBuffer: Buffer,
    viewportWidth: number,
    viewportHeight: number,
    browser: string,
  ): Promise<any> {
    // Calculate image hash
    const imageHash = crypto.createHash('sha256').update(screenshotBuffer).digest('hex');

    // Upload to S3
    const { url } = await this.s3.uploadBaseline(screenshotBuffer, projectId, testName);

    // Save baseline to database
    const baseline = await this.db.visualBaseline.upsert({
      where: {
        projectId_testName_viewportWidth_viewportHeight_browser: {
          projectId,
          testName,
          viewportWidth,
          viewportHeight,
          browser,
        },
      },
      create: {
        projectId,
        testName,
        viewportWidth,
        viewportHeight,
        browser,
        imageUrl: url,
        imageHash,
      },
      update: {
        imageUrl: url,
        imageHash,
      },
    });

    this.logger.log(`Baseline captured: ${testName}`);

    return baseline;
  }

  /**
   * Compare screenshot with baseline
   */
  async compareWithBaseline(
    projectId: string,
    testName: string,
    screenshotBuffer: Buffer,
    executionId: string,
    viewportWidth: number,
    viewportHeight: number,
    browser: string,
    threshold: number = 0.05, // 5% default
  ): Promise<VisualComparisonResult> {
    // Find baseline
    const baseline = await this.db.visualBaseline.findUnique({
      where: {
        projectId_testName_viewportWidth_viewportHeight_browser: {
          projectId,
          testName,
          viewportWidth,
          viewportHeight,
          browser,
        },
      },
    });

    if (!baseline) {
      throw new NotFoundException('Baseline not found. Please capture a baseline first.');
    }

    // Download baseline image
    const baselineBuffer = await this.s3.getFile(this.extractS3Key(baseline.imageUrl));

    // Upload current screenshot
    const { url: screenshotUrl } = await this.s3.uploadScreenshot(
      screenshotBuffer,
      executionId,
      testName,
    );

    // Compare images
    const { pixelDiff, percentageDiff, diffBuffer } = await this.compareImages(
      baselineBuffer,
      screenshotBuffer,
      viewportWidth,
      viewportHeight,
    );

    // Upload diff image if there are differences
    let diffImageUrl: string | undefined;
    if (pixelDiff > 0 && diffBuffer) {
      const { url } = await this.s3.uploadFile(
        diffBuffer,
        `diff_${testName}.png`,
        'visual-diffs',
        'image/png',
      );
      diffImageUrl = url;
    }

    // Determine if comparison passed
    const passed = percentageDiff <= threshold;

    // Save comparison to database
    const comparison = await this.db.visualComparison.create({
      data: {
        baselineId: baseline.id,
        executionId,
        screenshotUrl,
        diffUrl: diffImageUrl,
        pixelDiff,
        percentageDiff,
        threshold,
        passed,
      },
    });

    this.logger.log(
      `Visual comparison: ${testName} - ${passed ? 'PASSED' : 'FAILED'} (${percentageDiff.toFixed(2)}% diff)`,
    );

    return {
      comparisonId: comparison.id,
      passed,
      pixelDiff,
      percentageDiff,
      diffImageUrl,
      threshold,
    };
  }

  /**
   * Get all baselines for a project
   */
  async getBaselines(projectId: string, userId: string): Promise<any[]> {
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

    return this.db.visualBaseline.findMany({
      where: { projectId },
      orderBy: { createdAt: 'desc' },
    });
  }

  /**
   * Update baseline (approve new screenshot)
   */
  async updateBaseline(
    baselineId: string,
    userId: string,
    screenshotBuffer: Buffer,
  ): Promise<any> {
    const baseline = await this.db.visualBaseline.findUnique({
      where: { id: baselineId },
      include: { project: true },
    });

    if (!baseline) {
      throw new NotFoundException('Baseline not found');
    }

    // Verify user has access
    if (baseline.project.userId !== userId) {
      throw new NotFoundException('Baseline not found');
    }

    // Calculate new hash
    const imageHash = crypto.createHash('sha256').update(screenshotBuffer).digest('hex');

    // Upload new baseline
    const { url } = await this.s3.uploadBaseline(
      screenshotBuffer,
      baseline.projectId,
      baseline.testName,
    );

    // Update baseline
    const updated = await this.db.visualBaseline.update({
      where: { id: baselineId },
      data: {
        imageUrl: url,
        imageHash,
        approvedBy: userId,
        approvedAt: new Date(),
      },
    });

    this.logger.log(`Baseline updated: ${baseline.testName}`);

    return updated;
  }

  /**
   * Get comparison history
   */
  async getComparisons(
    projectId: string,
    userId: string,
    testName?: string,
  ): Promise<any[]> {
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

    const where: any = {
      baseline: { projectId },
    };

    if (testName) {
      where.baseline.testName = testName;
    }

    return this.db.visualComparison.findMany({
      where,
      include: {
        baseline: {
          select: {
            testName: true,
            viewportWidth: true,
            viewportHeight: true,
            browser: true,
          },
        },
      },
      orderBy: { createdAt: 'desc' },
      take: 50,
    });
  }

  /**
   * Compare two images pixel by pixel
   */
  private async compareImages(
    baselineBuffer: Buffer,
    currentBuffer: Buffer,
    width: number,
    height: number,
  ): Promise<{ pixelDiff: number; percentageDiff: number; diffBuffer?: Buffer }> {
    try {
      // Resize images to same dimensions if needed
      const baseline = await sharp(baselineBuffer)
        .resize(width, height, { fit: 'fill' })
        .raw()
        .toBuffer({ resolveWithObject: true });

      const current = await sharp(currentBuffer)
        .resize(width, height, { fit: 'fill' })
        .raw()
        .toBuffer({ resolveWithObject: true });

      const baselineData = baseline.data;
      const currentData = current.data;

      // Compare pixel by pixel
      let diffPixels = 0;
      const diffData = Buffer.alloc(baselineData.length);

      for (let i = 0; i < baselineData.length; i += 4) {
        const rDiff = Math.abs(baselineData[i] - currentData[i]);
        const gDiff = Math.abs(baselineData[i + 1] - currentData[i + 1]);
        const bDiff = Math.abs(baselineData[i + 2] - currentData[i + 2]);

        const totalDiff = rDiff + gDiff + bDiff;

        if (totalDiff > 30) {
          // Threshold for considering pixels different
          diffPixels++;
          // Highlight difference in red
          diffData[i] = 255;
          diffData[i + 1] = 0;
          diffData[i + 2] = 0;
          diffData[i + 3] = 255;
        } else {
          // Keep original pixel
          diffData[i] = currentData[i];
          diffData[i + 1] = currentData[i + 1];
          diffData[i + 2] = currentData[i + 2];
          diffData[i + 3] = currentData[i + 3];
        }
      }

      const totalPixels = width * height;
      const percentageDiff = (diffPixels / totalPixels) * 100;

      // Create diff image
      let diffBuffer: Buffer | undefined;
      if (diffPixels > 0) {
        diffBuffer = await sharp(diffData, {
          raw: {
            width,
            height,
            channels: 4,
          },
        })
          .png()
          .toBuffer();
      }

      return {
        pixelDiff: diffPixels,
        percentageDiff,
        diffBuffer,
      };
    } catch (error) {
      this.logger.error(`Image comparison failed: ${error.message}`);
      throw error;
    }
  }

  /**
   * Extract S3 key from URL
   */
  private extractS3Key(url: string): string {
    const urlParts = url.split('.com/');
    return urlParts[1] || url;
  }
}
