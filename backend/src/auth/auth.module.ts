import { Module } from '@nestjs/common';
import { JwtModule } from '@nestjs/jwt';
import { PassportModule } from '@nestjs/passport';
import { ConfigService } from '@nestjs/config';
import { AuthController } from './auth.controller';
import { AuthService } from './auth.service';
import { UsersService } from './users.service';
import { QuotaService } from './quota.service';
import { JwtStrategy } from './strategies/jwt.strategy';
import { GoogleStrategy } from './strategies/google.strategy';

@Module({
  imports: [
    PassportModule.register({ defaultStrategy: 'jwt' }),
    JwtModule.registerAsync({
      inject: [ConfigService],
      useFactory: (configService: ConfigService) => {
        const secret = configService.get<string>('JWT_SECRET');
        const nodeEnv = configService.get<string>('NODE_ENV');

        // Validate JWT secret
        if (!secret) {
          if (nodeEnv === 'production') {
            throw new Error('❌ JWT_SECRET environment variable must be set in production!');
          }
          // Development only: generate random secret
          const crypto = require('crypto');
          const devSecret = crypto.randomBytes(64).toString('hex');
          console.warn('⚠️  Using auto-generated JWT secret - FOR DEVELOPMENT ONLY!');
          console.warn('⚠️  Set JWT_SECRET in production: openssl rand -base64 64');

          return {
            secret: devSecret,
            signOptions: {
              expiresIn: configService.get<string>('JWT_EXPIRES_IN') || '3600s',
            },
          };
        }

        if (secret.length < 32) {
          throw new Error('❌ JWT_SECRET must be at least 32 characters long for security');
        }

        return {
          secret,
          signOptions: {
            expiresIn: configService.get<string>('JWT_EXPIRES_IN') || '3600s',
          },
        };
      },
    }),
  ],
  controllers: [AuthController],
  providers: [
    AuthService,
    UsersService,
    QuotaService,
    JwtStrategy,
    GoogleStrategy,
  ],
  exports: [AuthService, UsersService, QuotaService],
})
export class AuthModule {}
