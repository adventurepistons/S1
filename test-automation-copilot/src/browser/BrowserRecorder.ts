import * as playwright from 'playwright';
import * as vscode from 'vscode';
import { RecordingSession, RecordedPage, RecordedElement, RecordedInteraction } from './types';
import { ElementAnalyzer } from './ElementAnalyzer';

export class BrowserRecorder {
    private browser?: playwright.Browser;
    private context?: playwright.BrowserContext;
    private page?: playwright.Page;
    private session?: RecordingSession;
    private isRecording: boolean = false;
    private elementAnalyzer: ElementAnalyzer;
    private currentPageData?: RecordedPage;
    private recordingStartTime?: number;

    constructor() {
        this.elementAnalyzer = new ElementAnalyzer();
    }

    public async startRecording(startUrl?: string): Promise<void> {
        if (this.isRecording) {
            throw new Error('Recording already in progress');
        }

        // Launch browser
        this.browser = await playwright.chromium.launch({
            headless: false,
            args: ['--start-maximized']
        });

        this.context = await this.browser.newContext({
            viewport: null, // Use full screen
            recordVideo: {
                dir: './recordings/videos',
                size: { width: 1920, height: 1080 }
            }
        });

        this.page = await this.context.newPage();

        // Initialize session
        this.session = {
            id: this.generateSessionId(),
            startTime: Date.now(),
            pages: [],
            userFlows: [],
            applicationMap: {
                nodes: [],
                edges: []
            }
        };

        this.isRecording = true;
        this.recordingStartTime = Date.now();

        // Inject recorder script
        await this.injectRecorderScript();

        // Set up event listeners
        await this.setupEventListeners();

        // Navigate to start URL if provided
        if (startUrl) {
            await this.page.goto(startUrl);
        }

        vscode.window.showInformationMessage('🔴 Recording session started! Navigate your application normally.');
    }

    public async stopRecording(): Promise<RecordingSession> {
        if (!this.isRecording || !this.session) {
            throw new Error('No recording in progress');
        }

        // Save current page if any
        if (this.currentPageData) {
            this.session.pages.push(this.currentPageData);
        }

        this.session.endTime = Date.now();
        this.session.duration = this.session.endTime - this.session.startTime;

        this.isRecording = false;

        // Close browser
        if (this.browser) {
            await this.browser.close();
        }

        vscode.window.showInformationMessage(
            `✅ Recording stopped! Captured ${this.session.pages.length} pages with ${this.getTotalElements()} elements.`
        );

        return this.session;
    }

    public async pauseRecording(): Promise<void> {
        if (!this.isRecording) {
            throw new Error('No recording in progress');
        }

        this.isRecording = false;
        vscode.window.showInformationMessage('⏸️ Recording paused');
    }

    public async resumeRecording(): Promise<void> {
        this.isRecording = true;
        vscode.window.showInformationMessage('▶️ Recording resumed');
    }

