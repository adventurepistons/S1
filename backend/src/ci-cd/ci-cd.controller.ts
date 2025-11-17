import { Controller, Post, Body, Headers, BadRequestException, Logger } from '@nestjs/common';
import { WebhookService, WebhookPayload } from './webhook.service';
import { NotificationService, TestExecutionNotification } from './notification.service';

@Controller('webhooks')
export class CiCdController {
  private readonly logger = new Logger(CiCdController.name);

  constructor(
    private webhookService: WebhookService,
    private notificationService: NotificationService,
  ) {}

  /**
   * GitHub webhook endpoint
   * POST /webhooks/github
   */
  @Post('github')
  async handleGithubWebhook(
    @Body() payload: WebhookPayload,
    @Headers('x-github-event') event: string,
    @Headers('x-hub-signature-256') signature: string,
  ): Promise<any> {
    this.logger.log(`Received GitHub webhook: ${event}`);

    // Get webhook secret from environment
    const secret = process.env.GITHUB_WEBHOOK_SECRET;
    if (!secret) {
      throw new BadRequestException('GitHub webhook secret not configured');
    }

    // Add event to payload
    payload.event = event;

    // Handle webhook
    return this.webhookService.handleGithubWebhook(payload, signature, secret);
  }

  /**
   * GitLab webhook endpoint
   * POST /webhooks/gitlab
   */
  @Post('gitlab')
  async handleGitlabWebhook(
    @Body() payload: any,
    @Headers('x-gitlab-event') event: string,
    @Headers('x-gitlab-token') token: string,
  ): Promise<any> {
    this.logger.log(`Received GitLab webhook: ${event}`);

    const secret = process.env.GITLAB_WEBHOOK_SECRET;
    if (!secret || token !== secret) {
      throw new BadRequestException('Invalid token');
    }

    // TODO: Implement GitLab webhook handling
    return { message: 'GitLab webhook received' };
  }

  /**
   * Bitbucket webhook endpoint
   * POST /webhooks/bitbucket
   */
  @Post('bitbucket')
  async handleBitbucketWebhook(
    @Body() payload: any,
    @Headers('x-event-key') event: string,
  ): Promise<any> {
    this.logger.log(`Received Bitbucket webhook: ${event}`);

    // TODO: Implement Bitbucket webhook handling
    return { message: 'Bitbucket webhook received' };
  }

  /**
   * Slack webhook endpoint (for notifications)
   * POST /webhooks/slack
   */
  @Post('slack')
  async handleSlackWebhook(@Body() payload: any): Promise<any> {
    this.logger.log('Received Slack webhook');

    // Handle Slack slash commands or interactive messages
    // TODO: Implement Slack integration
    return { message: 'Slack webhook received' };
  }

  /**
   * Generic notification webhook
   * POST /webhooks/notify
   */
  @Post('notify')
  async handleNotificationWebhook(@Body() payload: {
    type: 'slack' | 'teams' | 'discord' | 'webhook';
    webhookUrl: string;
    notification: TestExecutionNotification;
  }): Promise<any> {
    this.logger.log(`Sending ${payload.type} notification`);

    try {
      switch (payload.type) {
        case 'slack':
          await this.notificationService.sendSlackNotification(
            payload.webhookUrl,
            payload.notification,
          );
          break;
        case 'teams':
          await this.notificationService.sendTeamsNotification(
            payload.webhookUrl,
            payload.notification,
          );
          break;
        case 'discord':
          await this.notificationService.sendDiscordNotification(
            payload.webhookUrl,
            payload.notification,
          );
          break;
        case 'webhook':
          await this.notificationService.sendWebhookNotification(
            payload.webhookUrl,
            payload.notification,
          );
          break;
      }

      return { message: 'Notification sent successfully' };
    } catch (error: any) {
      this.logger.error(`Failed to send notification: ${error.message}`);
      throw new BadRequestException('Failed to send notification');
    }
  }

  /**
   * Test webhook endpoint (for testing webhook configuration)
   * POST /webhooks/test
   */
  @Post('test')
  async testWebhook(@Body() payload: any): Promise<any> {
    this.logger.log('Received test webhook');
    return {
      message: 'Webhook received successfully',
      timestamp: new Date().toISOString(),
      payload,
    };
  }
}
