/**
 * Test Execution Extension Integration
 *
 * This file shows how to integrate test execution into your extension.
 * Add these imports and commands to your main extension.ts file.
 */

import * as vscode from 'vscode';
import { TestExecutionPanel } from './panels/TestExecutionPanel';
import { initializeApiClient } from './api/client';

/**
 * Register test execution commands
 * Call this in your activate() function
 */
export function registerTestExecutionCommands(context: vscode.ExtensionContext) {
  // Initialize API client
  initializeApiClient(context);

  // Register "Execute Tests" command
  const executeTestsCommand = vscode.commands.registerCommand(
    'testCopilot.executeTests',
    () => {
      TestExecutionPanel.show(context.extensionUri);
    },
  );

  context.subscriptions.push(executeTestsCommand);
}

/**
 * Example: How to add to your existing extension.ts
 *
 * import { registerTestExecutionCommands } from './extension-test-execution';
 *
 * export function activate(context: vscode.ExtensionContext) {
 *   // ... your existing code ...
 *
 *   // Register test execution
 *   registerTestExecutionCommands(context);
 * }
 */
