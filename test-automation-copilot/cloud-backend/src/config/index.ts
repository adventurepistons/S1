import dotenv from 'dotenv';

dotenv.config();

export const config = {
  // Server
  port: parseInt(process.env.PORT || '3000', 10),
  nodeEnv: process.env.NODE_ENV || 'development',

  // Database
  database: {
    host: process.env.DB_HOST || 'localhost',
    port: parseInt(process.env.DB_PORT || '5432', 10),
    name: process.env.DB_NAME || 'testcopilot',
    user: process.env.DB_USER || 'postgres',
    password: process.env.DB_PASSWORD || '',
    ssl: process.env.DB_SSL === 'true'
  },

  // OpenAI
  openai: {
    apiKey: process.env.OPENAI_API_KEY || '',
    model: process.env.OPENAI_MODEL || 'gpt-4-turbo-preview',
    maxTokens: parseInt(process.env.OPENAI_MAX_TOKENS || '2000', 10),
    temperature: parseFloat(process.env.OPENAI_TEMPERATURE || '0.7')
  },

  // JWT
  jwt: {
    secret: process.env.JWT_SECRET || 'your-secret-key-change-in-production',
    expiresIn: process.env.JWT_EXPIRES_IN || '30d'
  },

  // Rate limiting
  rateLimit: {
    windowMs: parseInt(process.env.RATE_LIMIT_WINDOW || '60000', 10), // 1 minute
    maxRequests: parseInt(process.env.RATE_LIMIT_MAX || '20', 10)
  },

  // Stripe (for payments)
  stripe: {
    secretKey: process.env.STRIPE_SECRET_KEY || '',
    webhookSecret: process.env.STRIPE_WEBHOOK_SECRET || ''
  },

  // Plans
  plans: {
    free: {
      name: 'Free',
      requestsPerMonth: 100,
      tokensPerMonth: 50000,
      price: 0
    },
    pro: {
      name: 'Pro',
      requestsPerMonth: 1000,
      tokensPerMonth: 500000,
      price: 2000 // in cents ($20)
    },
    team: {
      name: 'Team',
      requestsPerMonth: 5000,
      tokensPerMonth: 2500000,
      price: 5000 // in cents ($50)
    }
  }
};

// Validate required config
export function validateConfig() {
  const required = ['OPENAI_API_KEY', 'JWT_SECRET'];

  if (config.nodeEnv === 'production') {
    required.push('DB_PASSWORD', 'STRIPE_SECRET_KEY');
  }

  const missing = required.filter(key => !process.env[key]);

  if (missing.length > 0) {
    throw new Error(`Missing required environment variables: ${missing.join(', ')}`);
  }
}
