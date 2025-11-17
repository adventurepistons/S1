import {
  Controller,
  Post,
  Get,
  Put,
  Body,
  UseGuards,
  Req,
  Res,
  HttpCode,
  HttpStatus,
} from '@nestjs/common';
import { AuthGuard } from '@nestjs/passport';
import { JwtAuthGuard } from './guards/jwt-auth.guard';
import { AuthService, AuthResponse } from './auth.service';
import { UsersService } from './users.service';
import { QuotaService } from './quota.service';
import { RegisterDto } from './dto/register.dto';
import { LoginDto } from './dto/login.dto';
import { CurrentUser } from './decorators/current-user.decorator';
import { Public } from './decorators/public.decorator';
import { User, Tier, ApiMode } from '@prisma/client';

@Controller('auth')
export class AuthController {
  constructor(
    private authService: AuthService,
    private usersService: UsersService,
    private quotaService: QuotaService,
  ) {}

  /**
   * Register new user
   * POST /api/v1/auth/register
   */
  @Public()
  @Post('register')
  @HttpCode(HttpStatus.CREATED)
  async register(@Body() dto: RegisterDto): Promise<AuthResponse> {
    return this.authService.register(dto);
  }

  /**
   * Login with email/password
   * POST /api/v1/auth/login
   */
  @Public()
  @Post('login')
  @HttpCode(HttpStatus.OK)
  async login(@Body() dto: LoginDto): Promise<AuthResponse> {
    return this.authService.login(dto);
  }

  /**
   * Refresh access token
   * POST /api/v1/auth/refresh
   */
  @Public()
  @Post('refresh')
  @HttpCode(HttpStatus.OK)
  async refreshToken(
    @Body('refreshToken') refreshToken: string,
  ): Promise<{ accessToken: string; expiresIn: number }> {
    return this.authService.refreshToken(refreshToken);
  }

  /**
   * Logout
   * POST /api/v1/auth/logout
   */
  @Post('logout')
  @HttpCode(HttpStatus.OK)
  @UseGuards(JwtAuthGuard)
  async logout(@CurrentUser('id') userId: string): Promise<{ message: string }> {
    await this.authService.logout(userId);
    return { message: 'Logged out successfully' };
  }

  /**
   * Google OAuth - Initiate
   * GET /api/v1/auth/google
   */
  @Public()
  @Get('google')
  @UseGuards(AuthGuard('google'))
  async googleAuth() {
    // Redirects to Google
  }

  /**
   * Google OAuth - Callback
   * GET /api/v1/auth/google/callback
   */
  @Public()
  @Get('google/callback')
  @UseGuards(AuthGuard('google'))
  async googleAuthCallback(@Req() req: any, @Res() res: any) {
    const authResponse = await this.authService.oauthLogin(req.user, 'google');

    // Redirect to frontend with tokens
    const frontendUrl = process.env.FRONTEND_URL || 'http://localhost:5173';
    const redirectUrl = `${frontendUrl}/auth/callback?accessToken=${authResponse.accessToken}&refreshToken=${authResponse.refreshToken}`;

    return res.redirect(redirectUrl);
  }

  /**
   * Get current user profile
   * GET /api/v1/auth/me
   */
  @Get('me')
  @UseGuards(JwtAuthGuard)
  async getProfile(@CurrentUser() user: User) {
    const stats = await this.usersService.getUserStats(user.id);
    return stats;
  }

  /**
   * Update user profile
   * PUT /api/v1/auth/me
   */
  @Put('me')
  @UseGuards(JwtAuthGuard)
  async updateProfile(
    @CurrentUser('id') userId: string,
    @Body() dto: { displayName?: string; avatarUrl?: string },
  ) {
    const user = await this.usersService.updateProfile(userId, dto);

    return {
      id: user.id,
      email: user.email,
      displayName: user.displayName,
      avatarUrl: user.avatarUrl,
      tier: user.tier,
      apiKeyMode: user.apiKeyMode,
    };
  }

  /**
   * Get user usage and quotas
   * GET /api/v1/auth/usage
   */
  @Get('usage')
  @UseGuards(JwtAuthGuard)
  async getUsage(@CurrentUser('id') userId: string) {
    const quotas = await this.quotaService.getAllQuotas(userId);
    return { quotas };
  }

  /**
   * Update API key mode
   * PUT /api/v1/auth/api-mode
   */
  @Put('api-mode')
  @UseGuards(JwtAuthGuard)
  async updateApiMode(
    @CurrentUser('id') userId: string,
    @Body('mode') mode: ApiMode,
  ) {
    const user = await this.usersService.updateApiKeyMode(userId, mode);

    return {
      apiKeyMode: user.apiKeyMode,
      message: `API mode updated to ${mode}`,
    };
  }
}
