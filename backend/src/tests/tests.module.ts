import { Module } from '@nestjs/common';
import { BullModule } from '@nestjs/bullmq';
import { TestsController } from './tests.controller';
import { TestRunnerService } from './services/test-runner.service';
import { TestReportingService } from './services/test-reporting.service';
import { TestExecutionGateway } from './test-execution.gateway';
import { TestExecutionProcessor } from './test-execution.processor';
import { AuthModule } from '../auth/auth.module';

@Module({
  imports: [
    // Import AuthModule to use QuotaService
    AuthModule,
    // Register test execution queue
    BullModule.registerQueue({
      name: 'test-execution',
    }),
  ],
  controllers: [TestsController],
  providers: [
    TestRunnerService,
    TestReportingService,
    TestExecutionGateway,
    TestExecutionProcessor,
  ],
  exports: [TestRunnerService, TestReportingService],
})
export class TestsModule {}
