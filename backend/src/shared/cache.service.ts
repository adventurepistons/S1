import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import Redis from 'ioredis';

@Injectable()
export class CacheService implements OnModuleInit {
  private readonly logger = new Logger(CacheService.name);
  private redis: Redis;

  constructor(private configService: ConfigService) {
    const redisUrl = this.configService.get<string>('REDIS_URL') || 'redis://localhost:6379';
    this.redis = new Redis(redisUrl, {
      retryStrategy: (times) => {
        const delay = Math.min(times * 50, 2000);
        return delay;
      },
      maxRetriesPerRequest: 3,
    });
  }

  async onModuleInit() {
    this.redis.on('connect', () => {
      this.logger.log('✅ Redis connected successfully');
    });

    this.redis.on('error', (error) => {
      this.logger.error('❌ Redis connection error:', error);
    });

    this.redis.on('close', () => {
      this.logger.warn('⚠️ Redis connection closed');
    });
  }

  /**
   * Get value from cache
   */
  async get<T>(key: string): Promise<T | null> {
    try {
      const value = await this.redis.get(key);
      if (!value) return null;
      return JSON.parse(value) as T;
    } catch (error) {
      this.logger.error(`Error getting key ${key}:`, error);
      return null;
    }
  }

  /**
   * Set value in cache with optional TTL
   */
  async set(key: string, value: any, ttlSeconds?: number): Promise<void> {
    try {
      const serialized = JSON.stringify(value);
      if (ttlSeconds) {
        await this.redis.setex(key, ttlSeconds, serialized);
      } else {
        await this.redis.set(key, serialized);
      }
    } catch (error) {
      this.logger.error(`Error setting key ${key}:`, error);
    }
  }

  /**
   * Delete key from cache
   */
  async delete(key: string): Promise<void> {
    try {
      await this.redis.del(key);
    } catch (error) {
      this.logger.error(`Error deleting key ${key}:`, error);
    }
  }

  /**
   * Delete keys matching pattern
   */
  async deletePattern(pattern: string): Promise<void> {
    try {
      const keys = await this.redis.keys(pattern);
      if (keys.length > 0) {
        await this.redis.del(...keys);
      }
    } catch (error) {
      this.logger.error(`Error deleting pattern ${pattern}:`, error);
    }
  }

  /**
   * Check if key exists
   */
  async exists(key: string): Promise<boolean> {
    try {
      const result = await this.redis.exists(key);
      return result === 1;
    } catch (error) {
      this.logger.error(`Error checking existence of key ${key}:`, error);
      return false;
    }
  }

  /**
   * Increment counter
   */
  async increment(key: string, amount: number = 1): Promise<number> {
    try {
      return await this.redis.incrby(key, amount);
    } catch (error) {
      this.logger.error(`Error incrementing key ${key}:`, error);
      return 0;
    }
  }

  /**
   * Get or set value with factory function
   */
  async getOrSet<T>(
    key: string,
    factory: () => Promise<T>,
    ttlSeconds?: number,
  ): Promise<T> {
    // Try to get from cache
    const cached = await this.get<T>(key);
    if (cached !== null) {
      return cached;
    }

    // Cache miss - call factory
    const value = await factory();

    // Store in cache
    await this.set(key, value, ttlSeconds);

    return value;
  }

  /**
   * Publish message to channel (pub/sub)
   */
  async publish(channel: string, message: any): Promise<void> {
    try {
      await this.redis.publish(channel, JSON.stringify(message));
    } catch (error) {
      this.logger.error(`Error publishing to channel ${channel}:`, error);
    }
  }

  /**
   * Subscribe to channel (pub/sub)
   */
  subscribe(channel: string, callback: (message: any) => void): void {
    const subscriber = this.redis.duplicate();

    subscriber.subscribe(channel, (err) => {
      if (err) {
        this.logger.error(`Error subscribing to channel ${channel}:`, err);
      }
    });

    subscriber.on('message', (ch, message) => {
      if (ch === channel) {
        try {
          const parsed = JSON.parse(message);
          callback(parsed);
        } catch (error) {
          this.logger.error(`Error parsing message from channel ${channel}:`, error);
        }
      }
    });
  }

  /**
   * Get Redis client (for advanced operations)
   */
  getClient(): Redis {
    return this.redis;
  }
}
