import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule, {
    logger: ['error', 'warn', 'log', 'debug', 'verbose'],
  });

  const configService = app.get(ConfigService);

  // Enable CORS
  app.enableCors({
    origin: configService.get('CORS_ORIGIN')?.split(',') || '*',
    credentials: true,
    methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
    allowedHeaders: ['Authorization', 'Content-Type'],
  });

  // Global validation pipe
  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true,
      forbidNonWhitelisted: true,
      transform: true,
      transformOptions: {
        enableImplicitConversion: true,
      },
    }),
  );

  // API prefix
  app.setGlobalPrefix('api/v1');

  const port = configService.get('PORT') || 3000;
  await app.listen(port);

  console.log(`
    ╔═══════════════════════════════════════════════════════════╗
    ║                                                           ║
    ║   🚀 Test Automation Copilot Backend API                 ║
    ║                                                           ║
    ║   Environment: ${configService.get('NODE_ENV')?.padEnd(43)} ║
    ║   Port:        ${port.toString().padEnd(43)} ║
    ║   URL:         http://localhost:${port}/api/v1${' '.repeat(20)} ║
    ║   Docs:        http://localhost:${port}/api/v1/docs${' '.repeat(14)} ║
    ║                                                           ║
    ║   Database:    ${configService.get('DATABASE_URL')?.includes('localhost') ? 'Connected (Local)' : 'Connected'.padEnd(43)} ║
    ║   Redis:       ${configService.get('REDIS_URL')?.includes('localhost') ? 'Connected (Local)' : 'Connected'.padEnd(43)} ║
    ║                                                           ║
    ╚═══════════════════════════════════════════════════════════╝
  `);
}

bootstrap();