    private async injectRecorderScript(): Promise<void> {
        if (!this.page) return;

        // Inject element capture and interaction tracking script
        await this.page.addInitScript(() => {
            // This runs in browser context
            (window as any).__testCopilotRecorder = {
                interactions: [],
                elements: new Map(),

                captureElement(element: Element): any {
                    if (!element) return null;

                    const rect = element.getBoundingClientRect();
                    const computedStyle = window.getComputedStyle(element);

                    return {
                        tagName: element.tagName.toLowerCase(),
                        id: (element as HTMLElement).id || null,
                        className: element.className || null,
                        name: (element as any).name || null,
                        type: (element as any).type || null,
                        text: element.textContent?.trim().substring(0, 100) || null,
                        placeholder: (element as any).placeholder || null,
                        value: (element as any).value || null,
                        href: (element as any).href || null,
                        visible: rect.width > 0 && rect.height > 0 && computedStyle.visibility !== 'hidden',
                        position: {
                            x: rect.x,
                            y: rect.y,
                            width: rect.width,
                            height: rect.height
                        },
                        attributes: this.getRelevantAttributes(element),
                        path: this.getElementPath(element)
                    };
                },

                getRelevantAttributes(element: Element): Record<string, string> {
                    const attrs: Record<string, string> = {};
                    const relevantAttrs = [
                        'data-testid', 'data-test', 'data-qa',
                        'aria-label', 'role', 'title',
                        'alt', 'for', 'name'
                    ];

                    relevantAttrs.forEach(attr => {
                        const value = element.getAttribute(attr);
                        if (value) attrs[attr] = value;
                    });

                    return attrs;
                },

                getElementPath(element: Element): string {
                    const path = [];
                    let current: Element | null = element;

                    while (current && current !== document.body) {
                        let selector = current.tagName.toLowerCase();

                        if (current.id) {
                            selector += `#${current.id}`;
                            path.unshift(selector);
                            break; // ID is unique, stop here
                        }

                        if (current.className) {
                            const classes = current.className.split(' ').filter(c => c);
                            if (classes.length > 0) {
                                selector += `.${classes[0]}`;
                            }
                        }

                        path.unshift(selector);
                        current = current.parentElement;
                    }

                    return path.join(' > ');
                },

                recordInteraction(event: Event): void {
                    const element = event.target as Element;
                    const interaction = {
                        type: event.type,
                        timestamp: Date.now(),
                        element: this.captureElement(element),
                        url: window.location.href
                    };

                    this.interactions.push(interaction);

                    // Store in window for extension to retrieve
                    (window as any).__testCopilotLastInteraction = interaction;
                }
            };

            // Track all interactive events
            const eventsToTrack = ['click', 'input', 'change', 'submit', 'focus', 'keyup'];

            eventsToTrack.forEach(eventType => {
                document.addEventListener(eventType, (event) => {
                    (window as any).__testCopilotRecorder.recordInteraction(event);
                }, true);
            });

            // Track navigation
            let lastUrl = window.location.href;
            setInterval(() => {
                if (window.location.href !== lastUrl) {
                    (window as any).__testCopilotNavigation = {
                        from: lastUrl,
                        to: window.location.href,
                        timestamp: Date.now()
                    };
                    lastUrl = window.location.href;
                }
            }, 500);
        });
    }

    private async setupEventListeners(): Promise<void> {
        if (!this.page) return;

        // Listen for page navigation
        this.page.on('load', async () => {
            if (!this.isRecording) return;

            // Save previous page data
            if (this.currentPageData) {
                this.session!.pages.push(this.currentPageData);

                // Add edge to application map
                this.session!.applicationMap.edges.push({
                    from: this.session!.applicationMap.nodes.length - 1,
                    to: this.session!.applicationMap.nodes.length,
                    label: 'navigated'
                });
            }

            // Start capturing new page
            await this.capturePage();
        });

        // Listen for console messages (from our injected script)
        this.page.on('console', async (msg) => {
            if (msg.text().includes('__testCopilotInteraction')) {
                await this.handleInteraction();
            }
        });

        // Periodically capture elements
        setInterval(async () => {
            if (this.isRecording && this.page) {
                await this.captureInteractions();
            }
        }, 1000);
    }

    private async capturePage(): Promise<void> {
        if (!this.page || !this.isRecording) return;

        const url = this.page.url();
        const title = await this.page.title();

        // Capture all elements on page
        const elements = await this.captureAllElements();

        this.currentPageData = {
            id: this.generatePageId(),
            url,
            title,
            timestamp: Date.now(),
            elements,
            interactions: [],
            screenshot: await this.page.screenshot({ fullPage: true, type: 'png' }),
            domSnapshot: await this.page.content()
        };

        // Add to application map
        this.session!.applicationMap.nodes.push({
            id: this.currentPageData.id,
            label: this.inferPageName(url, title),
            url,
            elementCount: elements.length
        });

        console.log(`Captured page: ${url} with ${elements.length} elements`);
    }

