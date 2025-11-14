import * as playwright from 'playwright';
import { Locator, LocatorSet, ScoredLocator } from './types';

export class ElementAnalyzer {

    /**
     * Generate all possible locators for an element
     */
    public async generateLocators(elementHandle: playwright.ElementHandle): Promise<LocatorSet> {
        const locators: LocatorSet = {};

        try {
            const elementData = await elementHandle.evaluate((el: any) => {
                return {
                    id: el.id,
                    name: el.name,
                    className: el.className,
                    tagName: el.tagName.toLowerCase(),
                    type: el.type,
                    text: el.textContent?.trim(),
                    placeholder: el.placeholder,
                    testId: el.getAttribute('data-testid') || el.getAttribute('data-test'),
                    ariaLabel: el.getAttribute('aria-label')
                };
            });

            // ID locator (best)
            if (elementData.id) {
                locators.id = {
                    type: 'id',
                    value: elementData.id,
                    seleniumSyntax: `By.id("${elementData.id}")`,
                    playwrightSyntax: `page.locator('#${elementData.id}')`
                };
            }

            // Test ID locator (also excellent)
            if (elementData.testId) {
                locators.testId = {
                    type: 'testId',
                    value: elementData.testId,
                    seleniumSyntax: `By.cssSelector("[data-testid='${elementData.testId}']")`,
                    playwrightSyntax: `page.getByTestId('${elementData.testId}')`
                };
            }

            // Name locator (good for form fields)
            if (elementData.name) {
                locators.name = {
                    type: 'name',
                    value: elementData.name,
                    seleniumSyntax: `By.name("${elementData.name}")`,
                    playwrightSyntax: `page.locator('[name="${elementData.name}"]')`
                };
            }

            // CSS locator
            const cssSelector = this.generateCssSelector(elementData);
            if (cssSelector) {
                locators.css = {
                    type: 'css',
                    value: cssSelector,
                    seleniumSyntax: `By.cssSelector("${cssSelector}")`,
                    playwrightSyntax: `page.locator('${cssSelector}')`
                };
            }

            // XPath locator (last resort)
            const xpath = this.generateXPath(elementData);
            if (xpath) {
                locators.xpath = {
                    type: 'xpath',
                    value: xpath,
                    seleniumSyntax: `By.xpath("${xpath}")`,
                    playwrightSyntax: `page.locator('xpath=${xpath}')`
                };
            }

            // Text locator (for links/buttons)
            if (elementData.text && elementData.text.length < 50 &&
                (elementData.tagName === 'a' || elementData.tagName === 'button')) {
                locators.text = {
                    type: 'text',
                    value: elementData.text,
                    seleniumSyntax: `By.linkText("${elementData.text}")`,
                    playwrightSyntax: `page.getByText('${elementData.text}')`
                };
            }

        } catch (error) {
            console.error('Error generating locators:', error);
        }

        return locators;
    }

    /**
     * Score locators based on stability and best practices
     */
    public scoreLocators(locators: LocatorSet): ScoredLocator[] {
        const scored: ScoredLocator[] = [];

        // ID - Most stable (100)
        if (locators.id) {
            scored.push({
                locator: locators.id,
                score: 100,
                reasoning: 'ID is unique and most stable locator'
            });
        }

        // Test ID - Excellent for testing (95)
        if (locators.testId) {
            scored.push({
                locator: locators.testId,
                score: 95,
                reasoning: 'data-testid is designed for testing, very stable'
            });
        }

        // Name - Good for form fields (85)
        if (locators.name) {
            scored.push({
                locator: locators.name,
                score: 85,
                reasoning: 'Name attribute is stable for form elements'
            });
        }

        // Text - Good for buttons/links (80)
        if (locators.text) {
            scored.push({
                locator: locators.text,
                score: 80,
                reasoning: 'Text content is stable for buttons and links'
            });
        }

        // CSS - Moderate stability (60-70)
        if (locators.css) {
            const cssScore = this.scoreCssSelector(locators.css.value);
            scored.push({
                locator: locators.css,
                score: cssScore,
                reasoning: this.getCssScoreReasoning(locators.css.value, cssScore)
            });
        }

        // XPath - Least stable (30-50)
        if (locators.xpath) {
            const xpathScore = this.scoreXPath(locators.xpath.value);
            scored.push({
                locator: locators.xpath,
                score: xpathScore,
                reasoning: 'XPath can be brittle, use as last resort'
            });
        }

        // Sort by score descending
        scored.sort((a, b) => b.score - a.score);

        return scored;
    }

    /**
     * Generate CSS selector
     */
    private generateCssSelector(elementData: any): string {
        // Prefer stable selectors
        if (elementData.id) {
            return `#${elementData.id}`;
        }

        if (elementData.testId) {
            return `[data-testid="${elementData.testId}"]`;
        }

        if (elementData.name) {
            return `${elementData.tagName}[name="${elementData.name}"]`;
        }

        // Class-based (if not too generic)
        if (elementData.className && !this.isGenericClass(elementData.className)) {
            const classes = elementData.className.split(' ').filter((c: string) => c);
            if (classes.length > 0) {
                return `${elementData.tagName}.${classes[0]}`;
            }
        }

        // Type-based for inputs
        if (elementData.tagName === 'input' && elementData.type) {
            return `input[type="${elementData.type}"]`;
        }

        // Fallback to tag name (not recommended)
        return elementData.tagName;
    }

