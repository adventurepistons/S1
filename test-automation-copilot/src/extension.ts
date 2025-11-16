import * as vscode from 'vscode';
import { ChatPanelProvider } from './ui/ChatPanelProvider';
import { ElementRecorder } from './browser/ElementRecorder';
import { CoreClient } from './api/CoreClient';

export async function activate(context: vscode.ExtensionContext) {
    console.log('Test Automation Copilot is now active!');

    // Initialize Core Client (Go binary)
    const coreClient = new CoreClient(context.extensionPath, context);
    (global as any).testCopilotCoreClient = coreClient; // Store globally for deactivate

    try {
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Starting Copilot Core...",
            cancellable: false
        }, async () => {
            await coreClient.startServer();
        });

        vscode.window.showInformationMessage('✅ Copilot Core started successfully!');
    } catch (error) {
        vscode.window.showWarningMessage(
            `Copilot Core failed to start: ${error}. Some features may be limited.`,
            'Build Binary'
        ).then(selection => {
            if (selection === 'Build Binary') {
                vscode.window.showInformationMessage(
                    'Please build the Go binary:\ncd copilot-core && go build -o server main.go'
                );
            }
        });
    }

    context.subscriptions.push({
        dispose: () => coreClient.stopServer()
    });

    // Initialize chat panel (no more local storage/analyzer needed)
    const chatProvider = new ChatPanelProvider(context.extensionUri, coreClient);

    // Initialize element recorder
    const elementRecorder = new ElementRecorder(chatProvider);

    // Register chat panel
    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(
            'testCopilot.chatView',
            chatProvider
        )
    );

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.openChat', () => {
            chatProvider.show();
            vscode.window.showInformationMessage('Test Copilot Chat opened!');
        })
    );

    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.analyzeWorkspace', async () => {
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: "Indexing workspace...",
                cancellable: false
            }, async (progress) => {
                try {
                    progress.report({ increment: 0, message: "Starting indexing..." });

                    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
                    if (!workspaceFolder) {
                        vscode.window.showErrorMessage('No workspace folder open');
                        return;
                    }

                    progress.report({ increment: 20, message: "Indexing files in Go backend..." });

                    // Call Go backend indexing API
                    await coreClient.indexWorkspace(workspaceFolder.uri.fsPath);

                    progress.report({ increment: 70, message: "Fetching workspace stats..." });

                    // Get statistics
                    const stats = await coreClient.getWorkspaceStats();

                    progress.report({ increment: 100, message: "Complete!" });

                    vscode.window.showInformationMessage(
                        `✅ Workspace indexed!\n` +
                        `📁 ${stats.files} files\n` +
                        `📦 ${stats.classes} classes\n` +
                        `📄 ${stats.pageObjects} page objects\n` +
                        `✅ ${stats.testMethods} test methods\n` +
                        `🔍 ${stats.vectorDocs} code chunks indexed for semantic search`
                    );

                } catch (error) {
                    vscode.window.showErrorMessage(`Indexing failed: ${error}`);
                }
            });
        })
    );

    // Start Recording Session Command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.startRecording', async () => {
            try {
                await elementRecorder.startRecording();
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to start recording: ${error}`);
            }
        })
    );

    // Stop Recording Session Command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.stopRecording', async () => {
            try {
                await elementRecorder.stopRecording();
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to stop recording: ${error}`);
            }
        })
    );

    // Pause Recording Command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.pauseRecording', async () => {
            try {
                await elementRecorder.pauseRecording();
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to pause recording: ${error}`);
            }
        })
    );

    // Resume Recording Command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.resumeRecording', async () => {
            try {
                await elementRecorder.resumeRecording();
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to resume recording: ${error}`);
            }
        })
    );

    // Legacy recordElements command (for backward compatibility)
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.recordElements', async () => {
            const url = await vscode.window.showInputBox({
                prompt: 'Enter the URL to record elements from',
                placeHolder: 'https://example.com/login'
            });

            if (!url) {
                return;
            }

            try {
                await elementRecorder.startRecording(url);
            } catch (error) {
                vscode.window.showErrorMessage(`Recording failed: ${error}`);
            }
        })
    );

    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.generateTest', async () => {
            const testName = await vscode.window.showInputBox({
                prompt: 'What test would you like to generate?',
                placeHolder: 'e.g., Login test with valid credentials'
            });

            if (testName) {
                chatProvider.requestCodeGeneration('test', testName);
            }
        })
    );

    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.generatePageObject', async () => {
            const pageName = await vscode.window.showInputBox({
                prompt: 'Page Object name?',
                placeHolder: 'e.g., LoginPage'
            });

            if (pageName) {
                chatProvider.requestCodeGeneration('pageObject', pageName);
            }
        })
    );

    // Check cloud configuration
    const config = vscode.workspace.getConfiguration('testCopilot');
    const apiKey = config.get<string>('apiKey');

    if (!apiKey || apiKey === '') {
        vscode.window.showWarningMessage(
            'Test Copilot API key not configured. AI features will not work.',
            'Configure Now',
            'Get API Key'
        ).then(selection => {
            if (selection === 'Configure Now') {
                vscode.commands.executeCommand('workbench.action.openSettings', 'testCopilot.apiKey');
            } else if (selection === 'Get API Key') {
                vscode.env.openExternal(vscode.Uri.parse('https://testcopilot.ai/signup'));
            }
        });
    }

    // Create status bar item for usage tracking
    if (config.get('enableUsageTracking', true)) {
        const statusBarItem = vscode.window.createStatusBarItem(
            vscode.StatusBarAlignment.Right,
            100
        );
        statusBarItem.text = '$(cloud) Test Copilot';
        statusBarItem.tooltip = 'Click to view usage statistics';
        statusBarItem.command = 'testCopilot.showUsageStats';
        statusBarItem.show();
        context.subscriptions.push(statusBarItem);

        // Update usage stats periodically (every 5 minutes)
        const updateStats = async () => {
            const stats = await coreClient.getUsageStats();
            if (stats) {
                statusBarItem.text = `$(cloud) ${stats.usage.requestsThisMonth}/${stats.limits.requestsPerMonth}`;
                statusBarItem.tooltip = `Test Copilot Usage:\n${stats.usage.requestsThisMonth} requests this month\n${stats.limits.requestsRemaining} remaining`;
            }
        };

        // Initial update
        setTimeout(updateStats, 5000);

        // Periodic updates
        const interval = setInterval(updateStats, 300000); // 5 minutes
        context.subscriptions.push({
            dispose: () => clearInterval(interval)
        });
    }

    // Add command to show usage statistics
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.showUsageStats', async () => {
            const stats = await coreClient.getUsageStats();
            if (stats) {
                const message = `
**Test Copilot Usage Statistics**

**Plan**: ${stats.plan}
**Requests This Month**: ${stats.usage.requestsThisMonth} / ${stats.limits.requestsPerMonth}
**Tokens This Month**: ${stats.usage.tokensThisMonth.toLocaleString()}
**Estimated Cost**: $${stats.usage.estimatedCost.toFixed(2)}

**Requests by Type:**
- Page Objects: ${stats.usage.requestsByAction.pageobject}
- Tests: ${stats.usage.requestsByAction.test}
- Chat: ${stats.usage.requestsByAction.chat}
- Fixes: ${stats.usage.requestsByAction.fix}

**Remaining**: ${stats.limits.requestsRemaining} requests
                `.trim();

                vscode.window.showInformationMessage(message, 'Upgrade Plan', 'Close')
                    .then(selection => {
                        if (selection === 'Upgrade Plan') {
                            vscode.env.openExternal(vscode.Uri.parse('https://testcopilot.ai/pricing'));
                        }
                    });
            } else {
                vscode.window.showInformationMessage(
                    'Usage statistics not available. Make sure your API key is configured and cloud backend is accessible.',
                    'Configure API Key'
                ).then(selection => {
                    if (selection === 'Configure API Key') {
                        vscode.commands.executeCommand('workbench.action.openSettings', 'testCopilot.apiKey');
                    }
                });
            }
        })
    );

    // Auto-analyze workspace on activation if enabled
    if (config.get('autoAnalyze', true)) {
        setTimeout(() => {
            vscode.commands.executeCommand('testCopilot.analyzeWorkspace');
        }, 2000); // Give VS Code time to fully load
    }

    // Show welcome message
    vscode.window.showInformationMessage(
        'Test Automation Copilot activated! Click the Test Copilot icon to start.',
        'Open Chat'
    ).then(selection => {
        if (selection === 'Open Chat') {
            vscode.commands.executeCommand('testCopilot.openChat');
        }
    });
}

export function deactivate() {
    console.log('Test Automation Copilot deactivated');

    // Stop Core Client server
    const coreClient = (global as any).testCopilotCoreClient;
    if (coreClient) {
        coreClient.stopServer();
    }
}
