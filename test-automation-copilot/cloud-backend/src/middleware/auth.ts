import { Request, Response, NextFunction } from 'express';
import { UserModel, User } from '../models/User';
import { UsageModel } from '../models/Usage';
import { config } from '../config';
import logger from '../utils/logger';

// Extend Express Request type
declare global {
  namespace Express {
    interface Request {
      user?: User;
      userId?: string;
    }
  }
}

/**
 * Authenticate user by API key
 */
export async function authenticate(req: Request, res: Response, next: NextFunction) {
  try {
    const authHeader = req.headers.authorization;

    if (!authHeader) {
      return res.status(401).json({ error: 'Missing authorization header' });
    }

    const apiKey = authHeader.replace('Bearer ', '').trim();

    if (!apiKey) {
      return res.status(401).json({ error: 'Invalid authorization header format' });
    }

    // Find user by API key
    const user = await UserModel.findByApiKey(apiKey);

    if (!user) {
      logger.warn(`Invalid API key attempt: ${apiKey.substring(0, 10)}...`);
      return res.status(401).json({ error: 'Invalid API key' });
    }

    // Check if user is active
    if (!user.is_active) {
      return res.status(403).json({ error: 'Account is deactivated' });
    }

    // Get plan limits
    const planLimits = config.plans[user.plan];

    // Check monthly usage limit
    const hasExceeded = await UsageModel.checkLimit(user.id, planLimits.requestsPerMonth);

    if (hasExceeded) {
      return res.status(429).json({
        error: 'Monthly request limit exceeded. Please upgrade your plan or wait for next month.',
        retryAfter: getSecondsUntilNextMonth()
      });
    }

    // Attach user to request
    req.user = user;
    req.userId = user.id;

    logger.debug(`Authenticated user: ${user.email} (${user.plan})`);

    next();
  } catch (error) {
    logger.error('Authentication error:', error);
    res.status(500).json({ error: 'Authentication failed' });
  }
}

/**
 * Get seconds until next month
 */
function getSecondsUntilNextMonth(): number {
  const now = new Date();
  const nextMonth = new Date(now.getFullYear(), now.getMonth() + 1, 1);
  return Math.floor((nextMonth.getTime() - now.getTime()) / 1000);
}
