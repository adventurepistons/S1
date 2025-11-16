import { Router, Request, Response } from 'express';
import { authenticate } from '../middleware/auth';
import { UsageModel } from '../models/Usage';
import { config } from '../config';

const router = Router();

/**
 * GET /v1/usage
 * Get usage statistics for authenticated user
 */
router.get('/usage', authenticate, async (req: Request, res: Response) => {
  try {
    const userId = req.userId!;
    const user = req.user;

    // Get monthly stats
    const stats = await UsageModel.getMonthlyStats(userId);

    // Get plan limits
    const planLimits = config.plans[user.plan];

    // Calculate period dates
    const now = new Date();
    const monthStart = new Date(now.getFullYear(), now.getMonth(), 1);
    const monthEnd = new Date(now.getFullYear(), now.getMonth() + 1, 0);

    res.json({
      userId,
      plan: user.plan,
      period: {
        start: Math.floor(monthStart.getTime() / 1000),
        end: Math.floor(monthEnd.getTime() / 1000),
        current: Math.floor(now.getTime() / 1000)
      },
      usage: {
        totalRequests: stats.totalRequests,
        totalTokens: stats.totalTokens,
        estimatedCost: stats.estimatedCost,
        requestsThisMonth: stats.totalRequests,
        tokensThisMonth: stats.totalTokens,
        requestsByAction: stats.requestsByAction
      },
      limits: {
        requestsPerMonth: planLimits.requestsPerMonth,
        tokensPerMonth: planLimits.tokensPerMonth,
        requestsRemaining: Math.max(0, planLimits.requestsPerMonth - stats.totalRequests),
        tokensRemaining: Math.max(0, planLimits.tokensPerMonth - stats.totalTokens)
      }
    });
  } catch (error: any) {
    res.status(500).json({
      error: `Failed to get usage stats: ${error.message}`
    });
  }
});

/**
 * GET /v1/usage/history
 * Get usage history
 */
router.get('/usage/history', authenticate, async (req: Request, res: Response) => {
  try {
    const userId = req.userId!;
    const limit = Math.min(parseInt(req.query.limit as string) || 100, 1000);

    const history = await UsageModel.getHistory(userId, limit);

    res.json({
      history,
      count: history.length
    });
  } catch (error: any) {
    res.status(500).json({
      error: `Failed to get usage history: ${error.message}`
    });
  }
});

export default router;
