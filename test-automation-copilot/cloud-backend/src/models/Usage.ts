import pool from '../config/database';
import { v4 as uuidv4 } from 'uuid';

export interface UsageRecord {
  id: string;
  user_id: string;
  action: 'pageobject' | 'test' | 'chat' | 'fix';
  tokens_used: number;
  cost: number;
  model: string;
  created_at: Date;
}

export interface UsageStats {
  totalRequests: number;
  totalTokens: number;
  estimatedCost: number;
  requestsByAction: {
    pageobject: number;
    test: number;
    chat: number;
    fix: number;
  };
}

export class UsageModel {
  /**
   * Record usage
   */
  static async record(
    userId: string,
    action: 'pageobject' | 'test' | 'chat' | 'fix',
    tokensUsed: number,
    model: string
  ): Promise<UsageRecord> {
    const id = uuidv4();
    const cost = this.calculateCost(tokensUsed, model);

    const query = `
      INSERT INTO usage (id, user_id, action, tokens_used, cost, model)
      VALUES ($1, $2, $3, $4, $5, $6)
      RETURNING *
    `;

    const result = await pool.query(query, [id, userId, action, tokensUsed, cost, model]);
    return result.rows[0];
  }

  /**
   * Get usage stats for current month
   */
  static async getMonthlyStats(userId: string): Promise<UsageStats> {
    const query = `
      SELECT
        COUNT(*) as total_requests,
        SUM(tokens_used) as total_tokens,
        SUM(cost) as total_cost,
        SUM(CASE WHEN action = 'pageobject' THEN 1 ELSE 0 END) as pageobject_count,
        SUM(CASE WHEN action = 'test' THEN 1 ELSE 0 END) as test_count,
        SUM(CASE WHEN action = 'chat' THEN 1 ELSE 0 END) as chat_count,
        SUM(CASE WHEN action = 'fix' THEN 1 ELSE 0 END) as fix_count
      FROM usage
      WHERE user_id = $1
        AND created_at >= date_trunc('month', CURRENT_DATE)
        AND created_at < date_trunc('month', CURRENT_DATE) + interval '1 month'
    `;

    const result = await pool.query(query, [userId]);
    const row = result.rows[0];

    return {
      totalRequests: parseInt(row.total_requests) || 0,
      totalTokens: parseInt(row.total_tokens) || 0,
      estimatedCost: parseFloat(row.total_cost) || 0,
      requestsByAction: {
        pageobject: parseInt(row.pageobject_count) || 0,
        test: parseInt(row.test_count) || 0,
        chat: parseInt(row.chat_count) || 0,
        fix: parseInt(row.fix_count) || 0
      }
    };
  }

  /**
   * Get usage history
   */
  static async getHistory(userId: string, limit: number = 100): Promise<UsageRecord[]> {
    const query = `
      SELECT * FROM usage
      WHERE user_id = $1
      ORDER BY created_at DESC
      LIMIT $2
    `;

    const result = await pool.query(query, [userId, limit]);
    return result.rows;
  }

  /**
   * Calculate cost based on tokens and model
   */
  private static calculateCost(tokens: number, model: string): number {
    // GPT-4 Turbo pricing: $0.01/1K input tokens, $0.03/1K output tokens
    // Simplified: use average of $0.02/1K tokens
    if (model.includes('gpt-4')) {
      return (tokens / 1000) * 0.02;
    }
    // GPT-3.5 Turbo pricing: $0.0005/1K input, $0.0015/1K output
    // Simplified: use average of $0.001/1K tokens
    return (tokens / 1000) * 0.001;
  }

  /**
   * Check if user exceeded monthly limit
   */
  static async checkLimit(userId: string, limit: number): Promise<boolean> {
    const stats = await this.getMonthlyStats(userId);
    return stats.totalRequests >= limit;
  }
}
