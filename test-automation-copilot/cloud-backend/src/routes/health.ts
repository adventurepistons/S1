import { Router, Request, Response } from 'express';
import pool from '../config/database';
import openAIService from '../services/OpenAIService';

const router = Router();

/**
 * GET /health
 * Health check endpoint
 */
router.get('/health', async (req: Request, res: Response) => {
  try {
    // Check database
    const dbCheck = await pool.query('SELECT NOW()');
    const dbHealthy = dbCheck.rowCount !== null && dbCheck.rowCount > 0;

    // Check OpenAI
    const openAIHealthy = openAIService.isConfigured();

    const status = dbHealthy && openAIHealthy ? 'healthy' : 'degraded';

    res.status(dbHealthy && openAIHealthy ? 200 : 503).json({
      status,
      timestamp: Math.floor(Date.now() / 1000),
      version: '1.0.0',
      services: {
        database: dbHealthy,
        openai: openAIHealthy
      }
    });
  } catch (error: any) {
    res.status(503).json({
      status: 'unhealthy',
      timestamp: Math.floor(Date.now() / 1000),
      error: error.message
    });
  }
});

export default router;
