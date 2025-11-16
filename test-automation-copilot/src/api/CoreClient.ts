import * as vscode from 'vscode';
import * as child_process from 'child_process';
import * as path from 'path';
import axios, { AxiosInstance } from 'axios';
import WebSocket from 'ws';
import { RecordingSession } from '../browser/types';

export interface WorkspaceStats {
    files: number;
    classes: number;
    pageObjects: number;
    testMethods: number;
    vectorDocs: number;
    collectionName: string;
}

export interface SearchResult {
    id: string;
    name: string;
    content: string;
    filePath: string;
    type: string;
    similarity: number;
}

export interface GenerationResult {
    code: string;
    tokensUsed: number;
    model: string;
    finishReason: string;
}

export interface ChatResponse {
    response: string;
    tokensUsed: number;
    model: string;
    finishReason: string;
}

export interface UsageStats {
    userId: string;
    plan: string;
    period: {
        start: number;
        end: number;
        current: number;
    };
    usage: {
        totalRequests: number;
        totalTokens: number;
        estimatedCost: number;
        requestsThisMonth: number;
        tokensThisMonth: number;
        requestsByAction: {
            pageobject: number;
            test: number;
            chat: number;
            fix: number;
        };
    };
    limits: {
        requestsPerMonth: number;
        tokensPerMonth: number;
        requestsRemaining: number;
        tokensRemaining: number;
    };
}

export class CoreClient {
    private serverProcess?: child_process.ChildProcess;
    private axiosClient: AxiosInstance;
    private wsClient?: WebSocket;
    private port: number = 8080;
    private isServerRunning: boolean = false;

    constructor(private extensionPath: string, private context: vscode.ExtensionContext) {
        // Read port from configuration
        const config = vscode.workspace.getConfiguration('testCopilot');
        this.port = config.get<number>('localBackendPort') || 8080;

        this.axiosClient = axios.create({
            baseURL: `http://localhost:${this.port}`,
            timeout: 60000, // 60 seconds for LLM operations
            headers: {
                'Content-Type': 'application/json'
            }
        });
    }

    /**
     * Start the Go binary server
     */
    public async startServer(): Promise<void> {
        if (this.isServerRunning) {
            console.log('Server already running');
            return;
        }

        try {
            // Find the Go binary
            const binaryPath = this.findBinary();

            if (!binaryPath) {
                throw new Error('Copilot core binary not found. Please build the Go binary first.');
            }

            // Get cloud configuration
            const config = vscode.workspace.getConfiguration('testCopilot');
            const cloudUrl = config.get<string>('cloudUrl') || 'https://api.testcopilot.ai';
            const apiKey = config.get<string>('apiKey') || '';

            // Start the server process
            this.serverProcess = child_process.spawn(binaryPath, [
                '--port', this.port.toString(),
                '--db', path.join(this.extensionPath, 'data', 'copilot.db'),
                '--vector', path.join(this.extensionPath, 'data', 'vectordb')
            ], {
                env: {
                    ...process.env,
                    TESTCOPILOT_CLOUD_URL: cloudUrl,
                    TESTCOPILOT_API_KEY: apiKey
                }
            });

            // Handle server output
            this.serverProcess.stdout?.on('data', (data) => {
                console.log(`[Copilot Core] ${data}`);
            });

            this.serverProcess.stderr?.on('data', (data) => {
                console.error(`[Copilot Core Error] ${data}`);
            });

            this.serverProcess.on('exit', (code) => {
                console.log(`Copilot Core server exited with code ${code}`);
                this.isServerRunning = false;
            });

            // Wait for server to be ready
            await this.waitForServer();
            this.isServerRunning = true;

            console.log('Copilot Core server started successfully');

        } catch (error) {
            throw new Error(`Failed to start Copilot Core server: ${error}`);
        }
    }

    /**
     * Stop the Go binary server
     */
    public stopServer(): void {
        if (this.wsClient) {
            this.wsClient.close();
            this.wsClient = undefined;
        }

        if (this.serverProcess) {
            this.serverProcess.kill();
            this.serverProcess = undefined;
            this.isServerRunning = false;
            console.log('Copilot Core server stopped');
        }
    }

    /**
     * Index workspace
     */
    public async indexWorkspace(workspacePath: string): Promise<void> {
        try {
            await this.axiosClient.post('/api/v1/workspace/index', {
                workspacePath
            });
        } catch (error) {
            throw this.handleError(error, 'Failed to index workspace');
        }
    }

    /**
     * Get workspace statistics
     */
    public async getWorkspaceStats(): Promise<WorkspaceStats> {
        try {
            const response = await this.axiosClient.get('/api/v1/workspace/stats');
            return response.data;
        } catch (error) {
            throw this.handleError(error, 'Failed to get workspace stats');
        }
    }

