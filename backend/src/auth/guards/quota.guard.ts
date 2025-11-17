import { Injectable, CanActivate, ExecutionContext, ForbiddenException } from '@nestjs/common';
import { Reflector } from '@nestjs/core';
import { QuotaService } from '../quota.service';

@Injectable()
export class QuotaGuard implements CanActivate {
  constructor(
    private quotaService: QuotaService,
    private reflector: Reflector,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    // Get tool name from decorator
    const tool = this.reflector.get<string>('tool', context.getHandler());

    if (!tool) {
      // No tool specified, skip quota check
      return true;
    }

    const request = context.switchToHttp().getRequest();
    const user = request.user;

    if (!user) {
      throw new ForbiddenException('User not authenticated');
    }

    // Check quota
    const hasQuota = await this.quotaService.checkQuota(user.id, tool);

    if (!hasQuota) {
      throw new ForbiddenException(
        `Monthly quota exceeded for ${tool}. Please upgrade your plan or use your own API key (BYOK mode).`,
      );
    }

    return true;
  }
}
