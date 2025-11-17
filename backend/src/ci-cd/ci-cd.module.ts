import { Module } from '@nestjs/common';
import { CiCdController } from './ci-cd.controller';
import { WebhookService } from './webhook.service';
import { GithubService } from './github.service';
import { NotificationService } from './notification.service';
import { TestsModule } from '../tests/tests.module';
import { AuthModule } from '../auth/auth.module';

@Module({
  imports: [TestsModule, AuthModule],
  controllers: [CiCdController],
  providers: [WebhookService, GithubService, NotificationService],
  exports: [WebhookService],
})
export class CiCdModule {}
