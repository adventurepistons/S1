import { Injectable, Logger } from '@nestjs/common';
import axios from 'axios';

export interface TestExecutionNotification {
  executionId: string;
  projectName: string;
  status: 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'CANCELLED';
  totalTests: number;
  passedTests: number;
  failedTests: number;
  skippedTests: number;
  duration?: number;
  reportUrl: string;
  triggeredBy?: string;
  environment?: string;
  browser?: string;
}

@Injectable()
export class NotificationService {
  private readonly logger = new Logger(NotificationService.name);

  /**
   * Send notification to Slack
   */
  async sendSlackNotification(webhookUrl: string, notification: TestExecutionNotification): Promise<void> {
    try {
      const color = this.getStatusColor(notification.status);
      const emoji = this.getStatusEmoji(notification.status);

      const blocks = [
        {
          type: 'header',
          text: {
            type: 'plain_text',
            text: `${emoji} Test Execution ${notification.status}`,
          },
        },
        {
          type: 'section',
          fields: [
            {
              type: 'mrkdwn',
              text: `*Project:*\n${notification.projectName}`,
            },
            {
              type: 'mrkdwn',
              text: `*Status:*\n${notification.status}`,
            },
            {
              type: 'mrkdwn',
              text: `*Environment:*\n${notification.environment || 'N/A'}`,
            },
            {
              type: 'mrkdwn',
              text: `*Browser:*\n${notification.browser || 'N/A'}`,
            },
          ],
        },
        {
          type: 'section',
          fields: [
            {
              type: 'mrkdwn',
              text: `*Total Tests:*\n${notification.totalTests}`,
            },
            {
              type: 'mrkdwn',
              text: `*Passed:*\n✅ ${notification.passedTests}`,
            },
            {
              type: 'mrkdwn',
              text: `*Failed:*\n❌ ${notification.failedTests}`,
            },
            {
              type: 'mrkdwn',
              text: `*Skipped:*\n⏭️ ${notification.skippedTests}`,
            },
          ],
        },
      ];

      if (notification.duration) {
        blocks.push({
          type: 'section',
          fields: [
            {
              type: 'mrkdwn',
              text: `*Duration:*\n${this.formatDuration(notification.duration)}`,
            },
            {
              type: 'mrkdwn',
              text: `*Triggered By:*\n${notification.triggeredBy || 'System'}`,
            },
          ],
        });
      }

      blocks.push({
        type: 'actions',
        elements: [
          {
            type: 'button',
            text: {
              type: 'plain_text',
              text: '📊 View Report',
            },
            url: notification.reportUrl,
            style: 'primary',
          },
        ],
      });

      await axios.post(webhookUrl, {
        attachments: [
          {
            color,
            blocks,
          },
        ],
      });

      this.logger.log(`Slack notification sent for execution ${notification.executionId}`);
    } catch (error: any) {
      this.logger.error(`Failed to send Slack notification: ${error.message}`);
    }
  }

  /**
   * Send notification to Microsoft Teams
   */
  async sendTeamsNotification(webhookUrl: string, notification: TestExecutionNotification): Promise<void> {
    try {
      const emoji = this.getStatusEmoji(notification.status);
      const themeColor = this.getStatusColorHex(notification.status);

      const card = {
        '@type': 'MessageCard',
        '@context': 'https://schema.org/extensions',
        themeColor,
        summary: `Test Execution ${notification.status}`,
        sections: [
          {
            activityTitle: `${emoji} Test Execution ${notification.status}`,
            activitySubtitle: notification.projectName,
            facts: [
              {
                name: 'Status:',
                value: notification.status,
              },
              {
                name: 'Environment:',
                value: notification.environment || 'N/A',
              },
              {
                name: 'Browser:',
                value: notification.browser || 'N/A',
              },
              {
                name: 'Total Tests:',
                value: notification.totalTests.toString(),
              },
              {
                name: 'Passed:',
                value: `✅ ${notification.passedTests}`,
              },
              {
                name: 'Failed:',
                value: `❌ ${notification.failedTests}`,
              },
              {
                name: 'Skipped:',
                value: `⏭️ ${notification.skippedTests}`,
              },
            ],
          },
        ],
        potentialAction: [
          {
            '@type': 'OpenUri',
            name: '📊 View Report',
            targets: [
              {
                os: 'default',
                uri: notification.reportUrl,
              },
            ],
          },
        ],
      };

      if (notification.duration) {
        card.sections[0].facts.push({
          name: 'Duration:',
          value: this.formatDuration(notification.duration),
        });
      }

      if (notification.triggeredBy) {
        card.sections[0].facts.push({
          name: 'Triggered By:',
          value: notification.triggeredBy,
        });
      }

      await axios.post(webhookUrl, card);

      this.logger.log(`Teams notification sent for execution ${notification.executionId}`);
    } catch (error: any) {
      this.logger.error(`Failed to send Teams notification: ${error.message}`);
    }
  }