    private async captureAllElements(): Promise<RecordedElement[]> {
        if (!this.page) return [];

        // Get all interactive elements
        const elementSelectors = [
            'input',
            'button',
            'a',
            'select',
            'textarea',
            '[role="button"]',
            '[onclick]',
            '[data-testid]',
            '[data-test]'
        ];

        const elements: RecordedElement[] = [];

        for (const selector of elementSelectors) {
            try {
                const elementHandles = await this.page.$$(selector);

                for (const handle of elementHandles) {
                    const elementData = await handle.evaluate((el: any, analyzer: any) => {
                        return (window as any).__testCopilotRecorder.captureElement(el);
                    });

                    if (elementData && elementData.visible) {
                        // Generate all possible locators
                        const locators = await this.elementAnalyzer.generateLocators(handle);

                        // Score locator stability
                        const scoredLocators = this.elementAnalyzer.scoreLocators(locators);

                        const element: RecordedElement = {
                            ...elementData,
                            locators,
                            locatorScores: scoredLocators,
                            recommendedLocator: scoredLocators[0].locator,
                            interactionType: this.inferInteractionType(elementData)
                        };

                        elements.push(element);
                    }
                }
            } catch (error) {
                console.warn(`Failed to capture ${selector}:`, error);
            }
        }

        return elements;
    }

    private async captureInteractions(): Promise<void> {
        if (!this.page || !this.currentPageData) return;

        try {
            // Get interactions from browser
            const interactions = await this.page.evaluate(() => {
                const recorder = (window as any).__testCopilotRecorder;
                const interactions = recorder.interactions;
                recorder.interactions = []; // Clear after reading
                return interactions;
            });

            if (interactions && interactions.length > 0) {
                this.currentPageData.interactions.push(...interactions);
            }
        } catch (error) {
            // Page might have navigated, ignore
        }
    }

    private async handleInteraction(): Promise<void> {
        if (!this.page || !this.currentPageData) return;

        try {
            const lastInteraction = await this.page.evaluate(() => {
                return (window as any).__testCopilotLastInteraction;
            });

            if (lastInteraction) {
                this.currentPageData.interactions.push(lastInteraction);
            }
        } catch (error) {
            // Ignore
        }
    }

    private inferPageName(url: string, title: string): string {
        // Extract page name from URL or title
        const urlPath = new URL(url).pathname;
        const segments = urlPath.split('/').filter(s => s);

        if (segments.length > 0) {
            const lastSegment = segments[segments.length - 1];
            // Convert to PascalCase
            return lastSegment
                .split(/[-_]/)
                .map(word => word.charAt(0).toUpperCase() + word.slice(1))
                .join('') + 'Page';
        }

        // Fallback to title
        if (title) {
            return title.replace(/[^a-zA-Z0-9]/g, '') + 'Page';
        }

        return 'Page' + Date.now();
    }

    private inferInteractionType(elementData: any): string {
        const tag = elementData.tagName;
        const type = elementData.type;

        if (tag === 'input') {
            if (type === 'text' || type === 'email' || type === 'password') {
                return 'input';
            }
            if (type === 'checkbox' || type === 'radio') {
                return 'select';
            }
            if (type === 'submit' || type === 'button') {
                return 'click';
            }
        }

        if (tag === 'button' || tag === 'a') {
            return 'click';
        }

        if (tag === 'select') {
            return 'select';
        }

        if (tag === 'textarea') {
            return 'input';
        }

        return 'click';
    }

    private generateSessionId(): string {
        return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }

    private generatePageId(): string {
        return `page_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }

    private getTotalElements(): number {
        if (!this.session) return 0;
        return this.session.pages.reduce((sum, page) => sum + page.elements.length, 0);
    }

    public getSession(): RecordingSession | undefined {
        return this.session;
    }

    public isCurrentlyRecording(): boolean {
        return this.isRecording;
    }
}
