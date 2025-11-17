import * as vscode from 'vscode';
import * as path from 'path';
import { LanguageClient, LanguageClientOptions, ServerOptions } from 'vscode-languageclient/node';
import { ChatPanel } from './chatPanel';
import { AuthService } from './authService';
import { FrameworkDetector, FrameworkInfo } from './frameworkDetector';

let client: LanguageClient;
let chatPanel: ChatPanel | undefined;
let statusBarItem: vscode.StatusBarItem;
let totalCost: number = 0;
let authService: AuthService;
let frameworkDetector: FrameworkDetector;
let currentFramework: FrameworkInfo | null = null;

export function activate(context: vscode.ExtensionContext) {
    console.log('Test Automation Copilot is now active!');

    // Initialize services
    authService = new AuthService(context);
    frameworkDetector = new FrameworkDetector();

    // Detect framework
    detectCurrentFramework();

    // Create status bar item
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    updateStatusBar();
    statusBarItem.command = 'testCopilot.showAccount';
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);

    // Start LSP server
    startLanguageServer(context);

    // Register commands
    registerCommands(context);

    // Check if signed in
    checkAuthStatus();
}

async function detectCurrentFramework() {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (workspaceFolder) {
        currentFramework = await frameworkDetector.detectFramework(workspaceFolder);
        updateStatusBar();
    }
}

async function checkAuthStatus() {
    const isSignedIn = await authService.isSignedIn();
    if (!isSignedIn) {
        // Show welcome message
        const action = await vscode.window.showInformationMessage(
            'Welcome to Test Automation Copilot! Sign in to get started.',
            'Sign In',
            'Learn More'
        );

        if (action === 'Sign In') {
            await vscode.commands.executeCommand('testCopilot.signIn');
        } else if (action === 'Learn More') {
            vscode.env.openExternal(vscode.Uri.parse('https://testcopilot.dev'));
        }
    }
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
        documentSelector: [
            { scheme: 'file', language: 'java' },
            { scheme: 'file', language: 'typescript' },
            { scheme: 'file', language: 'javascript' },
            { scheme: 'file', language: 'python' }
        ],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.{java,ts,js,py}')
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
    const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath || '';

    // Get user data
    const userData = authService.getUserData();
    let apiKey = '';

    if (userData) {
        if (userData.apiKeyMode === 'byok') {
            // Use user's own API key
            apiKey = config.get<string>('apiKey') || process.env.ANTHROPIC_API_KEY || '';
            if (!apiKey) {
                vscode.window.showWarningMessage(
                    'BYOK mode requires API key. Please set it in settings.',
                    'Open Settings'
                ).then(selection => {
                    if (selection === 'Open Settings') {
                        vscode.commands.executeCommand('workbench.action.openSettings', 'testCopilot.apiKey');
                    }
                });
                return;
            }
        } else if (userData.apiKeyMode === 'managed') {
            // Use managed API - get token
            const token = await authService.getToken();
            if (!token) {
                vscode.window.showErrorMessage('Authentication required. Please sign in again.');
                return;
            }
            // Token will be used for backend API calls instead of direct Claude calls
            apiKey = '';  // Don't need API key for managed mode
        }
        // For 'local' mode, no API key needed
    }

    try {
        await client.sendRequest('testCopilot/initialize', {
            workspaceRoot,
            apiKey,
            userMode: userData?.apiKeyMode || 'byok'
        });

        if (userData?.apiKeyMode === 'byok' && apiKey) {
            vscode.window.showInformationMessage('Test Copilot initialized successfully (BYOK mode)!');
        } else if (userData?.apiKeyMode === 'local') {
            vscode.window.showInformationMessage('Test Copilot initialized (Local LLM mode)!');
        }
    } catch (error) {
        vscode.window.showErrorMessage(`Test Copilot initialization failed: ${error}`);
    }
}

