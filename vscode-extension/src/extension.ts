import * as vscode from 'vscode';
import * as path from 'path';
import { LanguageClient, LanguageClientOptions, ServerOptions } from 'vscode-languageclient/node';
import { ChatPanel } from './chatPanel';

let client: LanguageClient;
let chatPanel: ChatPanel | undefined;
let statusBarItem: vscode.StatusBarItem;
let totalCost: number = 0;

export function activate(context: vscode.ExtensionContext) {
    console.log('Test Automation Copilot is now active!');

    // Create status bar item
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.text = '$(robot) Test Copilot';
    statusBarItem.tooltip = 'Test Automation Copilot';
    statusBarItem.command = 'testCopilot.showStats';
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);

    // Start LSP server
    startLanguageServer(context);

    // Register commands
    registerCommands(context);
}

function startLanguageServer(context: vscode.ExtensionContext) {
    // Path to Go LSP server
    const serverPath = context.asAbsolutePath(path.join('bin', 'copilot-lsp'));

    // Server options
    const serverOptions: ServerOptions = {
        command: serverPath,
        args: [],
    };

    // Client options
    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'java' }],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.java')
        }
    };

    // Create language client
    client = new LanguageClient(
        'testCopilot',
        'Test Automation Copilot',
        serverOptions,
        clientOptions
    );

    // Start client
    client.start();

    // Initialize copilot backend after client is ready
    client.onReady().then(() => {
        initializeCopilot();
    });
}

async function initializeCopilot() {
    const config = vscode.workspace.getConfiguration('testCopilot');
    const apiKey = config.get<string>('apiKey') || process.env.ANTHROPIC_API_KEY || '';
    const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath || '';

    if (!apiKey) {
        vscode.window.showWarningMessage(
            'Test Copilot: No API key found. Please set testCopilot.apiKey in settings or ANTHROPIC_API_KEY environment variable.',
            'Open Settings'
        ).then(selection => {
            if (selection === 'Open Settings') {
                vscode.commands.executeCommand('workbench.action.openSettings', 'testCopilot.apiKey');
            }
        });
        return;
    }

    try {
        await client.sendRequest('testCopilot/initialize', {
            workspaceRoot,
            apiKey
        });
        vscode.window.showInformationMessage('Test Copilot initialized successfully!');
    } catch (error) {
        vscode.window.showErrorMessage(`Test Copilot initialization failed: ${error}`);
    }
}

function registerCommands(context: vscode.ExtensionContext) {
    // Generate command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.generate', async () => {
            const request = await vscode.window.showInputBox({
                prompt: 'What would you like to generate?',
                placeHolder: 'E.g., Create test for login with valid credentials'
            });

            if (!request) {
                return;
            }

            await generateCode(request);
        })
    );

    // Generate for current class
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.generateForClass', async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showErrorMessage('No active editor');
                return;
            }

            const document = editor.document;
            const className = getClassNameFromDocument(document);

            const request = await vscode.window.showInputBox({
                prompt: `Generate test for ${className}`,
                placeHolder: 'E.g., test login functionality',
                value: `Create test for ${className}`
            });

            if (!request) {
                return;
            }

            await generateCode(request);
        })
    );

    // Chat command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.chat', () => {
            if (!chatPanel) {
                chatPanel = new ChatPanel(context.extensionUri, client);
            }
            chatPanel.show();
        })
    );

    // Build index command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.buildIndex', async () => {
            vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: 'Building Test Copilot index...',
                cancellable: false
            }, async (progress) => {
                try {
                    progress.report({ increment: 0, message: 'Indexing codebase...' });

                    await client.sendRequest('testCopilot/buildIndex', {});

                    progress.report({ increment: 100, message: 'Complete!' });
                    vscode.window.showInformationMessage('Index built successfully!');
                } catch (error) {
                    vscode.window.showErrorMessage(`Build index failed: ${error}`);
                }
            });
        })
    );

    // Show stats command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.showStats', async () => {
            try {
                const stats = await client.sendRequest('testCopilot/getStats', {});

                const panel = vscode.window.createWebviewPanel(
                    'testCopilotStats',
                    'Test Copilot Statistics',
                    vscode.ViewColumn.Two,
                    {}
                );

                panel.webview.html = getStatsHTML(stats);
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to get stats: ${error}`);
            }
        })
    );
}

async function generateCode(request: string) {
    try {
        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Generating code...',
            cancellable: false
        }, async (progress) => {
            progress.report({ increment: 0, message: 'Building context...' });

            const result: any = await client.sendRequest('testCopilot/generate', { request });

            progress.report({ increment: 50, message: 'Processing response...' });

            // Update cost
            totalCost += result.cost;
            updateStatusBar();

            // Show generated code
            const document = await vscode.workspace.openTextDocument({
                content: result.code,
                language: 'java'
            });

            await vscode.window.showTextDocument(document, vscode.ViewColumn.Beside);

            progress.report({ increment: 100, message: 'Complete!' });

            // Show info
            const costStr = `$${result.cost.toFixed(4)}`;
            const action = await vscode.window.showInformationMessage(
                `Generated ${result.className} (${result.codeType}) - Cost: ${costStr}`,
                'Save to File',
                'Copy to Clipboard'
            );

            if (action === 'Save to File') {
                await saveGeneratedCode(result);
            } else if (action === 'Copy to Clipboard') {
                await vscode.env.clipboard.writeText(result.code);
                vscode.window.showInformationMessage('Code copied to clipboard!');
            }
        });
    } catch (error) {
        vscode.window.showErrorMessage(`Generation failed: ${error}`);
    }
}

async function saveGeneratedCode(result: any) {
    const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
    if (!workspaceRoot) {
        vscode.window.showErrorMessage('No workspace folder open');
        return;
    }

    const suggestedPath = path.join(workspaceRoot, 'src', 'test', 'java', result.suggestedPath);

    const uri = await vscode.window.showSaveDialog({
        defaultUri: vscode.Uri.file(suggestedPath),
        filters: { 'Java': ['java'] }
    });

    if (uri) {
        await vscode.workspace.fs.writeFile(uri, Buffer.from(result.code, 'utf8'));
        vscode.window.showInformationMessage(`Code saved to ${uri.fsPath}`);

        // Open the saved file
        const document = await vscode.workspace.openTextDocument(uri);
        await vscode.window.showTextDocument(document);
    }
}

function getClassNameFromDocument(document: vscode.TextDocument): string {
    const text = document.getText();
    const match = text.match(/class\s+(\w+)/);
    return match ? match[1] : 'Unknown';
}

function updateStatusBar() {
    const config = vscode.workspace.getConfiguration('testCopilot');
    const showCost = config.get<boolean>('showCost', true);

    if (showCost && totalCost > 0) {
        statusBarItem.text = `$(robot) Test Copilot | $${totalCost.toFixed(2)}`;
    } else {
        statusBarItem.text = '$(robot) Test Copilot';
    }
}

function getStatsHTML(stats: any): string {
    return `
        <!DOCTYPE html>
        <html>
        <head>
            <style>
                body {
                    font-family: var(--vscode-font-family);
                    padding: 20px;
                    color: var(--vscode-foreground);
                }
                h1 { color: var(--vscode-textLink-foreground); }
                .stat { margin: 10px 0; }
                .stat-label { font-weight: bold; }
                .stat-value { color: var(--vscode-textLink-activeForeground); }
            </style>
        </head>
        <body>
            <h1>📊 Test Copilot Statistics</h1>
            <pre>${JSON.stringify(stats, null, 2)}</pre>
        </body>
        </html>
    `;
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
