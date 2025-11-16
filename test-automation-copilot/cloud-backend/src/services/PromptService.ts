/**
 * Prompt Engineering Service
 *
 * This is your PROPRIETARY prompt engineering - your competitive advantage.
 * These prompts stay secret on the cloud backend.
 */

export interface ContextPayload {
  action: string;
  userQuery: string;
  spec?: string;
  relevantCode?: Array<{
    name: string;
    type: string;
    content: string;
    filePath: string;
    similarity?: number;
  }>;
  pageObjects?: Array<{
    name: string;
    filePath: string;
    methods: string[];
  }>;
  testMethods?: Array<{
    name: string;
    className: string;
  }>;
  workspace?: {
    totalClasses: number;
    totalTests: number;
    pageObjectCount: number;
  };
  framework?: string;
  testRunner?: string;
  elements?: Array<{
    name: string;
    locatorType: string;
    locatorValue: string;
  }>;
  errorInfo?: {
    brokenCode: string;
    errorMessage: string;
    errorType?: string;
  };
}

export class PromptService {
  /**
   * Get system prompt for action
   */
  static getSystemPrompt(action: string): string {
    const basePrompt = `You are an expert test automation engineer with 10+ years of experience.
You specialize in writing clean, maintainable, and robust test automation code following industry best practices.`;

    switch (action) {
      case 'pageobject':
        return `${basePrompt}

Your task is to generate Selenium Page Object classes that:
- Follow the Page Object Model design pattern
- Use PageFactory for element initialization
- Have clear, descriptive method names
- Include proper JavaDoc comments
- Handle waits appropriately
- Are thread-safe when needed
- Follow naming conventions (ClassName + "Page")`;

      case 'test':
        return `${basePrompt}

Your task is to generate test cases that:
- Are independent and can run in any order
- Have clear test names describing what they test
- Follow the Arrange-Act-Assert pattern
- Include proper annotations (@Test, @BeforeMethod, etc.)
- Have meaningful assertions
- Clean up after themselves`;

      case 'chat':
        return `${basePrompt}

Your task is to help users with test automation questions:
- Provide clear, accurate answers
- Include code examples when relevant
- Explain best practices
- Suggest improvements
- Reference the user's existing codebase when possible`;

      case 'fix':
        return `${basePrompt}

Your task is to fix broken test automation code:
- Identify the root cause of the error
- Provide a corrected version
- Explain what was wrong
- Suggest how to prevent similar issues
- Maintain the original code structure and style`;

      default:
        return basePrompt;
    }
  }

  /**
   * Build complete prompt from context
   */
  static buildPrompt(action: string, context: ContextPayload): string {
    switch (action) {
      case 'pageobject':
        return this.buildPageObjectPrompt(context);
      case 'test':
        return this.buildTestPrompt(context);
      case 'chat':
        return this.buildChatPrompt(context);
      case 'fix':
        return this.buildFixPrompt(context);
      default:
        return context.userQuery || '';
    }
  }

  /**
   * Build page object generation prompt
   */
  private static buildPageObjectPrompt(context: ContextPayload): string {
    const { userQuery, spec, relevantCode, pageObjects, framework, testRunner, elements, workspace } = context;

    let prompt = `Generate a ${framework || 'Selenium Java'} page object class.\n\n`;

    // User specification
    prompt += `## User Request\n${spec || userQuery}\n\n`;

    // Elements to include
    if (elements && elements.length > 0) {
      prompt += `## Page Elements\n`;
      elements.forEach(elem => {
        prompt += `- **${elem.name}**: ${elem.locatorType} = "${elem.locatorValue}"\n`;
      });
      prompt += '\n';
    }

    // Similar page objects for reference
    if (pageObjects && pageObjects.length > 0) {
      prompt += `## Similar Page Objects in Codebase\n`;
      prompt += `For reference, here are similar page objects from the user's project:\n\n`;

      pageObjects.slice(0, 2).forEach((po, i) => {
        prompt += `### ${i + 1}. ${po.name}\n`;
        prompt += `Path: ${po.filePath}\n`;
        if (po.methods.length > 0) {
          prompt += `Methods: ${po.methods.join(', ')}\n`;
        }
        prompt += '\n';
      });
    }

    // Relevant code examples
    if (relevantCode && relevantCode.length > 0) {
      prompt += `## Example Code from Project\n`;
      prompt += `Here's an example from the user's codebase to match the style:\n\n`;
      prompt += '```java\n';
      prompt += relevantCode[0].content.substring(0, 1000); // Limit to 1000 chars
      prompt += '\n```\n\n';
    }

    // Workspace context
    if (workspace) {
      prompt += `## Project Context\n`;
      prompt += `- Total Page Objects: ${workspace.pageObjectCount}\n`;
      prompt += `- Total Test Classes: ${workspace.totalClasses}\n`;
      prompt += `- Total Tests: ${workspace.totalTests}\n\n`;
    }

    // Framework specific requirements
    prompt += `## Requirements\n`;
    prompt += `- Framework: ${framework || 'Selenium WebDriver'}\n`;
    prompt += `- Test Runner: ${testRunner || 'TestNG'}\n`;
    prompt += `- Use PageFactory pattern with @FindBy annotations\n`;
    prompt += `- Include constructor that takes WebDriver\n`;
    prompt += `- Add methods for each element interaction\n`;
    prompt += `- Include JavaDoc comments\n`;
    prompt += `- Match the coding style from the examples above\n\n`;

    prompt += `Generate the complete Java page object class now:`;

    return prompt;
  }

