import * as vscode from 'vscode';
import { io, Socket } from 'socket.io-client';
import { getApiClient } from '../api/client';

export class TestExecutionPanel {
  public static currentPanel: TestExecutionPanel | undefined;
  private readonly panel: vscode.WebviewPanel;
  private socket: Socket | null = null;
  private currentExecutionId: string | null = null;
  private disposables: vscode.Disposable[] = [];

  private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri) {
    this.panel = panel;

    // Set webview content
    this.panel.webview.html = this.getWebviewContent();

    // Handle messages from webview
    this.panel.webview.onDidReceiveMessage(
      (message) => {
        switch (message.command) {
          case 'executeTests':
            this.executeTests(message.data);
            break;
          case 'cancelExecution':
            this.cancelExecution(message.executionId);
            break;
        }
      },
      null,
      this.disposables,
    );

    // Clean up when panel is closed
    this.panel.onDidDispose(() => this.dispose(), null, this.disposables);
  }

  public static show(extensionUri: vscode.Uri) {
    const column = vscode.ViewColumn.Two;

    // If we already have a panel, show it
    if (TestExecutionPanel.currentPanel) {
      TestExecutionPanel.currentPanel.panel.reveal(column);
      return;
    }

    // Otherwise, create a new panel
    const panel = vscode.window.createWebviewPanel(
      'testExecution',
      'Test Execution',
      column,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
      },
    );

    TestExecutionPanel.currentPanel = new TestExecutionPanel(panel, extensionUri);
  }

  /**
   * Execute tests
   */
  private async executeTests(data: any) {
    try {
      const apiClient = getApiClient();

      // Start execution
      const response = await apiClient.post('/tests/execute', data);
      const { executionId, websocketUrl } = response.data;

      this.currentExecutionId = executionId;

      // Update UI
      this.panel.webview.postMessage({
        command: 'executionStarted',
        executionId,
      });

      // Connect to WebSocket
      this.connectWebSocket(executionId);
    } catch (error: any) {
      vscode.window.showErrorMessage(`Failed to execute tests: ${error.message}`);

      this.panel.webview.postMessage({
        command: 'executionFailed',
        error: error.message,
      });
    }
  }

  /**
   * Connect to WebSocket for real-time logs
   */
  private connectWebSocket(executionId: string) {
    // Disconnect existing socket
    if (this.socket) {
      this.socket.disconnect();
    }

    const backendUrl = process.env.BACKEND_URL || 'http://localhost:3000';

    // Connect to WebSocket
    this.socket = io(`${backendUrl}/tests`, {
      transports: ['websocket'],
    });

    this.socket.on('connect', () => {
      console.log('WebSocket connected');

      // Subscribe to execution
      this.socket!.emit('stream', { executionId });
    });

    this.socket.on('subscribed', (data) => {
      console.log('Subscribed to execution:', data.executionId);
    });

    this.socket.on('log', (message) => {
      // Forward log to webview
      this.panel.webview.postMessage({
        command: 'log',
        message,
      });

      // Show completion notification
      if (message.type === 'execution_completed') {
        const summary = message.summary;
        const passed = summary.passed;
        const failed = summary.failed;

        if (failed === 0) {
          vscode.window.showInformationMessage(
            `✅ All tests passed! (${passed}/${summary.total})`,
          );
        } else {
          vscode.window.showWarningMessage(
            `⚠️ ${failed} test(s) failed (${passed} passed)`,
          );
        }

        // Disconnect socket
        this.socket?.disconnect();
      }
    });

    this.socket.on('disconnect', () => {
      console.log('WebSocket disconnected');
    });

    this.socket.on('error', (error) => {
      console.error('WebSocket error:', error);
      vscode.window.showErrorMessage(`WebSocket error: ${error.message}`);
    });
  }

  /**
   * Cancel execution
   */
  private async cancelExecution(executionId: string) {
    try {
      const apiClient = getApiClient();
      await apiClient.delete(`/tests/cancel/${executionId}`);

      vscode.window.showInformationMessage('Test execution cancelled');

      this.panel.webview.postMessage({
        command: 'executionCancelled',
      });

      // Disconnect WebSocket
      if (this.socket) {
        this.socket.disconnect();
      }
    } catch (error: any) {
      vscode.window.showErrorMessage(`Failed to cancel execution: ${error.message}`);
    }
  }

  /**
   * Dispose panel
   */
  public dispose() {
    TestExecutionPanel.currentPanel = undefined;

    if (this.socket) {
      this.socket.disconnect();
    }

    this.panel.dispose();

    while (this.disposables.length) {
      const disposable = this.disposables.pop();
      if (disposable) {
        disposable.dispose();
      }
    }
  }

  /**
   * Get webview HTML content
   */
  private getWebviewContent(): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test Execution</title>
    <style>
        body {
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            background-color: var(--vscode-editor-background);
            padding: 20px;
            margin: 0;
        }

        h1 {
            font-size: 24px;
            margin-bottom: 20px;
            color: var(--vscode-foreground);
        }

        .form-group {
            margin-bottom: 15px;
        }

        label {
            display: block;
            margin-bottom: 5px;
            font-weight: 600;
        }

        input, select {
            width: 100%;
            padding: 8px;
            background-color: var(--vscode-input-background);
            color: var(--vscode-input-foreground);
            border: 1px solid var(--vscode-input-border);
            border-radius: 4px;
        }

        button {
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 10px 20px;
            cursor: pointer;
            border-radius: 4px;
            font-size: 14px;
            margin-right: 10px;
        }

        button:hover {
            background-color: var(--vscode-button-hoverBackground);
        }

        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }

        .button-secondary {
            background-color: var(--vscode-button-secondaryBackground);
            color: var(--vscode-button-secondaryForeground);
        }

        .button-secondary:hover {
            background-color: var(--vscode-button-secondaryHoverBackground);
        }

        #execution-section {
            display: none;
            margin-top: 30px;
        }

        #execution-status {
            padding: 15px;
            background-color: var(--vscode-editor-background);
            border: 1px solid var(--vscode-panel-border);
            border-radius: 4px;
            margin-bottom: 20px;
        }

        .status-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
        }

        .status-queued { background-color: #858585; color: #fff; }
        .status-running { background-color: #0078d4; color: #fff; }
        .status-completed { background-color: #107c10; color: #fff; }
        .status-failed { background-color: #d13438; color: #fff; }
        .status-cancelled { background-color: #e3b505; color: #000; }

        #log-container {
            background-color: #1e1e1e;
            color: #d4d4d4;
            padding: 15px;
            border-radius: 4px;
            max-height: 500px;
            overflow-y: auto;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            line-height: 1.5;
        }

        .log-entry {
            margin-bottom: 5px;
            padding: 2px 0;
        }

        .log-info { color: #4ec9b0; }
        .log-warn { color: #dcdcaa; }
        .log-error { color: #f48771; }
        .log-success { color: #4ec9b0; }

        .test-summary {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 15px;
            margin-top: 20px;
        }

        .summary-card {
            padding: 15px;
            border-radius: 4px;
            text-align: center;
        }

        .summary-card h3 {
            margin: 0 0 10px 0;
            font-size: 32px;
        }

        .summary-card p {
            margin: 0;
            font-size: 14px;
            opacity: 0.8;
        }

        .card-total { background-color: #3794ff; }
        .card-passed { background-color: #107c10; }
        .card-failed { background-color: #d13438; }
        .card-skipped { background-color: #858585; }

        .checkbox-group {
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .checkbox-group input[type="checkbox"] {
            width: auto;
            margin: 0;
        }
    </style>
</head>
<body>
    <h1>🧪 Test Execution</h1>

    <!-- Test Configuration Form -->
    <div id="config-section">
        <div class="form-group">
            <label for="projectId">Project ID:</label>
            <input type="text" id="projectId" placeholder="proj_xxx">
        </div>

        <div class="form-group">
            <label for="tests">Tests (comma-separated or *):</label>
            <input type="text" id="tests" value="*" placeholder="LoginTest, CheckoutTest or *">
        </div>

        <div class="form-group">
            <label for="environment">Environment:</label>
            <select id="environment">
                <option value="dev">Development</option>
                <option value="staging" selected>Staging</option>
                <option value="production">Production</option>
            </select>
        </div>

        <div class="form-group">
            <label for="browser">Browser:</label>
            <select id="browser">
                <option value="chrome" selected>Chrome</option>
                <option value="firefox">Firefox</option>
                <option value="safari">Safari</option>
                <option value="edge">Edge</option>
            </select>
        </div>

        <div class="form-group checkbox-group">
            <input type="checkbox" id="headless" checked>
            <label for="headless">Headless Mode</label>
        </div>

        <div class="form-group checkbox-group">
            <input type="checkbox" id="parallel">
            <label for="parallel">Parallel Execution</label>
        </div>

        <div class="form-group">
            <label for="maxWorkers">Max Workers:</label>
            <input type="number" id="maxWorkers" value="1" min="1" max="10">
        </div>

        <div class="form-group">
            <label for="retryFailed">Retry Failed Tests:</label>
            <input type="number" id="retryFailed" value="0" min="0" max="5">
        </div>

        <button id="executeBtn">🚀 Execute Tests</button>
    </div>

    <!-- Execution Status -->
    <div id="execution-section">
        <div id="execution-status">
            <h2>Execution Status</h2>
            <p>
                <span class="status-badge" id="statusBadge">Queued</span>
                <span id="executionId" style="margin-left: 10px; opacity: 0.7;"></span>
            </p>
        </div>

        <!-- Test Summary -->
        <div id="test-summary" style="display: none;">
            <div class="test-summary">
                <div class="summary-card card-total">
                    <h3 id="totalTests">0</h3>
                    <p>Total</p>
                </div>
                <div class="summary-card card-passed">
                    <h3 id="passedTests">0</h3>
                    <p>Passed</p>
                </div>
                <div class="summary-card card-failed">
                    <h3 id="failedTests">0</h3>
                    <p>Failed</p>
                </div>
                <div class="summary-card card-skipped">
                    <h3 id="skippedTests">0</h3>
                    <p>Skipped</p>
                </div>
            </div>
        </div>

        <!-- Logs -->
        <h3>Logs</h3>
        <div id="log-container"></div>

        <div style="margin-top: 20px;">
            <button id="cancelBtn" class="button-secondary">Cancel Execution</button>
            <button id="newExecutionBtn" class="button-secondary">New Execution</button>
        </div>
    </div>

    <script>
        const vscode = acquireVsCodeApi();

        const configSection = document.getElementById('config-section');
        const executionSection = document.getElementById('execution-section');
        const logContainer = document.getElementById('log-container');
        const statusBadge = document.getElementById('statusBadge');
        const executionIdSpan = document.getElementById('executionId');
        const testSummary = document.getElementById('test-summary');

        let currentExecutionId = null;

        // Execute button
        document.getElementById('executeBtn').addEventListener('click', () => {
            const projectId = document.getElementById('projectId').value;

            if (!projectId) {
                alert('Please enter a Project ID');
                return;
            }

            const testsInput = document.getElementById('tests').value;
            const tests = testsInput === '*' ? ['*'] : testsInput.split(',').map(t => t.trim());

            const data = {
                projectId,
                tests,
                environment: document.getElementById('environment').value,
                browser: document.getElementById('browser').value,
                headless: document.getElementById('headless').checked,
                parallel: document.getElementById('parallel').checked,
                maxWorkers: parseInt(document.getElementById('maxWorkers').value),
                retryFailedTests: parseInt(document.getElementById('retryFailed').value),
            };

            vscode.postMessage({
                command: 'executeTests',
                data,
            });
        });

        // Cancel button
        document.getElementById('cancelBtn').addEventListener('click', () => {
            if (currentExecutionId) {
                vscode.postMessage({
                    command: 'cancelExecution',
                    executionId: currentExecutionId,
                });
            }
        });

        // New execution button
        document.getElementById('newExecutionBtn').addEventListener('click', () => {
            configSection.style.display = 'block';
            executionSection.style.display = 'none';
            logContainer.innerHTML = '';
            testSummary.style.display = 'none';
        });

        // Handle messages from extension
        window.addEventListener('message', event => {
            const message = event.data;

            switch (message.command) {
                case 'executionStarted':
                    currentExecutionId = message.executionId;
                    configSection.style.display = 'none';
                    executionSection.style.display = 'block';
                    executionIdSpan.textContent = \`ID: \${message.executionId}\`;
                    updateStatus('queued');
                    addLog('info', '🚀 Test execution started');
                    break;

                case 'log':
                    handleLog(message.message);
                    break;

                case 'executionFailed':
                    updateStatus('failed');
                    addLog('error', \`❌ Execution failed: \${message.error}\`);
                    break;

                case 'executionCancelled':
                    updateStatus('cancelled');
                    addLog('warn', '🛑 Execution cancelled');
                    break;
            }
        });

        function handleLog(log) {
            switch (log.type) {
                case 'status':
                    updateStatus(log.status);
                    break;

                case 'log':
                    addLog(log.level || 'info', log.message);
                    break;

                case 'test_started':
                    addLog('info', \`▶️  Starting: \${log.testName}\`);
                    break;

                case 'test_completed':
                    const icon = log.status === 'passed' ? '✅' : '❌';
                    addLog(log.status === 'passed' ? 'success' : 'error',
                        \`\${icon} \${log.testName}: \${log.status} (\${log.duration}ms)\`);
                    break;

                case 'execution_completed':
                    updateStatus('completed');
                    updateSummary(log.summary);
                    addLog('success', '✨ Execution completed');
                    break;

                case 'screenshot':
                    addLog('info', \`📸 Screenshot captured: \${log.testName}\`);
                    break;
            }
        }

        function updateStatus(status) {
            statusBadge.textContent = status;
            statusBadge.className = 'status-badge status-' + status;
        }

        function addLog(level, message) {
            const entry = document.createElement('div');
            entry.className = \`log-entry log-\${level}\`;
            entry.textContent = \`[\${new Date().toLocaleTimeString()}] \${message}\`;
            logContainer.appendChild(entry);
            logContainer.scrollTop = logContainer.scrollHeight;
        }

        function updateSummary(summary) {
            testSummary.style.display = 'block';
            document.getElementById('totalTests').textContent = summary.total;
            document.getElementById('passedTests').textContent = summary.passed;
            document.getElementById('failedTests').textContent = summary.failed;
            document.getElementById('skippedTests').textContent = summary.skipped;
        }
    </script>
</body>
</html>`;
  }
}
