import OpenAI from 'openai';
import { config } from '../config';
import logger from '../utils/logger';

export class OpenAIService {
  private client: OpenAI;

  constructor() {
    this.client = new OpenAI({
      apiKey: config.openai.apiKey
    });
  }

  /**
   * Generate code (non-streaming)
   */
  async generate(systemPrompt: string, userPrompt: string): Promise<{
    code: string;
    tokensUsed: number;
    model: string;
  }> {
    try {
      logger.debug('Calling OpenAI API...');

      const response = await this.client.chat.completions.create({
        model: config.openai.model,
        messages: [
          { role: 'system', content: systemPrompt },
          { role: 'user', content: userPrompt }
        ],
        temperature: config.openai.temperature,
        max_tokens: config.openai.maxTokens
      });

      const code = response.choices[0]?.message?.content || '';
      const tokensUsed = response.usage?.total_tokens || 0;

      logger.info(`OpenAI response: ${tokensUsed} tokens used`);

      return {
        code,
        tokensUsed,
        model: response.model
      };
    } catch (error: any) {
      logger.error('OpenAI API error:', error);

      if (error.status === 429) {
        throw new Error('OpenAI rate limit exceeded. Please try again later.');
      }

      if (error.status === 401) {
        throw new Error('OpenAI API key is invalid.');
      }

      throw new Error(`OpenAI API error: ${error.message}`);
    }
  }

  /**
   * Generate code (streaming)
   */
  async *generateStream(
    systemPrompt: string,
    userPrompt: string
  ): AsyncGenerator<string, { code: string; tokensUsed: number; model: string }> {
    try {
      logger.debug('Calling OpenAI API (streaming)...');

      const stream = await this.client.chat.completions.create({
        model: config.openai.model,
        messages: [
          { role: 'system', content: systemPrompt },
          { role: 'user', content: userPrompt }
        ],
        stream: true,
        temperature: config.openai.temperature,
        max_tokens: config.openai.maxTokens
      });

      let fullCode = '';

      for await (const chunk of stream) {
        const content = chunk.choices[0]?.delta?.content || '';
        if (content) {
          fullCode += content;
          yield content;
        }
      }

      // Estimate tokens (rough approximation: 1 token ≈ 4 characters)
      const tokensUsed = Math.ceil((systemPrompt.length + userPrompt.length + fullCode.length) / 4);

      logger.info(`OpenAI streaming complete: ~${tokensUsed} tokens used`);

      return {
        code: fullCode,
        tokensUsed,
        model: config.openai.model
      };
    } catch (error: any) {
      logger.error('OpenAI streaming error:', error);

      if (error.status === 429) {
        throw new Error('OpenAI rate limit exceeded. Please try again later.');
      }

      if (error.status === 401) {
        throw new Error('OpenAI API key is invalid.');
      }

      throw new Error(`OpenAI API error: ${error.message}`);
    }
  }

  /**
   * Check if OpenAI is configured
   */
  isConfigured(): boolean {
    return !!config.openai.apiKey && config.openai.apiKey.length > 0;
  }
}

// Export singleton instance
export default new OpenAIService();