    /**
     * Search code semantically
     */
    public async searchCode(query: string, limit: number = 10): Promise<SearchResult[]> {
        try {
            const response = await this.axiosClient.post('/api/v1/search/code', {
                query,
                limit
            });
            return response.data.results || [];
        } catch (error) {
            throw this.handleError(error, 'Failed to search code');
        }
    }

    /**
     * Search page objects
     */
    public async searchPageObjects(query: string, limit: number = 5): Promise<SearchResult[]> {
        try {
            const response = await this.axiosClient.post('/api/v1/search/pageobjects', {
                query,
                limit
            });
            return response.data.results || [];
        } catch (error) {
            throw this.handleError(error, 'Failed to search page objects');
        }
    }

    /**
     * Search tests
     */
    public async searchTests(query: string, limit: number = 5): Promise<SearchResult[]> {
        try {
            const response = await this.axiosClient.post('/api/v1/search/tests', {
                query,
                limit
            });
            return response.data.results || [];
        } catch (error) {
            throw this.handleError(error, 'Failed to search tests');
        }
    }

    /**
     * Generate page object
     */
    public async generatePageObject(spec: string, elements: any[]): Promise<GenerationResult> {
        try {
            const response = await this.axiosClient.post('/api/v1/generate/pageobject', {
                spec,
                elements
            });
            return {
                code: response.data.code,
                tokensUsed: response.data.tokensUsed,
                model: response.data.model,
                finishReason: response.data.finishReason
            };
        } catch (error) {
            throw this.handleError(error, 'Failed to generate page object');
        }
    }

    /**
     * Generate test case
     */
    public async generateTest(spec: string): Promise<GenerationResult> {
        try {
            const response = await this.axiosClient.post('/api/v1/generate/test', {
                spec
            });
            return {
                code: response.data.code,
                tokensUsed: response.data.tokensUsed,
                model: response.data.model,
                finishReason: response.data.finishReason
            };
        } catch (error) {
            throw this.handleError(error, 'Failed to generate test');
        }
    }

    /**
     * Fix broken code
     */
    public async fixCode(code: string, error: string): Promise<GenerationResult> {
        try {
            const response = await this.axiosClient.post('/api/v1/generate/fix', {
                code,
                error
            });
            return {
                code: response.data.fix,
                tokensUsed: response.data.tokensUsed,
                model: response.data.model,
                finishReason: response.data.finishReason
            };
        } catch (error) {
            throw this.handleError(error, 'Failed to fix code');
        }
    }

    /**
     * Chat with AI
     */
    public async chat(message: string): Promise<ChatResponse> {
        try {
            const response = await this.axiosClient.post('/api/v1/chat/message', {
                message
            });
            return {
                response: response.data.response,
                tokensUsed: response.data.tokensUsed,
                model: response.data.model,
                finishReason: response.data.finishReason
            };
        } catch (error) {
            throw this.handleError(error, 'Failed to chat with AI');
        }
    }

    /**
     * Get all classes from database
     */
    public async getClasses(): Promise<any[]> {
        try {
            const response = await this.axiosClient.get('/api/v1/db/classes');
            return response.data.classes || [];
        } catch (error) {
            throw this.handleError(error, 'Failed to get classes');
        }
    }

    /**
     * Get specific class
     */
    public async getClass(id: number): Promise<any> {
        try {
            const response = await this.axiosClient.get(`/api/v1/db/classes/${id}`);
            return response.data.class;
        } catch (error) {
            throw this.handleError(error, 'Failed to get class');
        }
    }

    /**
     * Get methods by class ID
     */
    public async getMethodsByClass(classId: number): Promise<any[]> {
        try {
            const response = await this.axiosClient.get(`/api/v1/db/methods/${classId}`);
            return response.data.methods || [];
        } catch (error) {
            throw this.handleError(error, 'Failed to get methods');
        }
    }

