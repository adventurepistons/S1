import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as AWS from 'aws-sdk';
import { v4 as uuidv4 } from 'uuid';

@Injectable()
export class S3Service {
  private readonly logger = new Logger(S3Service.name);
  private s3: AWS.S3;
  private bucket: string;

  constructor(private configService: ConfigService) {
    this.bucket = this.configService.get<string>('AWS_S3_BUCKET') || 'testcopilot-dev';

    this.s3 = new AWS.S3({
      accessKeyId: this.configService.get<string>('AWS_ACCESS_KEY_ID'),
      secretAccessKey: this.configService.get<string>('AWS_SECRET_ACCESS_KEY'),
      region: this.configService.get<string>('AWS_REGION') || 'us-east-1',
    });

    this.logger.log(`S3Service initialized with bucket: ${this.bucket}`);
  }

  /**
   * Upload file to S3
   */
  async uploadFile(
    file: Buffer,
    filename: string,
    folder: string = 'uploads',
    contentType?: string,
  ): Promise<{ url: string; key: string }> {
    const key = `${folder}/${uuidv4()}-${filename}`;

    try {
      const params: AWS.S3.PutObjectRequest = {
        Bucket: this.bucket,
        Key: key,
        Body: file,
        ContentType: contentType || 'application/octet-stream',
        ACL: 'private', // Files are private by default
      };

      await this.s3.upload(params).promise();

      const url = `https://${this.bucket}.s3.amazonaws.com/${key}`;

      this.logger.log(`File uploaded successfully: ${key}`);

      return { url, key };
    } catch (error) {
      this.logger.error(`Error uploading file to S3: ${error.message}`);
      throw error;
    }
  }

  /**
   * Upload screenshot
   */
  async uploadScreenshot(
    file: Buffer,
    executionId: string,
    testName: string,
  ): Promise<{ url: string; key: string }> {
    const filename = `${executionId}_${testName.replace(/[^a-zA-Z0-9]/g, '_')}_${Date.now()}.png`;
    return this.uploadFile(file, filename, 'screenshots', 'image/png');
  }

  /**
   * Upload video
   */
  async uploadVideo(
    file: Buffer,
    executionId: string,
    testName: string,
  ): Promise<{ url: string; key: string }> {
    const filename = `${executionId}_${testName.replace(/[^a-zA-Z0-9]/g, '_')}_${Date.now()}.mp4`;
    return this.uploadFile(file, filename, 'videos', 'video/mp4');
  }

  /**
   * Upload visual baseline
   */
  async uploadBaseline(
    file: Buffer,
    projectId: string,
    testName: string,
  ): Promise<{ url: string; key: string }> {
    const filename = `${projectId}_${testName.replace(/[^a-zA-Z0-9]/g, '_')}.png`;
    return this.uploadFile(file, filename, 'baselines', 'image/png');
  }

  /**
   * Get signed URL for temporary access
   */
  async getSignedUrl(key: string, expiresIn: number = 3600): Promise<string> {
    try {
      const params = {
        Bucket: this.bucket,
        Key: key,
        Expires: expiresIn,
      };

      const url = await this.s3.getSignedUrlPromise('getObject', params);
      return url;
    } catch (error) {
      this.logger.error(`Error generating signed URL: ${error.message}`);
      throw error;
    }
  }

  /**
   * Delete file from S3
   */
  async deleteFile(key: string): Promise<void> {
    try {
      await this.s3
        .deleteObject({
          Bucket: this.bucket,
          Key: key,
        })
        .promise();

      this.logger.log(`File deleted: ${key}`);
    } catch (error) {
      this.logger.error(`Error deleting file: ${error.message}`);
      throw error;
    }
  }

  /**
   * Delete files older than specified days
   */
  async deleteOldFiles(folder: string, olderThanDays: number): Promise<number> {
    try {
      const cutoffDate = new Date();
      cutoffDate.setDate(cutoffDate.getDate() - olderThanDays);

      const params = {
        Bucket: this.bucket,
        Prefix: folder + '/',
      };

      const objects = await this.s3.listObjectsV2(params).promise();

      if (!objects.Contents || objects.Contents.length === 0) {
        return 0;
      }

      const oldObjects = objects.Contents.filter(
        (obj) => obj.LastModified && obj.LastModified < cutoffDate,
      );

      if (oldObjects.length === 0) {
        return 0;
      }

      const deleteParams = {
        Bucket: this.bucket,
        Delete: {
          Objects: oldObjects.map((obj) => ({ Key: obj.Key! })),
        },
      };

      await this.s3.deleteObjects(deleteParams).promise();

      this.logger.log(`Deleted ${oldObjects.length} old files from ${folder}`);

      return oldObjects.length;
    } catch (error) {
      this.logger.error(`Error deleting old files: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get file from S3
   */
  async getFile(key: string): Promise<Buffer> {
    try {
      const params = {
        Bucket: this.bucket,
        Key: key,
      };

      const data = await this.s3.getObject(params).promise();
      return data.Body as Buffer;
    } catch (error) {
      this.logger.error(`Error getting file from S3: ${error.message}`);
      throw error;
    }
  }
}
