import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import { RecordingSession, ApplicationMap } from '../browser/types';
import { CoreClient } from '../api/CoreClient';

export class ChatPanelProvider implements vscode.WebviewViewProvider {
    public static readonly viewType = 'testCopilot.chatView';
    private _view?: vscode.WebviewView;
    private extensionUri: vscode.Uri;
    private coreClient: CoreClient;

    constructor(
        extensionUri: vscode.Uri,
        coreClient: CoreClient
    ) {
        this.extensionUri = extensionUri;
        this.coreClient = coreClient;
    }

    public resolveWebviewView(
        webviewView: vscode.WebviewView,
        context: vscode.WebviewViewResolveContext,
        _token: vscode.CancellationToken
    ): void {
        this._view = webviewView;

        webviewView.webview.options = {
            enableScripts: true,
            localResourceRoots: [this.extensionUri]
        };

        webviewView.webview.html = this.getHtmlContent(webviewView.webview);

        // Handle messages from webview
        webviewView.webview.onDidReceiveMessage(async (data) => {
            switch (data.type) {
                case 'chat':
                    await this.handleChatMessage(data.message);
                    break;
                case 'generateCode':
                    await this.handleCodeGeneration(data.spec);
                    break;
            }
        });
    }

    private getHtmlContent(webview: vscode.Webview): string {
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test Copilot Chat</title>
    <style>
        body {
            padding: 10px;
            color: var(--vscode-foreground);
            font-family: var(--vscode-font-family);
        }
        .chat-container {
            display: flex;
            flex-direction: column;
            height: 100vh;
        }
        .messages {
            flex: 1;
            overflow-y: auto;
            padding: 10px;
        }
        .message {
            margin: 10px 0;
            padding: 10px;
            border-radius: 5px;
        }
        .user-message {
            background: var(--vscode-input-background);
            margin-left: 20px;
        }
        .assistant-message {
            background: var(--vscode-editor-background);
            margin-right: 20px;
        }
        .input-container {
            display: flex;
            gap: 5px;
            padding: 10px;
            border-top: 1px solid var(--vscode-widget-border);
        }
        input {
            flex: 1;
            padding: 8px;
            background: var(--vscode-input-background);
            color: var(--vscode-input-foreground);
            border: 1px solid var(--vscode-input-border);
            border-radius: 3px;
        }
        button {
            padding: 8px 15px;
            background: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            border-radius: 3px;
            cursor: pointer;
        }
        button:hover {
            background: var(--vscode-button-hoverBackground);
        }
        .recording-summary {
            padding: 15px;
            margin: 10px 0;
            border: 1px solid var(--vscode-widget-border);
            border-radius: 5px;
        }
        .page-item {
            padding: 8px;
            margin: 5px 0;
            background: var(--vscode-editor-background);
            border-radius: 3px;
        }
        .page-item:hover {
            background: var(--vscode-list-hoverBackground);
        }
    </style>
</head>
<body>
    <div class="chat-container">
        <div class="messages" id="messages">
            <div class="assistant-message message">
                <strong>Test Copilot</strong>
                <p>Hi! I'm your Test Automation Copilot. I can help you:</p>
                <ul>
                    <li>🎬 Record your application and generate page objects</li>
                    <li>💬 Chat about your test framework</li>
                    <li>✨ Generate test cases and page objects</li>
                    <li>🔧 Fix broken tests and update locators</li>
                </ul>
                <p>Start by recording a session or ask me anything!</p>
            </div>
        </div>
        <div class="input-container">
            <input type="text" id="messageInput" placeholder="Type your message..." />
            <button onclick="sendMessage()">Send</button>
        </div>
    </div>

    <script>
        const vscode = acquireVsCodeApi();

        function sendMessage() {
            const input = document.getElementById('messageInput');
            const message = input.value.trim();

            if (!message) return;

            // Add user message to UI
            addMessage('user', message);
            input.value = '';

            // Send to extension
            vscode.postMessage({
                type: 'chat',
                message: message
            });
        }

        function addMessage(role, content) {
            const messagesDiv = document.getElementById('messages');
            const messageDiv = document.createElement('div');
            messageDiv.className = role === 'user' ? 'user-message message' : 'assistant-message message';
            messageDiv.innerHTML = '<strong>' + (role === 'user' ? 'You' : 'Copilot') + '</strong><p>' + content + '</p>';
            messagesDiv.appendChild(messageDiv);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }

        // Handle Enter key
        document.getElementById('messageInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });

        // Handle messages from extension
        window.addEventListener('message', event => {
            const message = event.data;

            switch (message.type) {
                case 'response':
                    addMessage('assistant', message.content);
                    break;
                case 'recordingResults':
                    showRecordingResults(message.session);
                    break;
            }
        });

        function showRecordingResults(session) {
            const messagesDiv = document.getElementById('messages');
            const html = \`
                <div class="recording-summary">
                    <h3>📊 Recording Session Complete!</h3>
                    <p><strong>Pages Captured:</strong> \${session.pages.length}</p>
                    <div>
                        \${session.pages.map((page, i) => \`
                            <div class="page-item">
                                <strong>\${i + 1}. \${page.title || 'Untitled Page'}</strong><br>
                                <small>\${page.url}</small><br>
                                <small>\${page.elements.length} elements found</small>
                            </div>
                        \`).join('')}
                    </div>
                    <button onclick="generateAllPages()">Generate All Page Objects</button>
                </div>
            \`;

            const div = document.createElement('div');
            div.className = 'assistant-message message';
            div.innerHTML = html;
            messagesDiv.appendChild(div);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }

        function generateAllPages() {
            vscode.postMessage({
                type: 'generateCode',
                spec: { type: 'all-pages' }
            });
        }
    </script>
</body>
</html>`;
    }

    private async handleChatMessage(message: string): Promise<void> {
        // TODO: Send to Go binary for processing
        this.sendResponse(`You said: ${message}. (AI integration coming soon...)`);
    }

    private async handleCodeGeneration(spec: any): Promise<void> {
        // TODO: Generate code via Go binary
        this.sendResponse('Code generation will be implemented soon!');
    }

    public sendResponse(content: string): void {
        if (this._view) {
            this._view.webview.postMessage({
                type: 'response',
                content
            });
        }
    }

    public showRecordingResults(session: RecordingSession): void {
        if (this._view) {
            this._view.webview.postMessage({
                type: 'recordingResults',
                session: {
                    pages: session.pages.map(p => ({
                        title: p.title,
                        url: p.url,
                        elements: p.elements.map(e => ({
                            name: e.text || e.id || 'unnamed',
                            locator: e.recommendedLocator
                        }))
                    }))
                }
            });
        }
    }

    public sendAnalysisResults(analysis: any): void {
        this.sendResponse(
            `Framework analyzed!\n` +
            `- ${analysis.pageObjects.length} page objects\n` +
            `- ${analysis.testCases.length} test cases\n` +
            `- Framework: ${analysis.framework}\n` +
            `- Test Runner: ${analysis.testRunner}`
        );
    }

    public requestCodeGeneration(type: string, spec: string): void {
        // TODO: Implement
    }

    public async requestPageObjectGeneration(session: RecordingSession): Promise<void> {
        this.sendResponse(`Starting page object generation for ${session.pages.length} pages...`);

        try {
            // Check if Core Client is running
            if (!this.coreClient.isRunning()) {
                throw new Error('Copilot Core is not running. Please restart the extension.');
            }

            // Get workspace path
            const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
            if (!workspaceFolder) {
                throw new Error('No workspace folder open');
            }

            const workspacePath = workspaceFolder.uri.fsPath;

            // Generate code from recorded session
            const generatedFiles = await this.coreClient.generateFromSession(session, 'selenium-java');

            if (generatedFiles.length === 0) {
                this.sendResponse('❌ No files generated. Please check the logs.');
                return;
            }

            // Write generated files to workspace
            let createdCount = 0;
            for (const file of generatedFiles) {
                const fullPath = path.join(workspacePath, file.filePath);

                // Create directory if doesn't exist
                const dir = path.dirname(fullPath);
                if (!fs.existsSync(dir)) {
                    fs.mkdirSync(dir, { recursive: true });
                }

                // Write file
                fs.writeFileSync(fullPath, file.content, 'utf8');
                createdCount++;

                this.sendResponse(`✅ Created: ${file.filePath}`);
            }

            this.sendResponse(
                `\n🎉 Successfully generated ${createdCount} Page Object${createdCount > 1 ? 's' : ''}!\n` +
                `\nFiles created in: src/test/java/pages/`
            );

            // Show success message
            vscode.window.showInformationMessage(
                `Generated ${createdCount} Page Objects!`,
                'Open Files'
            ).then(selection => {
                if (selection === 'Open Files') {
                    // Open first generated file
                    if (generatedFiles.length > 0) {
                        const firstFile = path.join(workspacePath, generatedFiles[0].filePath);
                        vscode.workspace.openTextDocument(firstFile).then(doc => {
                            vscode.window.showTextDocument(doc);
                        });
                    }
                }
            });

        } catch (error) {
            this.sendResponse(`❌ Error: ${error}`);
            vscode.window.showErrorMessage(`Failed to generate Page Objects: ${error}`);
        }
    }

    public showApplicationMap(map: ApplicationMap): void {
        this.sendResponse(
            `Application Map:\n` +
            `Nodes: ${map.nodes.length}\n` +
            `Edges: ${map.edges.length}\n` +
            `(Visualization coming soon)`
        );
    }

    public show(): void {
        if (this._view) {
            this._view.show(true);
        }
    }
}