    /**
     * Generate code with streaming (WebSocket)
     */
    public async generateWithStreaming(
        action: 'pageobject' | 'test' | 'chat' | 'fix',
        payload: any,
        onChunk: (chunk: string) => void,
        onComplete: (result: GenerationResult) => void,
        onError: (error: string) => void
    ): Promise<void> {
        return new Promise((resolve, reject) => {
            try {
                // Create WebSocket connection
                this.wsClient = new WebSocket(`ws://localhost:${this.port}/ws/stream`);

                this.wsClient.on('open', () => {
                    // Send request
                    const request = {
                        type: 'request',
                        payload: {
                            action,
                            ...payload
                        }
                    };
                    this.wsClient!.send(JSON.stringify(request));
                });

                this.wsClient.on('message', (data: WebSocket.Data) => {
                    try {
                        const message = JSON.parse(data.toString());

                        switch (message.type) {
                            case 'chunk':
                                onChunk(message.payload.chunk);
                                break;

                            case 'complete':
                                const result: GenerationResult = {
                                    code: message.payload.content,
                                    tokensUsed: message.payload.tokensUsed,
                                    model: message.payload.model,
                                    finishReason: message.payload.finishReason
                                };
                                onComplete(result);
                                this.wsClient?.close();
                                resolve();
                                break;

                            case 'error':
                                onError(message.payload.error);
                                this.wsClient?.close();
                                reject(new Error(message.payload.error));
                                break;
                        }
                    } catch (err) {
                        console.error('Failed to parse WebSocket message:', err);
                    }
                });

                this.wsClient.on('error', (err: Error) => {
                    onError(err.message);
                    reject(err);
                });

                this.wsClient.on('close', () => {
                    this.wsClient = undefined;
                });

            } catch (error) {
                reject(error);
            }
        });
    }

    /**
     * Generate code from recorded session (legacy compatibility)
     */
    public async generateFromSession(session: RecordingSession, framework: string = 'selenium-java'): Promise<any[]> {
        const generatedFiles: any[] = [];

        // Generate page objects for each page
        for (const page of session.pages) {
            try {
                const result = await this.generatePageObject(
                    `Page object for ${page.title || page.url}`,
                    page.elements.map(elem => ({
                        name: elem.text || elem.id || 'unnamed',
                        locatorType: elem.recommendedLocator.type,
                        locatorValue: elem.recommendedLocator.value
                    }))
                );

                generatedFiles.push({
                    type: 'pageObject',
                    name: this.getPageObjectName(page.title || page.url),
                    code: result.code
                });
            } catch (error) {
                console.error(`Failed to generate page object for ${page.url}:`, error);
            }
        }

        return generatedFiles;
    }

    /**
     * Get usage statistics from cloud backend
     */
    public async getUsageStats(): Promise<UsageStats | null> {
        try {
            // This would call the cloud backend through local backend
            // For now, returning null as cloud backend is not implemented yet
            // TODO: Implement once cloud backend is ready
            return null;
        } catch (error) {
            console.error('Failed to get usage stats:', error);
            return null;
        }
    }

    /**
     * Check if server is healthy
     */
    public async checkHealth(): Promise<boolean> {
        try {
            const response = await this.axiosClient.get('/health');
            return response.data.status === 'healthy';
        } catch (error) {
            return false;
        }
    }

    /**
     * Wait for server to be ready
     */
    private async waitForServer(maxAttempts: number = 30): Promise<void> {
        for (let i = 0; i < maxAttempts; i++) {
            await this.sleep(1000); // Wait 1 second

            if (await this.checkHealth()) {
                return;
            }
        }

        throw new Error('Server failed to start within timeout');
    }

    /**
     * Find the Go binary
     */
    private findBinary(): string | null {
        const possiblePaths = [
            // Development paths
            path.join(this.extensionPath, 'copilot-core', 'cmd', 'server', 'server'),
            path.join(this.extensionPath, 'copilot-core', 'cmd', 'server', 'server.exe'),
            // Production paths
            path.join(this.extensionPath, 'bin', 'test-copilot-server'),
            path.join(this.extensionPath, 'bin', 'test-copilot-server.exe'),
            path.join(this.extensionPath, 'copilot-core', 'test-copilot-server'),
            path.join(this.extensionPath, 'copilot-core', 'test-copilot-server.exe'),
        ];

        for (const p of possiblePaths) {
            if (require('fs').existsSync(p)) {
                return p;
            }
        }

        return null;
    }

    /**
     * Handle errors
     */
    private handleError(error: any, message: string): Error {
        if (axios.isAxiosError(error)) {
            const errorMessage = error.response?.data?.error || error.message;
            return new Error(`${message}: ${errorMessage}`);
        }
        return new Error(`${message}: ${error}`);
    }

    /**
     * Sleep helper
     */
    private sleep(ms: number): Promise<void> {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    /**
     * Get server status
     */
    public isRunning(): boolean {
        return this.isServerRunning;
    }

    /**
     * Get page object name from URL/title
     */
    private getPageObjectName(urlOrTitle: string): string {
        // Convert URL or title to PascalCase class name
        const cleaned = urlOrTitle
            .replace(/https?:\/\//, '')
            .replace(/[^a-zA-Z0-9]/g, ' ')
            .trim();

        const words = cleaned.split(/\s+/);
        const pascalCase = words
            .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
            .join('');

        return `${pascalCase}Page`;
    }
}