  /**
   * Build test generation prompt
   */
  private static buildTestPrompt(context: ContextPayload): string {
    const { userQuery, spec, relevantCode, testMethods, pageObjects, framework, testRunner } = context;

    let prompt = `Generate a ${testRunner || 'TestNG'} test case.\n\n`;

    prompt += `## User Request\n${spec || userQuery}\n\n`;

    // Available page objects
    if (pageObjects && pageObjects.length > 0) {
      prompt += `## Available Page Objects\n`;
      pageObjects.forEach(po => {
        prompt += `- ${po.name}: ${po.methods.join(', ')}\n`;
      });
      prompt += '\n';
    }

    // Similar tests for reference
    if (testMethods && testMethods.length > 0) {
      prompt += `## Similar Tests in Project\n`;
      testMethods.slice(0, 3).forEach(test => {
        prompt += `- ${test.className}.${test.name}\n`;
      });
      prompt += '\n';
    }

    // Example code
    if (relevantCode && relevantCode.length > 0) {
      prompt += `## Example Test Code\n`;
      prompt += '```java\n';
      prompt += relevantCode[0].content.substring(0, 1000);
      prompt += '\n```\n\n';
    }

    prompt += `## Requirements\n`;
    prompt += `- Framework: ${framework || 'Selenium WebDriver'}\n`;
    prompt += `- Test Runner: ${testRunner || 'TestNG'}\n`;
    prompt += `- Use @Test annotation\n`;
    prompt += `- Follow Arrange-Act-Assert pattern\n`;
    prompt += `- Include setup (@BeforeMethod) if needed\n`;
    prompt += `- Use the available page objects\n`;
    prompt += `- Add meaningful assertions\n`;
    prompt += `- Match the style from examples\n\n`;

    prompt += `Generate the complete test class now:`;

    return prompt;
  }

  /**
   * Build chat/Q&A prompt
   */
  private static buildChatPrompt(context: ContextPayload): string {
    const { userQuery, relevantCode, workspace, framework } = context;

    let prompt = `## User Question\n${userQuery}\n\n`;

    if (workspace) {
      prompt += `## User's Project Context\n`;
      prompt += `- Framework: ${framework || 'Selenium WebDriver'}\n`;
      prompt += `- Page Objects: ${workspace.pageObjectCount}\n`;
      prompt += `- Test Classes: ${workspace.totalClasses}\n`;
      prompt += `- Tests: ${workspace.totalTests}\n\n`;
    }

    if (relevantCode && relevantCode.length > 0) {
      prompt += `## Relevant Code from User's Project\n`;
      relevantCode.slice(0, 3).forEach((code, i) => {
        prompt += `\n### ${code.name} (${code.type})\n`;
        prompt += '```java\n';
        prompt += code.content.substring(0, 500);
        prompt += '\n```\n';
      });
      prompt += '\n';
    }

    prompt += `Please provide a helpful, accurate answer with code examples if relevant.`;

    return prompt;
  }

  /**
   * Build code fix prompt
   */
  private static buildFixPrompt(context: ContextPayload): string {
    const { userQuery, errorInfo, relevantCode } = context;

    if (!errorInfo) {
      return `Fix this code:\n\n${userQuery}`;
    }

    let prompt = `## Broken Code\n`;
    prompt += '```java\n';
    prompt += errorInfo.brokenCode;
    prompt += '\n```\n\n';

    prompt += `## Error Message\n`;
    prompt += '```\n';
    prompt += errorInfo.errorMessage;
    prompt += '\n```\n\n';

    if (errorInfo.errorType) {
      prompt += `## Error Type\n${errorInfo.errorType}\n\n`;
    }

    // Working examples
    if (relevantCode && relevantCode.length > 0) {
      prompt += `## Working Examples from Project\n`;
      prompt += `Here are similar working code snippets:\n\n`;

      relevantCode.slice(0, 2).forEach((code, i) => {
        prompt += `### Example ${i + 1}: ${code.name}\n`;
        prompt += '```java\n';
        prompt += code.content.substring(0, 500);
        prompt += '\n```\n\n`;
      });
    }

    prompt += `## Task\n`;
    prompt += `1. Identify what's wrong with the code\n`;
    prompt += `2. Provide the corrected version\n`;
    prompt += `3. Explain what was fixed and why\n`;
    prompt += `4. Suggest how to prevent this in the future\n\n`;

    prompt += `Provide the fixed code now:`;

    return prompt;
  }
}
