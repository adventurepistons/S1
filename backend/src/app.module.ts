import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { ThrottlerModule } from '@nestjs/throttler';
import { BullModule } from '@nestjs/bullmq';
import { AuthModule } from './auth/auth.module';
import { StorageModule } from './storage/storage.module';
import { SharedModule } from './shared/shared.module';
import { TestsModule } from './tests/tests.module';
import { VisualModule } from './visual/visual.module';
import { ApiTestingModule } from './api-testing/api-testing.module';
import { CiCdModule } from './ci-cd/ci-cd.module';

@Module({
  imports: [
    // Configuration
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: '.env',
    }),

    // Rate Limiting
    ThrottlerModule.forRoot([
      {
        ttl: 60000, // 1 minute
        limit: 10, // Default limit (will be overridden by guards)
      },
    ]),

    // BullMQ Queue
    BullModule.forRootAsync({
      useFactory: () => ({
        connection: {
          host: process.env.REDIS_URL?.replace('redis://', '').split(':')[0] || 'localhost',
          port: parseInt(process.env.REDIS_URL?.split(':')[2] || '6379'),
        },
      }),
    }),

    // Feature Modules
    SharedModule,
    StorageModule,
    AuthModule,
    TestsModule, // ✅ Phase 2: Test Execution
    VisualModule, // ✅ Phase 3: Visual Testing
    ApiTestingModule, // ✅ Phase 3: API Testing
    CiCdModule, // ✅ Phase 3: CI/CD Integration
    // AIModule will be added later
  ],
  controllers: [],
  providers: [],
})
export class AppModule {}
