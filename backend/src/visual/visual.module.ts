import { Module } from '@nestjs/common';
import { VisualController } from './visual.controller';
import { VisualTestingService } from './visual-testing.service';
import { AuthModule } from '../auth/auth.module';
import { StorageModule } from '../storage/storage.module';

@Module({
  imports: [AuthModule, StorageModule],
  controllers: [VisualController],
  providers: [VisualTestingService],
  exports: [VisualTestingService],
})
export class VisualModule {}
