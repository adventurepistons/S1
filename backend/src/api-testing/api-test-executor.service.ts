import { Injectable, Logger } from '@nestjs/common';

/**
 * API Test Executor - Handles batch execution and scheduling
 */
@Injectable()
export class ApiTestExecutor {
  private readonly logger = new Logger(ApiTestExecutor.name);

  /**
   * Execute tests in parallel with concurrency control
   */
  async executeInParallel<T>(
    items: T[],
    executor: (item: T) => Promise<any>,
    concurrency: number = 5,
  ): Promise<any[]> {
    const results: any[] = [];
    const executing: Promise<any>[] = [];

    for (const item of items) {
      const promise = executor(item).then((result) => {
        executing.splice(executing.indexOf(promise), 1);
        return result;
      });

      results.push(promise);
      executing.push(promise);

      if (executing.length >= concurrency) {
        await Promise.race(executing);
      }
    }

    return Promise.all(results);
  }
}
