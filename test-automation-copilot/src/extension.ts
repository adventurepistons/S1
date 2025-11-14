import * as vscode from 'vscode';
import { WorkspaceAnalyzer } from './core/analyzers/WorkspaceAnalyzer';
import { ChatPanelProvider } from './ui/ChatPanelProvider';
import { ElementRecorder } from './browser/ElementRecorder';
import { StorageManager } from './core/storage/StorageManager';
import { CoreClient } from './api/CoreClient';

export async function activate(context: vscode.ExtensionContext) {
    console.log('Test Automation Copilot is now active!');

    // Initialize Core Client (Go binary)
    const coreClient = new CoreClient(context.extensionPath);
    (global as any).testCopilotCoreClient = coreClient; // Store globally for deactivate

    try {
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Starting Copilot Core...",
            cancellable: false
        }, async () => {
            await coreClient.startServer();
        });
    } catch (error) {
        vscode.window.showWarningMessage(
            `Copilot Core failed to start: ${error}. Some features may be limited.`,
            'Build Binary'
        ).then(selection => {
            if (selection === 'Build Binary') {
                vscode.window.showInformationMessage(
                    'Please build the Go binary:\ncd copilot-core && go build -o copilot-core main.go'
                );
            }
        });
    }

    context.subscriptions.push({
        dispose: () => coreClient.stopServer()
    });

    // Initialize storage (SQLite + ChromaDB)
    const storageManager = new StorageManager(context.globalStorageUri.fsPath);
    await storageManager.initialize();

    // Initialize workspace analyzer
    const workspaceAnalyzer = new WorkspaceAnalyzer(storageManager);

    // Initialize chat panel
    const chatProvider = new ChatPanelProvider(context.extensionUri, storageManager, workspaceAnalyzer, coreClient);

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
                title: "Analyzing test framework...",
                cancellable: false
            }, async (progress) => {
                try {
                    progress.report({ increment: 0, message: "Scanning workspace..." });

                    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
                    if (!workspaceFolder) {
                        vscode.window.showErrorMessage('No workspace folder open');
                        return;
                    }

                    progress.report({ increment: 30, message: "Parsing Java files..." });
                    const analysis = await workspaceAnalyzer.analyzeWorkspace(workspaceFolder.uri.fsPath);

                    progress.report({ increment: 60, message: "Building context..." });
                    await storageManager.storeAnalysis(analysis);

                    progress.report({ increment: 100, message: "Complete!" });

                    vscode.window.showInformationMessage(
                        `Framework analyzed! Found: ${analysis.pageObjects.length} page objects, ${analysis.testCases.length} tests`
                    );

                    // Send to chat panel
                    chatProvider.sendAnalysisResults(analysis);
                } catch (error) {
                    vscode.window.showErrorMessage(`Analysis failed: ${error}`);
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

    // Auto-analyze workspace on activation if enabled
    const config = vscode.workspace.getConfiguration('testCopilot');
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
