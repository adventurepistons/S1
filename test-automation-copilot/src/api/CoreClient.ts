import * as vscode from 'vscode';
import * as child_process from 'child_process';
import * as path from 'path';
import axios, { AxiosInstance } from 'axios';
import { RecordingSession } from '../browser/types';

export class CoreClient {
    private serverProcess?: child_process.ChildProcess;
    private axiosClient: AxiosInstance;
    private port: number = 47823;
    private isServerRunning: boolean = false;

    constructor(private extensionPath: string) {
        this.axiosClient = axios.create({
            baseURL: `http://localhost:${this.port}`,
            timeout: 30000, // 30 seconds
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

            // Start the server process
            this.serverProcess = child_process.spawn(binaryPath, ['server'], {
                env: {
                    ...process.env,
                    COPILOT_PORT: this.port.toString()
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
        if (this.serverProcess) {
            this.serverProcess.kill();
            this.serverProcess = undefined;
            this.isServerRunning = false;
            console.log('Copilot Core server stopped');
        }
    }

    /**
     * Analyze workspace
     */
    public async analyzeWorkspace(workspacePath: string): Promise<any> {
        try {
            const response = await this.axiosClient.post('/analyze', {
                workspacePath
            });
            return response.data;
        } catch (error) {
            throw this.handleError(error, 'Failed to analyze workspace');
        }
    }

    /**
     * Chat with AI
     */
    public async chat(message: string, workspacePath: string, context?: any): Promise<string> {
        try {
            const response = await this.axiosClient.post('/chat', {
                message,
                workspacePath,
                context: context || {}
            });
            return response.data.response;
        } catch (error) {
            throw this.handleError(error, 'Failed to chat with AI');
        }
    }

    /**
     * Generate code from recorded session
     */
    public async generateFromSession(session: RecordingSession, framework: string = 'selenium-java'): Promise<any[]> {
        try {
            const response = await this.axiosClient.post('/generate', {
                type: 'session',
                sessionData: {
                    pages: session.pages.map(page => ({
                        url: page.url,
                        title: page.title,
                        elements: page.elements.map(elem => ({
                            id: elem.id,
                            name: elem.text || elem.id || 'unnamed',
                            text: elem.text,
                            tagName: elem.tagName,
                            locatorType: elem.recommendedLocator.type,
                            locatorValue: elem.recommendedLocator.value,
                            interactionType: elem.interactionType,
                            attributes: elem.attributes,
                            recommendedLocator: elem.recommendedLocator
                        })),
                        interactions: page.interactions
                    })),
                    framework
                }
            });

            return response.data.generatedFiles || [];
        } catch (error) {
            throw this.handleError(error, 'Failed to generate code from session');
        }
    }

    /**
     * Generate single page object
     */
    public async generatePageObject(spec: string, elements: any[], framework: string = 'selenium-java', existingCode?: any[]): Promise<any> {
        try {
            const response = await this.axiosClient.post('/generate', {
                type: 'pageObject',
                specification: spec,
                elements,
                framework,
                existingCode: existingCode || []
            });

            return response.data;
        } catch (error) {
            throw this.handleError(error, 'Failed to generate page object');
        }
    }

    /**
     * Generate test case
     */
    public async generateTest(spec: string, pageObjects: any[], framework: string = 'selenium-java'): Promise<any> {
        try {
            const response = await this.axiosClient.post('/generate', {
                type: 'test',
                specification: spec,
                context: {
                    pageObjects
                },
                framework
            });

            return response.data;
        } catch (error) {
            throw this.handleError(error, 'Failed to generate test');
        }
    }

    /**
     * Semantic search in workspace
     */
    public async search(query: string, workspacePath: string, topK: number = 5): Promise<any[]> {
        try {
            const response = await this.axiosClient.post('/search', {
                query,
                workspacePath,
                topK
            });
            return response.data;
        } catch (error) {
            throw this.handleError(error, 'Failed to search');
        }
    }

    /**
     * Check if server is healthy
     */
    public async checkHealth(): Promise<boolean> {
        try {
            const response = await this.axiosClient.get('/health');
            return response.data.status === 'ok';
        } catch (error) {
            return false;
        }
    }

    /**
     * Wait for server to be ready
     */
    private async waitForServer(maxAttempts: number = 10): Promise<void> {
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
            path.join(this.extensionPath, 'copilot-core', 'copilot-core'),
            path.join(this.extensionPath, 'copilot-core', 'copilot-core.exe'),
            path.join(this.extensionPath, 'bin', 'copilot-core'),
            path.join(this.extensionPath, 'bin', 'copilot-core.exe'),
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
}
