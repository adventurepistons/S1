import pool from '../config/database';
import logger from '../utils/logger';

const migrations = [
  {
    name: '001_create_users_table',
    sql: `
      CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        api_key VARCHAR(255) UNIQUE NOT NULL,
        plan VARCHAR(20) NOT NULL DEFAULT 'free',
        stripe_customer_id VARCHAR(255),
        stripe_subscription_id VARCHAR(255),
        is_active BOOLEAN DEFAULT true,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      );

      CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
      CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);
      CREATE INDEX IF NOT EXISTS idx_users_stripe_customer ON users(stripe_customer_id);
    `
  },
  {
    name: '002_create_usage_table',
    sql: `
      CREATE TABLE IF NOT EXISTS usage (
        id UUID PRIMARY KEY,
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        action VARCHAR(50) NOT NULL,
        tokens_used INTEGER NOT NULL,
        cost DECIMAL(10, 4) NOT NULL,
        model VARCHAR(100) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      );

      CREATE INDEX IF NOT EXISTS idx_usage_user_id ON usage(user_id);
      CREATE INDEX IF NOT EXISTS idx_usage_created_at ON usage(created_at);
      CREATE INDEX IF NOT EXISTS idx_usage_action ON usage(action);
    `
  },
  {
    name: '003_create_migrations_table',
    sql: `
      CREATE TABLE IF NOT EXISTS migrations (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) UNIQUE NOT NULL,
        applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      );
    `
  }
];

async function runMigrations() {
  const client = await pool.connect();

  try {
    logger.info('Starting database migrations...');

    // Create migrations table first
    await client.query(migrations[2].sql);

    // Check which migrations have been applied
    const result = await client.query('SELECT name FROM migrations');
    const appliedMigrations = new Set(result.rows.map((row: { name: string }) => row.name));

    // Run pending migrations
    for (const migration of migrations) {
      if (!appliedMigrations.has(migration.name)) {
        logger.info(`Running migration: ${migration.name}`);
        await client.query('BEGIN');

        try {
          await client.query(migration.sql);
          await client.query('INSERT INTO migrations (name) VALUES ($1)', [migration.name]);
          await client.query('COMMIT');
          logger.info(`✅ Migration ${migration.name} completed`);
        } catch (error) {
          await client.query('ROLLBACK');
          throw error;
        }
      } else {
        logger.info(`⏭️  Migration ${migration.name} already applied`);
      }
    }

    logger.info('✅ All migrations completed successfully');
  } catch (error) {
    logger.error('❌ Migration failed:', error);
    throw error;
  } finally {
    client.release();
  }
}

// Run migrations if this script is executed directly
if (require.main === module) {
  runMigrations()
    .then(() => {
      logger.info('Database migration complete');
      process.exit(0);
    })
    .catch((error) => {
      logger.error('Migration failed:', error);
      process.exit(1);
    });
}

export default runMigrations;
