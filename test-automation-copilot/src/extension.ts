import * as vscode from 'vscode';
import { WorkspaceAnalyzer } from './core/analyzers/WorkspaceAnalyzer';
import { ChatPanelProvider } from './ui/ChatPanelProvider';
import { ElementRecorder } from './browser/ElementRecorder';
import { StorageManager } from './core/storage/StorageManager';

export async function activate(context: vscode.ExtensionContext) {
    console.log('Test Automation Copilot is now active!');

    // Initialize storage (SQLite + ChromaDB)
    const storageManager = new StorageManager(context.globalStorageUri.fsPath);
    await storageManager.initialize();

    // Initialize workspace analyzer
    const workspaceAnalyzer = new WorkspaceAnalyzer(storageManager);

    // Initialize chat panel
    const chatProvider = new ChatPanelProvider(context.extensionUri, storageManager, workspaceAnalyzer);

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
}
