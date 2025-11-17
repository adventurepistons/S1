import { Module } from '@nestjs/common';
import { ApiTestingController } from './api-testing.controller';
import { ApiTestingService } from './api-testing.service';
import { ApiTestExecutor } from './api-test-executor.service';
import { AuthModule } from '../auth/auth.module';

@Module({
  imports: [AuthModule],
  controllers: [ApiTestingController],
  providers: [ApiTestingService, ApiTestExecutor],
  exports: [ApiTestingService],
})
export class ApiTestingModule {}
