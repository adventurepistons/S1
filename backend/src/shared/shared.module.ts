import { Global, Module } from '@nestjs/common';
import { DatabaseService } from './database.service';
import { CacheService } from './cache.service';

@Global()
@Module({
  providers: [DatabaseService, CacheService],
  exports: [DatabaseService, CacheService],
})
export class SharedModule {}
