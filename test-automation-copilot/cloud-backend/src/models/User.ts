import pool from '../config/database';
import bcrypt from 'bcrypt';
import { v4 as uuidv4 } from 'uuid';
import crypto from 'crypto';

export interface User {
  id: string;
  email: string;
  password_hash: string;
  api_key: string;
  plan: 'free' | 'pro' | 'team';
  stripe_customer_id?: string;
  stripe_subscription_id?: string;
  is_active: boolean;
  created_at: Date;
  updated_at: Date;
}

export class UserModel {
  /**
   * Create a new user
   */
  static async create(email: string, password: string, plan: 'free' | 'pro' | 'team' = 'free'): Promise<User> {
    const id = uuidv4();
    const password_hash = await bcrypt.hash(password, 10);
    const api_key = `tc_${plan}_${crypto.randomBytes(32).toString('hex')}`;

    const query = `
      INSERT INTO users (id, email, password_hash, api_key, plan)
      VALUES ($1, $2, $3, $4, $5)
      RETURNING *
    `;

    const result = await pool.query(query, [id, email, password_hash, api_key, plan]);
    return result.rows[0];
  }

  /**
   * Find user by email
   */
  static async findByEmail(email: string): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE email = $1';
    const result = await pool.query(query, [email]);
    return result.rows[0] || null;
  }

  /**
   * Find user by API key
   */
  static async findByApiKey(apiKey: string): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE api_key = $1 AND is_active = true';
    const result = await pool.query(query, [apiKey]);
    return result.rows[0] || null;
  }

  /**
   * Find user by ID
   */
  static async findById(id: string): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE id = $1';
    const result = await pool.query(query, [id]);
    return result.rows[0] || null;
  }

  /**
   * Verify password
   */
  static async verifyPassword(password: string, hash: string): Promise<boolean> {
    return bcrypt.compare(password, hash);
  }

  /**
   * Update user plan
   */
  static async updatePlan(userId: string, plan: 'free' | 'pro' | 'team'): Promise<void> {
    const query = `
      UPDATE users
      SET plan = $1, updated_at = NOW()
      WHERE id = $2
    `;
    await pool.query(query, [plan, userId]);
  }

  /**
   * Update Stripe customer ID
   */
  static async updateStripeCustomer(userId: string, customerId: string, subscriptionId?: string): Promise<void> {
    const query = `
      UPDATE users
      SET stripe_customer_id = $1, stripe_subscription_id = $2, updated_at = NOW()
      WHERE id = $3
    `;
    await pool.query(query, [customerId, subscriptionId, userId]);
  }

  /**
   * Regenerate API key
   */
  static async regenerateApiKey(userId: string): Promise<string> {
    const user = await this.findById(userId);
    if (!user) throw new Error('User not found');

    const api_key = `tc_${user.plan}_${crypto.randomBytes(32).toString('hex')}`;
    const query = 'UPDATE users SET api_key = $1, updated_at = NOW() WHERE id = $2';
    await pool.query(query, [api_key, userId]);

    return api_key;
  }

  /**
   * Deactivate user
   */
  static async deactivate(userId: string): Promise<void> {
    const query = 'UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1';
    await pool.query(query, [userId]);
  }
}