function registerCommands(context: vscode.ExtensionContext) {
    // Sign in command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.signIn', async () => {
            const userData = await authService.signIn();
            if (userData) {
                updateStatusBar();
                await initializeCopilot();
            }
        })
    );

    // Sign out command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.signOut', async () => {
            await authService.signOutUser();
            updateStatusBar();
        })
    );

    // Show account command
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.showAccount', async () => {
            const userData = authService.getUserData();
            if (!userData) {
                await vscode.commands.executeCommand('testCopilot.signIn');
                return;
            }

            const tierInfo = authService.getTierInfo();
            const panel = vscode.window.createWebviewPanel(
                'testCopilotAccount',
                'Test Copilot Account',
                vscode.ViewColumn.Two,
                {}
            );

            panel.webview.html = getAccountHTML(userData, tierInfo, currentFramework);
        })
    );

    // Generate command (with auth check)
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.generate', async () => {
            // Check auth
            const userData = authService.getUserData();
            if (!userData) {
                const action = await vscode.window.showInformationMessage(
                    'Please sign in to use Test Copilot',
                    'Sign In'
                );
                if (action === 'Sign In') {
                    await vscode.commands.executeCommand('testCopilot.signIn');
                }
                return;
            }

            // Check framework allowed
            if (currentFramework && !authService.isFrameworkAllowed(currentFramework.id)) {
                await authService.showUpgradeDialog(
                    `${currentFramework.name} requires Pro tier. Upgrade to access all frameworks!`
                );
                return;
            }

            // Check usage limit
            if (authService.isUsageLimitReached()) {
                await authService.showUpgradeDialog(
                    'Monthly usage limit reached. Upgrade for more generations or switch to BYOK mode!'
                );
                return;
            }

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
            // Check auth first
            const userData = authService.getUserData();
            if (!userData) {
                vscode.commands.executeCommand('testCopilot.generate');
                return;
            }

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

    // Chat command (with auth check)
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.chat', async () => {
            const userData = authService.getUserData();
            if (!userData) {
                const action = await vscode.window.showInformationMessage(
                    'Please sign in to use Chat',
                    'Sign In'
                );
                if (action === 'Sign In') {
                    await vscode.commands.executeCommand('testCopilot.signIn');
                }
                return;
            }

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

    // Switch API key mode
    context.subscriptions.push(
        vscode.commands.registerCommand('testCopilot.switchMode', async () => {
            const userData = authService.getUserData();
            if (!userData) {
                vscode.window.showErrorMessage('Please sign in first');
                return;
            }

            const modes = [
                { label: '🔑 BYOK', description: 'Use your own Claude API key', value: 'byok' },
                { label: '💻 Local LLM', description: 'Use local LLM (basic)', value: 'local' },
            ];

            if (userData.tier !== 'free') {
                modes.push({ label: '☁️  Managed API', description: 'Use our managed API', value: 'managed' });
            }

            const selection = await vscode.window.showQuickPick(modes, {
                placeHolder: 'Select API key mode'
            });

            if (selection) {
                // Update in Firestore (would need backend API)
                vscode.window.showInformationMessage(`Switched to ${selection.label} mode`);
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
                language: currentFramework?.language || 'java'
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

    const suggestedPath = path.join(workspaceRoot, 'src', 'test', currentFramework?.language || 'java', result.suggestedPath);

    const uri = await vscode.window.showSaveDialog({
        defaultUri: vscode.Uri.file(suggestedPath),
        filters: currentFramework?.language === 'java'
            ? { 'Java': ['java'] }
            : currentFramework?.language === 'python'
            ? { 'Python': ['py'] }
            : { 'TypeScript': ['ts'], 'JavaScript': ['js'] }
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
    const userData = authService.getUserData();

    let text = '$(robot) Test Copilot';

    if (userData) {
        const tierInfo = authService.getTierInfo();
        text += ` [${tierInfo.name}]`;
    }

    if (currentFramework) {
        text += ` ${currentFramework.icon}`;
    }

    if (showCost && totalCost > 0) {
        text += ` | $${totalCost.toFixed(2)}`;
    }

    statusBarItem.text = text;
    statusBarItem.tooltip = userData
        ? `Signed in as ${userData.displayName}\nTier: ${userData.tier}\nClick for account details`
        : 'Click to sign in';
}

function getAccountHTML(userData: any, tierInfo: any, framework: FrameworkInfo | null): string {
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
                .tier {
                    background: ${tierInfo.color};
                    color: white;
                    padding: 10px 20px;
                    border-radius: 5px;
                    display: inline-block;
                    font-weight: bold;
                }
                .info { margin: 20px 0; }
                .feature {
                    padding: 5px 0;
                }
                .feature:before {
                    content: "✓ ";
                    color: #4CAF50;
                }
                button {
                    padding: 10px 20px;
                    background: var(--vscode-button-background);
                    color: var(--vscode-button-foreground);
                    border: none;
                    border-radius: 3px;
                    cursor: pointer;
                    margin: 10px 10px 0 0;
                }
            </style>
        </head>
        <body>
            <h1>Test Automation Copilot</h1>

            <div class="info">
                <strong>Signed in as:</strong> ${userData.displayName}<br>
                <strong>Email:</strong> ${userData.email}<br>
                <strong>User ID:</strong> ${userData.uid}
            </div>

            <div class="info">
                <div class="tier">${tierInfo.name} TIER</div>
            </div>

            <div class="info">
                <strong>Current Framework:</strong> ${framework ? framework.name : 'Not detected'}<br>
                <strong>API Mode:</strong> ${userData.apiKeyMode.toUpperCase()}<br>
                <strong>Monthly Usage:</strong> ${userData.monthlyUsage}${userData.usageLimit > 0 ? ` / ${userData.usageLimit}` : ' (unlimited)'}
            </div>

            <div class="info">
                <h3>Features:</h3>
                ${tierInfo.features.map((f: string) => `<div class="feature">${f}</div>`).join('')}
            </div>

            ${userData.tier === 'free' ? `
                <div style="margin-top: 30px; padding: 20px; background: var(--vscode-editor-inactiveSelectionBackground); border-radius: 5px;">
                    <h3>Upgrade to Pro</h3>
                    <p>Get access to 5+ frameworks, managed API, and more!</p>
                    <button onclick="window.open('https://testcopilot.dev/upgrade')">Upgrade Now - $19/month</button>
                </div>
            ` : ''}
        </body>
        </html>
    `;
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
