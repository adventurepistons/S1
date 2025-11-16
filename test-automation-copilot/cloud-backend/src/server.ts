import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import rateLimit from 'express-rate-limit';
import { config, validateConfig } from './config';
import logger from './utils/logger';
import healthRouter from './routes/health';
import generateRouter from './routes/generate';
import usageRouter from './routes/usage';
import runMigrations from './scripts/migrate';

// Validate configuration
try {
  validateConfig();
} catch (error: any) {
  logger.error('Configuration error:', error);
  process.exit(1);
}

// Create Express app
const app = express();

// Security middleware
app.use(helmet());

// CORS
app.use(cors({
  origin: process.env.CORS_ORIGIN || '*',
  credentials: true
}));

// Body parser
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true, limit: '10mb' }));

// Global rate limiting (per IP)
const globalLimiter = rateLimit({
  windowMs: config.rateLimit.windowMs,
  max: config.rateLimit.maxRequests,
  message: { error: 'Too many requests. Please try again later.' },
  standardHeaders: true,
  legacyHeaders: false
});
app.use('/v1/', globalLimiter);

// Request logging
app.use((req, res, next) => {
  logger.info(`${req.method} ${req.path}`, {
    ip: req.ip,
    userAgent: req.get('user-agent')
  });
  next();
});

// Routes
app.use('/', healthRouter);
app.use('/v1', generateRouter);
app.use('/v1', usageRouter);

// 404 handler
app.use((req, res) => {
  res.status(404).json({
    error: 'Endpoint not found',
    path: req.path,
    method: req.method
  });
});

// Error handler
app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
  logger.error('Server error:', {
    error: err.message,
    stack: err.stack,
    path: req.path
  });

  res.status(err.status || 500).json({
    error: config.nodeEnv === 'production'
      ? 'Internal server error'
      : err.message,
    ...(config.nodeEnv === 'development' && { stack: err.stack })
  });
});

// Initialize server
async function startServer() {
  try {
    // Run database migrations
    logger.info('Running database migrations...');
    await runMigrations();

    // Start listening
    const port = config.port;
    app.listen(port, () => {
      logger.info('='.repeat(60));
      logger.info('🚀 Test Copilot Cloud Backend');
      logger.info('='.repeat(60));
      logger.info(`📡 Server:      http://localhost:${port}`);
      logger.info(`🌍 Environment: ${config.nodeEnv}`);
      logger.info(`📊 Database:    ${config.database.host}:${config.database.port}/${config.database.name}`);
      logger.info(`🤖 OpenAI:      ${config.openai.apiKey ? 'Configured ✅' : 'Not configured ❌'}`);
      logger.info(`💳 Stripe:      ${config.stripe.secretKey ? 'Configured ✅' : 'Not configured ❌'}`);
      logger.info('='.repeat(60));
      logger.info('📝 Available Endpoints:');
      logger.info('   GET  /health');
      logger.info('   POST /v1/generate');
      logger.info('   POST /v1/generate/stream');
      logger.info('   GET  /v1/usage');
      logger.info('   GET  /v1/usage/history');
      logger.info('='.repeat(60));
    });
  } catch (error) {
    logger.error('Failed to start server:', error);
    process.exit(1);
  }
}

// Handle graceful shutdown
process.on('SIGTERM', () => {
  logger.info('SIGTERM received, shutting down gracefully...');
  process.exit(0);
});

process.on('SIGINT', () => {
  logger.info('SIGINT received, shutting down gracefully...');
  process.exit(0);
});

// Start the server
startServer();

export default app;
