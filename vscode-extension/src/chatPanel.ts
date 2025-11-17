import * as vscode from 'vscode';
import { LanguageClient } from 'vscode-languageclient/node';

export class ChatPanel {
    private panel: vscode.WebviewPanel | undefined;
    private messages: ChatMessage[] = [];
    private totalCost: number = 0;

    constructor(
        private extensionUri: vscode.Uri,
        private client: LanguageClient
    ) {}

    public show() {
        if (this.panel) {
            this.panel.reveal();
            return;
        }

        this.panel = vscode.window.createWebviewPanel(
            'testCopilotChat',
            'Test Copilot Chat',
            vscode.ViewColumn.Two,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        this.panel.webview.html = this.getWebviewContent();

        // Handle messages from webview
        this.panel.webview.onDidReceiveMessage(
            async message => {
                switch (message.command) {
                    case 'send':
                        await this.handleUserMessage(message.text);
                        break;
                    case 'save':
                        await this.handleSaveCode(message.code, message.suggestedPath);
                        break;
                    case 'insert':
                        await this.handleInsertCode(message.code);
                        break;
                }
            }
        );

        this.panel.onDidDispose(() => {
            this.panel = undefined;
        });
    }

    private async handleUserMessage(text: string) {
        // Add user message
        this.messages.push({
            role: 'user',
            content: text,
            timestamp: new Date()
        });

        this.updateWebview();

        try {
            // Show typing indicator
            this.panel?.webview.postMessage({
                command: 'typing',
                value: true
            });

            // Call backend
            const result: any = await this.client.sendRequest('testCopilot/chat', {
                request: text
            });

            // Hide typing indicator
            this.panel?.webview.postMessage({
                command: 'typing',
                value: false
            });

            // Update cost
            this.totalCost += result.cost;

            // Add assistant message
            this.messages.push({
                role: 'assistant',
                content: result.code,
                metadata: {
                    className: result.className,
                    codeType: result.codeType,
                    suggestedPath: result.suggestedPath,
                    cost: result.cost,
                    tokens: result.tokens
                },
                timestamp: new Date()
            });

            this.updateWebview();

        } catch (error) {
            this.panel?.webview.postMessage({
                command: 'typing',
                value: false
            });

            this.panel?.webview.postMessage({
                command: 'error',
                message: `Error: ${error}`
            });
        }
    }

    private async handleSaveCode(code: string, suggestedPath: string) {
        const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
        if (!workspaceRoot) {
            vscode.window.showErrorMessage('No workspace folder open');
            return;
        }

        const fullPath = `${workspaceRoot}/src/test/java/${suggestedPath}`;

        const uri = await vscode.window.showSaveDialog({
            defaultUri: vscode.Uri.file(fullPath),
            filters: { 'Java': ['java'] }
        });

        if (uri) {
            await vscode.workspace.fs.writeFile(uri, Buffer.from(code, 'utf8'));
            vscode.window.showInformationMessage(`Saved to ${uri.fsPath}`);

            const document = await vscode.workspace.openTextDocument(uri);
            await vscode.window.showTextDocument(document);
        }
    }

    private async handleInsertCode(code: string) {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
            vscode.window.showErrorMessage('No active editor');
            return;
        }

        await editor.edit(editBuilder => {
            editBuilder.insert(editor.selection.active, code);
        });

        vscode.window.showInformationMessage('Code inserted!');
    }

    private updateWebview() {
        if (!this.panel) {
            return;
        }

        this.panel.webview.postMessage({
            command: 'update',
            messages: this.messages,
            totalCost: this.totalCost
        });
    }

