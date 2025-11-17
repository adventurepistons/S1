import { Injectable, UnauthorizedException, ConflictException, Logger } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { ConfigService } from '@nestjs/config';
import * as bcrypt from 'bcrypt';
import { DatabaseService } from '../shared/database.service';
import { CacheService } from '../shared/cache.service';
import { RegisterDto } from './dto/register.dto';
import { LoginDto } from './dto/login.dto';
import { User, Tier, ApiMode } from '@prisma/client';

export interface JwtPayload {
  sub: string;
  email: string;
  tier: Tier;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: {
    id: string;
    email: string;
    displayName: string | null;
    tier: Tier;
    apiKeyMode: ApiMode;
    emailVerified: boolean;
  };
}

@Injectable()
export class AuthService {
  private readonly logger = new Logger(AuthService.name);
  private readonly saltRounds = 12;

  constructor(
    private db: DatabaseService,
    private jwtService: JwtService,
    private configService: ConfigService,
    private cacheService: CacheService,
  ) {}

  /**
   * Register new user
   */
  async register(dto: RegisterDto): Promise<AuthResponse> {
    // Check if user already exists
    const existing = await this.db.user.findUnique({
      where: { email: dto.email },
    });

    if (existing) {
      throw new ConflictException('User with this email already exists');
    }

    // Hash password
    const passwordHash = await bcrypt.hash(dto.password, this.saltRounds);

    // Create user
    const user = await this.db.user.create({
      data: {
        email: dto.email,
        passwordHash,
        displayName: dto.displayName || dto.email.split('@')[0],
        tier: Tier.FREE,
        apiKeyMode: ApiMode.BYOK,
        emailVerified: false,
      },
    });

    // Create default quotas
    await this.createDefaultQuotas(user.id);

    this.logger.log(`New user registered: ${user.email}`);

    // Generate tokens
    return this.generateTokens(user);
  }

  /**
   * Login user
   */
  async login(dto: LoginDto): Promise<AuthResponse> {
    // Find user
    const user = await this.db.user.findUnique({
      where: { email: dto.email },
    });

    if (!user || !user.passwordHash) {
      throw new UnauthorizedException('Invalid credentials');
    }

    // Check if account is active
    if (!user.isActive) {
      throw new UnauthorizedException('Account is disabled');
    }

    // Verify password
    const isPasswordValid = await bcrypt.compare(dto.password, user.passwordHash);

    if (!isPasswordValid) {
      throw new UnauthorizedException('Invalid credentials');
    }

    // Update last login
    await this.db.user.update({
      where: { id: user.id },
      data: { lastLoginAt: new Date() },
    });

    this.logger.log(`User logged in: ${user.email}`);

    // Generate tokens
    return this.generateTokens(user);
  }

  /**
   * OAuth login (Google, GitHub, etc.)
   */
  async oauthLogin(profile: any, provider: string): Promise<AuthResponse> {
    // Check if OAuth provider exists
    let oauthProvider = await this.db.oAuthProvider.findUnique({
      where: {
        provider_providerUserId: {
          provider,
          providerUserId: profile.id,
        },
      },
      include: { user: true },
    });

    let user: User;

    if (oauthProvider) {
      // Existing user
      user = oauthProvider.user;

      // Update last login
      await this.db.user.update({
        where: { id: user.id },
        data: { lastLoginAt: new Date() },
      });
    } else {
      // New user - create account
      user = await this.db.user.create({
        data: {
          email: profile.email,
          displayName: profile.displayName || profile.name,
          avatarUrl: profile.photo || profile.avatar,
          tier: Tier.FREE,
          apiKeyMode: ApiMode.BYOK,
          emailVerified: true, // OAuth emails are pre-verified
        },
      });

      // Create OAuth provider record
      await this.db.oAuthProvider.create({
        data: {
          userId: user.id,
          provider,
          providerUserId: profile.id,
          accessToken: profile.accessToken,
          refreshToken: profile.refreshToken,
        },
      });

      // Create default quotas
      await this.createDefaultQuotas(user.id);

      this.logger.log(`New user via OAuth (${provider}): ${user.email}`);
    }

    // Generate tokens
    return this.generateTokens(user);
  }

  /**
   * Refresh access token
   */
  async refreshToken(refreshToken: string): Promise<{ accessToken: string; expiresIn: number }> {
    try {
      // Verify refresh token
      const payload = this.jwtService.verify(refreshToken, {
        secret: this.configService.get<string>('JWT_SECRET'),
      });

      // Check if token type is refresh
      if (payload.type !== 'refresh') {
        throw new UnauthorizedException('Invalid token type');
      }

      // Get user
      const user = await this.db.user.findUnique({
        where: { id: payload.sub },
      });

      if (!user || !user.isActive) {
        throw new UnauthorizedException('User not found or inactive');
      }

      // Generate new access token
      const accessToken = await this.generateAccessToken(user);

      return {
        accessToken,
        expiresIn: parseInt(this.configService.get<string>('JWT_EXPIRES_IN') || '3600'),
      };
    } catch (error) {
      throw new UnauthorizedException('Invalid refresh token');
    }
  }

  /**
   * Validate user by ID (used by JWT strategy)
   */
  async validateUser(userId: string): Promise<User | null> {
    // Try cache first
    const cacheKey = `user:${userId}`;
    const cached = await this.cacheService.get<User>(cacheKey);

    if (cached) {
      return cached;
    }

    // Get from database
    const user = await this.db.user.findUnique({
      where: { id: userId },
    });

    if (user && user.isActive) {
      // Cache for 1 hour
      await this.cacheService.set(cacheKey, user, 3600);
      return user;
    }

    return null;
  }

  /**
   * Logout user (invalidate tokens)
   */
  async logout(userId: string): Promise<void> {
    // Clear user cache
    await this.cacheService.delete(`user:${userId}`);

    this.logger.log(`User logged out: ${userId}`);
  }

  /**
   * Generate JWT tokens
   */
  private async generateTokens(user: User): Promise<AuthResponse> {
    const accessToken = await this.generateAccessToken(user);
    const refreshToken = await this.generateRefreshToken(user);

    return {
      accessToken,
      refreshToken,
      expiresIn: parseInt(this.configService.get<string>('JWT_EXPIRES_IN') || '3600'),
      user: {
        id: user.id,
        email: user.email,
        displayName: user.displayName,
        tier: user.tier,
        apiKeyMode: user.apiKeyMode,
        emailVerified: user.emailVerified,
      },
    };
  }

  /**
   * Generate access token
   */
  private async generateAccessToken(user: User): Promise<string> {
    const payload: JwtPayload = {
      sub: user.id,
      email: user.email,
      tier: user.tier,
    };

    return this.jwtService.signAsync(payload);
  }

  /**
   * Generate refresh token
   */
  private async generateRefreshToken(user: User): Promise<string> {
    const payload = {
      sub: user.id,
      type: 'refresh',
    };

    return this.jwtService.signAsync(payload, {
      expiresIn: this.configService.get<string>('REFRESH_TOKEN_EXPIRES_IN') || '30d',
    });
  }

  /**
   * Create default quotas for new user
   */
  private async createDefaultQuotas(userId: string): Promise<void> {
    const tools = ['test-copilot', 'code-review', 'docs'];

    for (const tool of tools) {
      await this.db.userQuota.create({
        data: {
          userId,
          tool,
          monthlyLimit: -1, // Unlimited for BYOK
          monthlyUsage: 0,
        },
      });
    }
  }
}