  /**
   * Send notification to Discord
   */
  async sendDiscordNotification(webhookUrl: string, notification: TestExecutionNotification): Promise<void> {
    try {
      const color = this.getStatusColorDecimal(notification.status);
      const emoji = this.getStatusEmoji(notification.status);

      const embed = {
        title: `${emoji} Test Execution ${notification.status}`,
        description: `**Project:** ${notification.projectName}`,
        color,
        fields: [
          {
            name: 'Status',
            value: notification.status,
            inline: true,
          },
          {
            name: 'Environment',
            value: notification.environment || 'N/A',
            inline: true,
          },
          {
            name: 'Browser',
            value: notification.browser || 'N/A',
            inline: true,
          },
          {
            name: 'Total Tests',
            value: notification.totalTests.toString(),
            inline: true,
          },
          {
            name: 'Passed',
            value: `✅ ${notification.passedTests}`,
            inline: true,
          },
          {
            name: 'Failed',
            value: `❌ ${notification.failedTests}`,
            inline: true,
          },
          {
            name: 'Skipped',
            value: `⏭️ ${notification.skippedTests}`,
            inline: true,
          },
        ],
        timestamp: new Date().toISOString(),
        footer: {
          text: 'Test Copilot',
        },
      };

      if (notification.duration) {
        embed.fields.push({
          name: 'Duration',
          value: this.formatDuration(notification.duration),
          inline: true,
        });
      }

      if (notification.triggeredBy) {
        embed.fields.push({
          name: 'Triggered By',
          value: notification.triggeredBy,
          inline: true,
        });
      }

      await axios.post(webhookUrl, {
        embeds: [embed],
        components: [
          {
            type: 1,
            components: [
              {
                type: 2,
                style: 5,
                label: '📊 View Report',
                url: notification.reportUrl,
              },
            ],
          },
        ],
      });

      this.logger.log(`Discord notification sent for execution ${notification.executionId}`);
    } catch (error: any) {
      this.logger.error(`Failed to send Discord notification: ${error.message}`);
    }
  }

  /**
   * Send generic webhook notification
   */
  async sendWebhookNotification(webhookUrl: string, notification: TestExecutionNotification): Promise<void> {
    try {
      await axios.post(webhookUrl, notification, {
        headers: {
          'Content-Type': 'application/json',
          'User-Agent': 'TestCopilot-Webhook',
        },
      });

      this.logger.log(`Webhook notification sent for execution ${notification.executionId}`);
    } catch (error: any) {
      this.logger.error(`Failed to send webhook notification: ${error.message}`);
    }
  }

  /**
   * Send email notification (placeholder - requires email service integration)
   */
  async sendEmailNotification(to: string, notification: TestExecutionNotification): Promise<void> {
    // TODO: Integrate with SendGrid, AWS SES, or similar
    this.logger.log(`Email notification to ${to} - Not implemented yet`);
  }

  /**
   * Get status color for Slack
   */
  private getStatusColor(status: string): string {
    switch (status) {
      case 'COMPLETED':
        return 'good';
      case 'FAILED':
        return 'danger';
      case 'RUNNING':
        return 'warning';
      default:
        return '#808080';
    }
  }

  /**
   * Get status color hex
   */
  private getStatusColorHex(status: string): string {
    switch (status) {
      case 'COMPLETED':
        return '00FF00';
      case 'FAILED':
        return 'FF0000';
      case 'RUNNING':
        return 'FFA500';
      case 'QUEUED':
        return '0000FF';
      default:
        return '808080';
    }
  }

  /**
   * Get status color as decimal (for Discord)
   */
  private getStatusColorDecimal(status: string): number {
    switch (status) {
      case 'COMPLETED':
        return 0x00ff00;
      case 'FAILED':
        return 0xff0000;
      case 'RUNNING':
        return 0xffa500;
      case 'QUEUED':
        return 0x0000ff;
      default:
        return 0x808080;
    }
  }

  /**
   * Get status emoji
   */
  private getStatusEmoji(status: string): string {
    switch (status) {
      case 'COMPLETED':
        return '✅';
      case 'FAILED':
        return '❌';
      case 'RUNNING':
        return '🔄';
      case 'QUEUED':
        return '⏳';
      case 'CANCELLED':
        return '🚫';
      default:
        return '📊';
    }
  }

  /**
   * Format duration in human-readable format
   */
  private formatDuration(seconds: number): string {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);

    if (hours > 0) {
      return `${hours}h ${minutes}m ${secs}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${secs}s`;
    } else {
      return `${secs}s`;
    }
  }
}
