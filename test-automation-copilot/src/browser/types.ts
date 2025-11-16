// Recording Session Types

export interface RecordingSession {
    id: string;
    startTime: number;
    endTime?: number;
    duration?: number;
    pages: RecordedPage[];
    userFlows: UserFlow[];
    applicationMap: ApplicationMap;
}

export interface RecordedPage {
    id: string;
    url: string;
    title: string;
    timestamp: number;
    elements: RecordedElement[];
    interactions: RecordedInteraction[];
    screenshot?: Buffer;
    domSnapshot?: string;
}

export interface RecordedElement {
    tagName: string;
    id: string | null;
    className: string | null;
    name: string | null;
    type: string | null;
    text: string | null;
    placeholder: string | null;
    value: string | null;
    href: string | null;
    visible: boolean;
    position: ElementPosition;
    attributes: Record<string, string>;
    path: string;
    locators: LocatorSet;
    locatorScores: ScoredLocator[];
    recommendedLocator: Locator;
    interactionType: 'click' | 'input' | 'select' | 'hover' | 'navigate';
}

export interface ElementPosition {
    x: number;
    y: number;
    width: number;
    height: number;
}

export interface LocatorSet {
    id?: Locator;
    css?: Locator;
    xpath?: Locator;
    testId?: Locator;
    text?: Locator;
    name?: Locator;
}

export interface Locator {
    type: 'id' | 'css' | 'xpath' | 'testId' | 'text' | 'name';
    value: string;
    seleniumSyntax: string;
    playwrightSyntax: string;
}

export interface ScoredLocator {
    locator: Locator;
    score: number;
    reasoning: string;
}

export interface RecordedInteraction {
    type: string;
    timestamp: number;
    element: any;
    url: string;
    value?: string;
}

export interface UserFlow {
    name: string;
    pages: string[]; // Page IDs
    actions: string[];
}

export interface ApplicationMap {
    nodes: MapNode[];
    edges: MapEdge[];
}

export interface MapNode {
    id: string;
    label: string;
    url: string;
    elementCount: number;
}

export interface MapEdge {
    from: number;
    to: number;
    label: string;
}

// Code Generation Types

export interface PageObjectSpec {
    className: string;
    packageName: string;
    elements: ElementSpec[];
    methods: MethodSpec[];
    imports: string[];
}

export interface ElementSpec {
    name: string;
    locator: Locator;
    type: string;
    comment?: string;
}

export interface MethodSpec {
    name: string;
    returnType: string;
    parameters: ParameterSpec[];
    body: string;
    comment?: string;
}

export interface ParameterSpec {
    name: string;
    type: string;
}

// Analysis Types

export interface ElementGroup {
    type: 'form' | 'navigation' | 'list' | 'card' | 'modal';
    elements: RecordedElement[];
    purpose: string;
}

export interface DetectedPattern {
    type: string;
    confidence: number;
    description: string;
    elements: string[]; // Element IDs
}
