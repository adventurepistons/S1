import { Injectable, ForbiddenException, Logger } from '@nestjs/common';
import { DatabaseService } from '../shared/database.service';
import { CacheService } from '../shared/cache.service';
import { ApiMode } from '@prisma/client';

@Injectable()
export class QuotaService {
  private readonly logger = new Logger(QuotaService.name);

  constructor(
    private db: DatabaseService,
    private cacheService: CacheService,
  ) {}

  /**
   * Check if user has quota available
   */
  async checkQuota(userId: string, tool: string): Promise<boolean> {
    // Get user
    const user = await this.db.user.findUnique({
      where: { id: userId },
    });

    if (!user) {
      return false;
    }

    // BYOK and LOCAL modes have no quota limits
    if (user.apiKeyMode === ApiMode.BYOK || user.apiKeyMode === ApiMode.LOCAL) {
      return true;
    }

    // Get quota
    const quota = await this.db.userQuota.findUnique({
      where: {
        userId_tool: {
          userId,
          tool,
        },
      },
    });

    if (!quota) {
      return false;
    }

    // Check if monthly reset is needed
    await this.resetQuotaIfNeeded(quota.id, quota.lastResetAt);

    // Refresh quota
    const updatedQuota = await this.db.userQuota.findUnique({
      where: { id: quota.id },
    });

    // Unlimited quota
    if (updatedQuota!.monthlyLimit === -1) {
      return true;
    }

    // Check if limit reached
    return updatedQuota!.monthlyUsage < updatedQuota!.monthlyLimit;
  }

  /**
   * Increment usage counter
   */
  async incrementUsage(userId: string, tool: string, amount: number = 1): Promise<void> {
    // Get user
    const user = await this.db.user.findUnique({
      where: { id: userId },
    });

    // Don't track for BYOK/LOCAL
    if (user?.apiKeyMode === ApiMode.BYOK || user?.apiKeyMode === ApiMode.LOCAL) {
      return;
    }

    // Increment usage
    await this.db.userQuota.update({
      where: {
        userId_tool: {
          userId,
          tool,
        },
      },
      data: {
        monthlyUsage: {
          increment: amount,
        },
      },
    });

    this.logger.log(`Usage incremented for user ${userId}, tool ${tool}: +${amount}`);
  }

  /**
   * Get user's quota info
   */
  async getQuota(userId: string, tool: string): Promise<any> {
    const user = await this.db.user.findUnique({
      where: { id: userId },
    });

    const quota = await this.db.userQuota.findUnique({
      where: {
        userId_tool: {
          userId,
          tool,
        },
      },
    });

    if (!quota) {
      return null;
    }

    // Check if monthly reset is needed
    await this.resetQuotaIfNeeded(quota.id, quota.lastResetAt);

    // Refresh quota
    const updatedQuota = await this.db.userQuota.findUnique({
      where: { id: quota.id },
    });

    const isUnlimited =
      user?.apiKeyMode === ApiMode.BYOK ||
      user?.apiKeyMode === ApiMode.LOCAL ||
      updatedQuota!.monthlyLimit === -1;

    return {
      tool,
      monthlyLimit: isUnlimited ? -1 : updatedQuota!.monthlyLimit,
      monthlyUsage: updatedQuota!.monthlyUsage,
      remaining: isUnlimited ? -1 : updatedQuota!.monthlyLimit - updatedQuota!.monthlyUsage,
      isUnlimited,
      resetDate: this.getNextResetDate(updatedQuota!.lastResetAt),
    };
  }

  /**
   * Get all quotas for user
   */
  async getAllQuotas(userId: string): Promise<any[]> {
    const tools = ['test-copilot', 'code-review', 'docs'];
    const quotas = [];

    for (const tool of tools) {
      const quota = await this.getQuota(userId, tool);
      if (quota) {
        quotas.push(quota);
      }
    }

    return quotas;
  }

  /**
   * Reset quota if needed (monthly reset)
   */
  private async resetQuotaIfNeeded(quotaId: string, lastResetAt: Date): Promise<void> {
    const now = new Date();
    const lastReset = new Date(lastResetAt);

    // Check if a month has passed
    const monthsPassed =
      (now.getFullYear() - lastReset.getFullYear()) * 12 +
      (now.getMonth() - lastReset.getMonth());

    if (monthsPassed >= 1) {
      await this.db.userQuota.update({
        where: { id: quotaId },
        data: {
          monthlyUsage: 0,
          lastResetAt: now,
        },
      });

      this.logger.log(`Quota reset for quota ID: ${quotaId}`);
    }
  }

  /**
   * Get next reset date
   */
  private getNextResetDate(lastResetAt: Date): Date {
    const nextReset = new Date(lastResetAt);
    nextReset.setMonth(nextReset.getMonth() + 1);
    return nextReset;
  }

  /**
   * Enforce quota check (throws exception if exceeded)
   */
  async enforceQuota(userId: string, tool: string): Promise<void> {
    const hasQuota = await this.checkQuota(userId, tool);

    if (!hasQuota) {
      throw new ForbiddenException('Monthly quota exceeded. Please upgrade your plan.');
    }
  }
}
