import { Router, Request, Response } from 'express';
import { authenticate } from '../middleware/auth';
import { PromptService, ContextPayload } from '../services/PromptService';
import openAIService from '../services/OpenAIService';
import { UsageModel } from '../models/Usage';
import { config } from '../config';
import logger from '../utils/logger';

const router = Router();

/**
 * POST /v1/generate
 * Generate code (non-streaming)
 */
router.post('/generate', authenticate, async (req: Request, res: Response) => {
  try {
    const context: ContextPayload = req.body;

    if (!context.action) {
      return res.status(400).json({ error: 'Missing required field: action' });
    }

    logger.info(`[${req.userId}] Generate request: ${context.action}`);

    // Build prompts using our proprietary prompt engineering
    const systemPrompt = PromptService.getSystemPrompt(context.action);
    const userPrompt = PromptService.buildPrompt(context.action, context);

    // Generate code using OpenAI
    const result = await openAIService.generate(systemPrompt, userPrompt);

    // Track usage
    await UsageModel.record(
      req.userId!,
      context.action as any,
      result.tokensUsed,
      result.model
    );

    logger.info(`[${req.userId}] Generated successfully: ${result.tokensUsed} tokens`);

    // Return response
    res.json({
      code: result.code,
      tokensUsed: result.tokensUsed,
      model: result.model,
      finishReason: 'stop'
    });
  } catch (error: any) {
    logger.error('Generate error:', error);
    res.status(500).json({
      error: `Failed to generate code: ${error.message}`
    });
  }
});

/**
 * POST /v1/generate/stream
 * Generate code (streaming)
 */
router.post('/generate/stream', authenticate, async (req: Request, res: Response) => {
  try {
    const context: ContextPayload = req.body;

    if (!context.action) {
      return res.status(400).json({ error: 'Missing required field: action' });
    }

    logger.info(`[${req.userId}] Stream request: ${context.action}`);

    // Set SSE headers
    res.setHeader('Content-Type', 'text/event-stream');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');

    // Build prompts
    const systemPrompt = PromptService.getSystemPrompt(context.action);
    const userPrompt = PromptService.buildPrompt(context.action, context);

    // Stream generation
    const stream = openAIService.generateStream(systemPrompt, userPrompt);

    let finalResult: { code: string; tokensUsed: number; model: string } | undefined;

    for await (const chunk of stream) {
      // Check if this is the final result or a chunk
      if (typeof chunk === 'string') {
        res.write(`data: ${JSON.stringify({ type: 'chunk', content: chunk })}\n\n`);
      } else {
        // This is the return value from the generator
        finalResult = chunk;
      }
    }

    // If we didn't get a final result, create a default one
    if (!finalResult) {
      finalResult = { code: '', tokensUsed: 0, model: config.openai.model };
    }

    // Track usage
    await UsageModel.record(
      req.userId!,
      context.action as any,
      finalResult.tokensUsed,
      finalResult.model
    );

    logger.info(`[${req.userId}] Stream complete: ${finalResult.tokensUsed} tokens`);

    // Send completion
    res.write(`data: ${JSON.stringify({
      type: 'complete',
      result: {
        code: finalResult.code,
        tokensUsed: finalResult.tokensUsed,
        model: finalResult.model,
        finishReason: 'stop'
      }
    })}\n\n`);

    res.end();
  } catch (error: any) {
    logger.error('Stream error:', error);
    res.write(`data: ${JSON.stringify({ type: 'error', error: error.message })}\n\n`);
    res.end();
  }
});

export default router;
