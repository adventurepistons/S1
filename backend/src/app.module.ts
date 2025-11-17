import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { ThrottlerModule } from '@nestjs/throttler';
import { BullModule } from '@nestjs/bullmq';
import { AuthModule } from './auth/auth.module';
import { StorageModule } from './storage/storage.module';
import { SharedModule } from './shared/shared.module';

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
    // AIModule will be added next
    // TestsModule will be added in Phase 2
  ],
  controllers: [],
  providers: [],
})
export class AppModule {}