    private getWebviewContent(): string {
        return `
            <!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
                <title>Test Copilot Chat</title>
                <style>
                    * {
                        box-sizing: border-box;
                    }

                    body {
                        font-family: var(--vscode-font-family);
                        padding: 0;
                        margin: 0;
                        background-color: var(--vscode-editor-background);
                        color: var(--vscode-editor-foreground);
                        display: flex;
                        flex-direction: column;
                        height: 100vh;
                    }

                    #header {
                        padding: 15px;
                        background-color: var(--vscode-titleBar-activeBackground);
                        border-bottom: 1px solid var(--vscode-panel-border);
                    }

                    #header h2 {
                        margin: 0;
                        font-size: 16px;
                        color: var(--vscode-titleBar-activeForeground);
                    }

                    #cost {
                        font-size: 12px;
                        color: var(--vscode-descriptionForeground);
                        margin-top: 5px;
                    }

                    #messages {
                        flex: 1;
                        overflow-y: auto;
                        padding: 20px;
                    }

                    .message {
                        margin-bottom: 20px;
                        display: flex;
                        flex-direction: column;
                    }

                    .message-header {
                        font-weight: bold;
                        margin-bottom: 5px;
                        font-size: 12px;
                        color: var(--vscode-textLink-foreground);
                    }

                    .message-content {
                        padding: 12px;
                        border-radius: 6px;
                        white-space: pre-wrap;
                        word-wrap: break-word;
                    }

                    .user .message-content {
                        background-color: var(--vscode-input-background);
                        border: 1px solid var(--vscode-input-border);
                    }

                    .assistant .message-content {
                        background-color: var(--vscode-editor-inactiveSelectionBackground);
                        font-family: var(--vscode-editor-font-family);
                        font-size: 13px;
                    }

                    .message-metadata {
                        margin-top: 10px;
                        padding: 8px;
                        background-color: var(--vscode-editorWidget-background);
                        border-radius: 4px;
                        font-size: 11px;
                        color: var(--vscode-descriptionForeground);
                    }

                    .message-actions {
                        margin-top: 8px;
                        display: flex;
                        gap: 8px;
                    }

                    button {
                        padding: 6px 12px;
                        background-color: var(--vscode-button-background);
                        color: var(--vscode-button-foreground);
                        border: none;
                        border-radius: 3px;
                        cursor: pointer;
                        font-size: 12px;
                    }

                    button:hover {
                        background-color: var(--vscode-button-hoverBackground);
                    }

                    #input-container {
                        padding: 15px;
                        background-color: var(--vscode-editor-background);
                        border-top: 1px solid var(--vscode-panel-border);
                    }

                    #input-box {
                        display: flex;
                        gap: 10px;
                    }

                    #user-input {
                        flex: 1;
                        padding: 10px;
                        background-color: var(--vscode-input-background);
                        color: var(--vscode-input-foreground);
                        border: 1px solid var(--vscode-input-border);
                        border-radius: 4px;
                        font-family: var(--vscode-font-family);
                        resize: none;
                    }

                    #send-btn {
                        padding: 10px 20px;
                    }

                    #typing {
                        display: none;
                        padding: 10px;
                        font-style: italic;
                        color: var(--vscode-descriptionForeground);
                    }

                    .error {
                        color: var(--vscode-errorForeground);
                        padding: 10px;
                        background-color: var(--vscode-inputValidation-errorBackground);
                        border: 1px solid var(--vscode-inputValidation-errorBorder);
                        border-radius: 4px;
                        margin: 10px 0;
                    }
                </style>
            </head>
            <body>
                <div id="header">
                    <h2>🤖 Test Automation Copilot</h2>
                    <div id="cost">Session cost: $0.00</div>
                </div>

                <div id="messages"></div>

                <div id="typing">🤖 Generating code...</div>

                <div id="input-container">
                    <div id="input-box">
                        <textarea
                            id="user-input"
                            rows="3"
                            placeholder="What would you like to generate? (e.g., Create test for login)"
                        ></textarea>
                        <button id="send-btn">Send</button>
                    </div>
                </div>

                <script>
                    const vscode = acquireVsCodeApi();
                    const messagesDiv = document.getElementById('messages');
                    const userInput = document.getElementById('user-input');
                    const sendBtn = document.getElementById('send-btn');
                    const typingDiv = document.getElementById('typing');
                    const costDiv = document.getElementById('cost');

                    let messages = [];
                    let totalCost = 0;

                    sendBtn.addEventListener('click', sendMessage);
                    userInput.addEventListener('keypress', (e) => {
                        if (e.key === 'Enter' && e.ctrlKey) {
                            sendMessage();
                        }
                    });

                    function sendMessage() {
                        const text = userInput.value.trim();
                        if (!text) return;

                        vscode.postMessage({
                            command: 'send',
                            text: text
                        });

                        userInput.value = '';
                    }

                    function saveCode(code, suggestedPath) {
                        vscode.postMessage({
                            command: 'save',
                            code: code,
                            suggestedPath: suggestedPath
                        });
                    }

                    function insertCode(code) {
                        vscode.postMessage({
                            command: 'insert',
                            code: code
                        });
                    }

                    function renderMessages() {
                        messagesDiv.innerHTML = '';

                        messages.forEach((msg, idx) => {
                            const msgDiv = document.createElement('div');
                            msgDiv.className = \`message \${msg.role}\`;

                            const header = document.createElement('div');
                            header.className = 'message-header';
                            header.textContent = msg.role === 'user' ? 'You' : '🤖 Test Copilot';
                            msgDiv.appendChild(header);

                            const content = document.createElement('div');
                            content.className = 'message-content';
                            content.textContent = msg.content;
                            msgDiv.appendChild(content);

                            if (msg.metadata) {
                                const meta = document.createElement('div');
                                meta.className = 'message-metadata';
                                meta.innerHTML = \`
                                    📄 <strong>\${msg.metadata.className}</strong> (\${msg.metadata.codeType})<br>
                                    💰 Cost: $\${msg.metadata.cost.toFixed(4)} |
                                    📊 Tokens: \${msg.metadata.tokens.input} in / \${msg.metadata.tokens.output} out
                                \`;
                                msgDiv.appendChild(meta);

                                const actions = document.createElement('div');
                                actions.className = 'message-actions';
                                actions.innerHTML = \`
                                    <button onclick="saveCode(\${idx})">💾 Save to File</button>
                                    <button onclick="insertCode(\${idx})">📝 Insert at Cursor</button>
                                \`;
                                msgDiv.appendChild(actions);
                            }

                            messagesDiv.appendChild(msgDiv);
                        });

                        messagesDiv.scrollTop = messagesDiv.scrollHeight;
                    }

                    window.saveCode = function(idx) {
                        const msg = messages[idx];
                        vscode.postMessage({
                            command: 'save',
                            code: msg.content,
                            suggestedPath: msg.metadata.suggestedPath
                        });
                    };

                    window.insertCode = function(idx) {
                        const msg = messages[idx];
                        vscode.postMessage({
                            command: 'insert',
                            code: msg.content
                        });
                    };

                    window.addEventListener('message', event => {
                        const message = event.data;

                        switch (message.command) {
                            case 'update':
                                messages = message.messages;
                                totalCost = message.totalCost;
                                costDiv.textContent = \`Session cost: $\${totalCost.toFixed(2)}\`;
                                renderMessages();
                                break;

                            case 'typing':
                                typingDiv.style.display = message.value ? 'block' : 'none';
                                break;

                            case 'error':
                                const errorDiv = document.createElement('div');
                                errorDiv.className = 'error';
                                errorDiv.textContent = message.message;
                                messagesDiv.appendChild(errorDiv);
                                break;
                        }
                    });
                </script>
            </body>
            </html>
        `;
    }
}

interface ChatMessage {
    role: 'user' | 'assistant';
    content: string;
    metadata?: {
        className: string;
        codeType: string;
        suggestedPath: string;
        cost: number;
        tokens: {
            input: number;
            output: number;
            cached: number;
        };
    };
    timestamp: Date;
}
