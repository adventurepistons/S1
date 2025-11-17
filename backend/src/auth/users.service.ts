import { Injectable, NotFoundException } from '@nestjs/common';
import { DatabaseService } from '../shared/database.service';
import { User, Tier, ApiMode } from '@prisma/client';

export interface UpdateUserDto {
  displayName?: string;
  avatarUrl?: string;
}

export interface UserWithUsage extends User {
  quotas: Array<{
    tool: string;
    monthlyLimit: number;
    monthlyUsage: number;
    lastResetAt: Date;
  }>;
}

@Injectable()
export class UsersService {
  constructor(private db: DatabaseService) {}

  /**
   * Get user by ID
   */
  async findById(id: string): Promise<User | null> {
    return this.db.user.findUnique({
      where: { id },
    });
  }

  /**
   * Get user by email
   */
  async findByEmail(email: string): Promise<User | null> {
    return this.db.user.findUnique({
      where: { email },
    });
  }

  /**
   * Get user with quotas
   */
  async findByIdWithQuotas(id: string): Promise<UserWithUsage | null> {
    return this.db.user.findUnique({
      where: { id },
      include: {
        quotas: true,
      },
    }) as Promise<UserWithUsage | null>;
  }

  /**
   * Update user profile
   */
  async updateProfile(userId: string, dto: UpdateUserDto): Promise<User> {
    return this.db.user.update({
      where: { id: userId },
      data: dto,
    });
  }

  /**
   * Update user tier
   */
  async updateTier(userId: string, tier: Tier): Promise<User> {
    const user = await this.db.user.update({
      where: { id: userId },
      data: { tier },
    });

    // Update quotas based on tier
    await this.updateQuotasForTier(userId, tier);

    return user;
  }

  /**
   * Update API key mode
   */
  async updateApiKeyMode(userId: string, mode: ApiMode): Promise<User> {
    return this.db.user.update({
      where: { id: userId },
      data: { apiKeyMode: mode },
    });
  }

  /**
   * Delete user account
   */
  async deleteAccount(userId: string): Promise<void> {
    await this.db.user.delete({
      where: { id: userId },
    });
  }

  /**
   * Get user statistics
   */
  async getUserStats(userId: string): Promise<any> {
    const user = await this.findByIdWithQuotas(userId);

    if (!user) {
      throw new NotFoundException('User not found');
    }

    const projects = await this.db.project.count({
      where: { userId },
    });

    const executions = await this.db.testExecution.count({
      where: { userId },
    });

    const conversations = await this.db.conversation.count({
      where: { userId },
    });

    return {
      user: {
        id: user.id,
        email: user.email,
        displayName: user.displayName,
        tier: user.tier,
        apiKeyMode: user.apiKeyMode,
        createdAt: user.createdAt,
      },
      counts: {
        projects,
        executions,
        conversations,
      },
      quotas: user.quotas,
    };
  }

  /**
   * Update quotas based on tier
   */
  private async updateQuotasForTier(userId: string, tier: Tier): Promise<void> {
    let limit = -1; // Unlimited

    if (tier === Tier.PRO) {
      limit = 500; // 500 generations/month for Pro
    } else if (tier === Tier.TEAM) {
      limit = -1; // Unlimited for Team
    }

    // Update all quotas
    await this.db.userQuota.updateMany({
      where: { userId },
      data: { monthlyLimit: limit },
    });
  }
}