    /**
     * Generate XPath
     */
    private generateXPath(elementData: any): string {
        // Try to use attributes
        if (elementData.id) {
            return `//${elementData.tagName}[@id="${elementData.id}"]`;
        }

        if (elementData.name) {
            return `//${elementData.tagName}[@name="${elementData.name}"]`;
        }

        if (elementData.testId) {
            return `//${elementData.tagName}[@data-testid="${elementData.testId}"]`;
        }

        // Text-based XPath for links/buttons
        if (elementData.text && elementData.text.length < 50) {
            return `//${elementData.tagName}[text()="${elementData.text}"]`;
        }

        // Generic XPath (not ideal)
        return `//${elementData.tagName}`;
    }

    /**
     * Score CSS selector based on stability
     */
    private scoreCssSelector(selector: string): number {
        let score = 50; // Base score

        // ID-based: +20
        if (selector.includes('#')) {
            score += 20;
        }

        // Attribute-based: +15
        if (selector.includes('[') && (
            selector.includes('data-testid') ||
            selector.includes('data-test') ||
            selector.includes('name=')
        )) {
            score += 15;
        }

        // Single class (not chained): +10
        const classCount = (selector.match(/\./g) || []).length;
        if (classCount === 1) {
            score += 10;
        } else if (classCount > 2) {
            score -= 10; // Too specific
        }

        // nth-child penalty: -20
        if (selector.includes(':nth-child') || selector.includes(':nth-of-type')) {
            score -= 20;
        }

        // Deep nesting penalty
        const depth = (selector.match(/>/g) || []).length;
        if (depth > 2) {
            score -= (depth - 2) * 5;
        }

        return Math.max(30, Math.min(100, score));
    }

    /**
     * Score XPath based on stability
     */
    private scoreXPath(xpath: string): number {
        let score = 40; // Base score for XPath

        // Attribute-based: +10
        if (xpath.includes('@id') || xpath.includes('@data-testid') || xpath.includes('@name')) {
            score += 10;
        }

        // Text-based: +5
        if (xpath.includes('text()')) {
            score += 5;
        }

        // Position-based penalty: -15
        if (xpath.includes('[1]') || xpath.includes('[2]')) {
            score -= 15;
        }

        // Complex path penalty
        const depth = (xpath.match(/\//g) || []).length;
        if (depth > 3) {
            score -= (depth - 3) * 5;
        }

        return Math.max(20, Math.min(60, score));
    }

    /**
     * Get reasoning for CSS score
     */
    private getCssScoreReasoning(selector: string, score: number): string {
        if (score >= 80) {
            return 'Stable CSS selector using attributes';
        } else if (score >= 60) {
            return 'Moderately stable CSS selector';
        } else if (score >= 40) {
            return 'CSS selector may be brittle, consider improving';
        } else {
            return 'Brittle CSS selector, should be improved';
        }
    }

    /**
     * Check if class name is too generic
     */
    private isGenericClass(className: string): boolean {
        const genericClasses = [
            'container', 'wrapper', 'content', 'main', 'item',
            'row', 'col', 'section', 'div', 'box',
            'active', 'hidden', 'visible', 'disabled'
        ];

        const classes = className.split(' ');
        return classes.some(c => genericClasses.includes(c.toLowerCase()));
    }

    /**
     * Group related elements (forms, lists, etc.)
     */
    public groupElements(elements: any[]): any[] {
        const groups: any[] = [];

        // Detect forms
        const formElements = elements.filter(el =>
            el.tagName === 'input' ||
            el.tagName === 'select' ||
            el.tagName === 'textarea'
        );

        if (formElements.length > 0) {
            groups.push({
                type: 'form',
                elements: formElements,
                purpose: 'Form input group'
            });
        }

        // Detect navigation
        const navElements = elements.filter(el =>
            el.tagName === 'a' &&
            el.text &&
            !el.text.includes('@') // Not email links
        );

        if (navElements.length > 2) {
            groups.push({
                type: 'navigation',
                elements: navElements,
                purpose: 'Navigation links'
            });
        }

        // Detect buttons
        const buttonElements = elements.filter(el =>
            el.tagName === 'button' ||
            (el.tagName === 'input' && (el.type === 'submit' || el.type === 'button'))
        );

        if (buttonElements.length > 0) {
            groups.push({
                type: 'actions',
                elements: buttonElements,
                purpose: 'Action buttons'
            });
        }

        return groups;
    }

    /**
     * Infer element purpose/name
     */
    public inferElementName(element: any): string {
        // Use ID if available
        if (element.id) {
            return this.toCamelCase(element.id);
        }

        // Use name attribute
        if (element.name) {
            return this.toCamelCase(element.name);
        }

        // Use test ID
        if (element.attributes['data-testid']) {
            return this.toCamelCase(element.attributes['data-testid']);
        }

        // Use placeholder
        if (element.placeholder) {
            return this.toCamelCase(element.placeholder);
        }

        // Use text content for buttons/links
        if (element.text && element.text.length < 30) {
            return this.toCamelCase(element.text);
        }

        // Fallback to type
        return `${element.tagName}${Math.random().toString(36).substr(2, 4)}`;
    }

    /**
     * Convert string to camelCase
     */
    private toCamelCase(str: string): string {
        return str
            .replace(/[^a-zA-Z0-9]+(.)/g, (_, char) => char.toUpperCase())
            .replace(/^[A-Z]/, char => char.toLowerCase())
            .replace(/[^a-zA-Z0-9]/g, '');
    }
}
