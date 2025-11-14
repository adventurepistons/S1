import * as vscode from 'vscode';
import { BrowserRecorder } from './BrowserRecorder';
import { RecordingSession } from './types';
import { ChatPanelProvider } from '../ui/ChatPanelProvider';

export class ElementRecorder {
    private recorder: BrowserRecorder;
    private chatProvider: ChatPanelProvider;

    constructor(chatProvider: ChatPanelProvider) {
        this.recorder = new BrowserRecorder();
        this.chatProvider = chatProvider;
    }

    public async startRecording(url?: string): Promise<void> {
        try {
            // Check if already recording
            if (this.recorder.isCurrentlyRecording()) {
                const choice = await vscode.window.showWarningMessage(
                    'Recording already in progress. Stop current recording?',
                    'Stop Current',
                    'Cancel'
                );

                if (choice === 'Stop Current') {
                    await this.stopRecording();
                } else {
                    return;
                }
            }

            // Get URL if not provided
            if (!url) {
                url = await vscode.window.showInputBox({
                    prompt: 'Enter starting URL (optional - you can navigate manually)',
                    placeHolder: 'https://example.com',
                    validateInput: (value) => {
                        if (!value) return null; // Optional
                        try {
                            new URL(value);
                            return null;
                        } catch {
                            return 'Please enter a valid URL';
                        }
                    }
                });
            }

            // Show recording info
            await vscode.window.showInformationMessage(
                '🎬 Starting recording session...\n' +
                'A browser will open. Navigate your application normally.\n' +
                'All pages and interactions will be recorded.',
                { modal: false }
            );

            // Start recording
            await this.recorder.startRecording(url);

            // Show status bar item
            this.showRecordingStatus();

        } catch (error) {
            vscode.window.showErrorMessage(`Failed to start recording: ${error}`);
        }
    }

    public async stopRecording(): Promise<void> {
        try {
            if (!this.recorder.isCurrentlyRecording()) {
                vscode.window.showWarningMessage('No recording in progress');
                return;
            }

            // Stop recording
            const session = await this.recorder.stopRecording();

            // Hide status bar
            this.hideRecordingStatus();

            // Show session summary
            await this.showSessionSummary(session);

            // Send to chat provider for display
            this.chatProvider.showRecordingResults(session);

        } catch (error) {
            vscode.window.showErrorMessage(`Failed to stop recording: ${error}`);
        }
    }

    public async pauseRecording(): Promise<void> {
        try {
            await this.recorder.pauseRecording();
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to pause recording: ${error}`);
        }
    }

    public async resumeRecording(): Promise<void> {
        try {
            await this.recorder.resumeRecording();
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to resume recording: ${error}`);
        }
    }

    private async showSessionSummary(session: RecordingSession): Promise<void> {
        const duration = session.duration ? Math.round(session.duration / 1000) : 0;
        const totalElements = session.pages.reduce((sum, page) => sum + page.elements.length, 0);
        const totalInteractions = session.pages.reduce((sum, page) => sum + page.interactions.length, 0);

        const summary = `
📊 Recording Session Complete!

⏱️ Duration: ${duration} seconds
📄 Pages Captured: ${session.pages.length}
🎯 Elements Found: ${totalElements}
👆 Interactions: ${totalInteractions}

Pages:
${session.pages.map((page, i) => `${i + 1}. ${page.title} (${page.elements.length} elements)`).join('\n')}

What would you like to do?
        `.trim();

        const choice = await vscode.window.showInformationMessage(
            summary,
            { modal: true },
            'Generate Page Objects',
            'View Application Map',
            'Save Session'
        );

        if (choice === 'Generate Page Objects') {
            this.chatProvider.requestPageObjectGeneration(session);
        } else if (choice === 'View Application Map') {
            this.chatProvider.showApplicationMap(session.applicationMap);
        } else if (choice === 'Save Session') {
            await this.saveSession(session);
        }
    }

    private async saveSession(session: RecordingSession): Promise<void> {
        const uri = await vscode.window.showSaveDialog({
            defaultUri: vscode.Uri.file(`recording-${Date.now()}.json`),
            filters: {
                'JSON files': ['json']
            }
        });

        if (uri) {
            const fs = require('fs');
            fs.writeFileSync(uri.fsPath, JSON.stringify(session, null, 2));
            vscode.window.showInformationMessage(`Session saved to ${uri.fsPath}`);
        }
    }

    private statusBarItem?: vscode.StatusBarItem;

    private showRecordingStatus(): void {
        this.statusBarItem = vscode.window.createStatusBarItem(
            vscode.StatusBarAlignment.Left,
            100
        );

        this.statusBarItem.text = '$(circle-filled) Recording...';
        this.statusBarItem.tooltip = 'Recording session in progress. Click to stop.';
        this.statusBarItem.command = 'testCopilot.stopRecording';
        this.statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.errorBackground');
        this.statusBarItem.show();

        // Animate recording indicator
        let frame = 0;
        const interval = setInterval(() => {
            if (!this.recorder.isCurrentlyRecording()) {
                clearInterval(interval);
                return;
            }

            const frames = ['⚫', '⚪'];
            this.statusBarItem!.text = `${frames[frame % 2]} Recording...`;
            frame++;
        }, 1000);
    }

    private hideRecordingStatus(): void {
        if (this.statusBarItem) {
            this.statusBarItem.dispose();
            this.statusBarItem = undefined;
        }
    }

    public getRecorder(): BrowserRecorder {
        return this.recorder;
    }
}
