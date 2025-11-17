import { Processor, WorkerHost } from '@nestjs/bullmq';
import { Logger } from '@nestjs/common';
import { Job } from 'bullmq';
import { TestRunnerService } from './services/test-runner.service';

@Processor('test-execution')
export class TestExecutionProcessor extends WorkerHost {
  private readonly logger = new Logger(TestExecutionProcessor.name);

  constructor(private testRunner: TestRunnerService) {
    super();
  }

  async process(job: Job<any>): Promise<any> {
    this.logger.log(`Processing job ${job.id}: ${job.name}`);

    const { executionId } = job.data;

    try {
      await this.testRunner.runTests(executionId, job.data);

      this.logger.log(`Job ${job.id} completed successfully`);

      return { success: true, executionId };
    } catch (error) {
      this.logger.error(`Job ${job.id} failed:`, error);
      throw error;
    }
  }
}
