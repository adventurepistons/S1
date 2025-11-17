import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';

export interface FrameworkInfo {
    id: string;
    name: string;
    language: string;
    requiresPro: boolean;
    icon: string;
}

const FRAMEWORKS: { [key: string]: FrameworkInfo } = {
    'selenium-java': {
        id: 'selenium-java',
        name: 'Selenium WebDriver (Java)',
        language: 'java',
        requiresPro: false,
        icon: '☕'
    },
    'playwright-typescript': {
        id: 'playwright-typescript',
        name: 'Playwright (TypeScript)',
        language: 'typescript',
        requiresPro: true,
        icon: '🎭'
    },
    'playwright-javascript': {
        id: 'playwright-javascript',
        name: 'Playwright (JavaScript)',
        language: 'javascript',
        requiresPro: true,
        icon: '🎭'
    },
    'cypress-javascript': {
        id: 'cypress-javascript',
        name: 'Cypress (JavaScript)',
        language: 'javascript',
        requiresPro: true,
        icon: '🌲'
    },
    'cypress-typescript': {
        id: 'cypress-typescript',
        name: 'Cypress (TypeScript)',
        language: 'typescript',
        requiresPro: true,
        icon: '🌲'
    },
    'pytest-python': {
        id: 'pytest-python',
        name: 'Pytest + Selenium (Python)',
        language: 'python',
        requiresPro: true,
        icon: '🐍'
    },
    'webdriverio-javascript': {
        id: 'webdriverio-javascript',
        name: 'WebdriverIO (JavaScript)',
        language: 'javascript',
        requiresPro: true,
        icon: '🤖'
    }
};

export class FrameworkDetector {

    // Detect framework from workspace
    async detectFramework(workspace: vscode.WorkspaceFolder): Promise<FrameworkInfo | null> {
        const rootPath = workspace.uri.fsPath;

        // Check for Node.js projects (package.json)
        const packageJsonPath = path.join(rootPath, 'package.json');
        if (fs.existsSync(packageJsonPath)) {
            const framework = await this.detectFromPackageJson(packageJsonPath);
            if (framework) return framework;
        }

        // Check for Java projects (pom.xml or build.gradle)
        const pomPath = path.join(rootPath, 'pom.xml');
        const gradlePath = path.join(rootPath, 'build.gradle');
        if (fs.existsSync(pomPath) || fs.existsSync(gradlePath)) {
            return FRAMEWORKS['selenium-java'];
        }

        // Check for Python projects (requirements.txt or pyproject.toml)
        const requirementsPath = path.join(rootPath, 'requirements.txt');
        const pyprojectPath = path.join(rootPath, 'pyproject.toml');
        if (fs.existsSync(requirementsPath)) {
            const framework = await this.detectFromRequirements(requirementsPath);
            if (framework) return framework;
        }
        if (fs.existsSync(pyprojectPath)) {
            return FRAMEWORKS['pytest-python'];
        }

        // Default: check active file language
        const activeEditor = vscode.window.activeTextEditor;
        if (activeEditor) {
            const language = activeEditor.document.languageId;
            if (language === 'java') {
                return FRAMEWORKS['selenium-java'];
            }
        }

        return null;
    }

    // Detect from package.json
    private async detectFromPackageJson(packageJsonPath: string): Promise<FrameworkInfo | null> {
        try {
            const content = fs.readFileSync(packageJsonPath, 'utf-8');
            const packageJson = JSON.parse(content);
            const deps = { ...packageJson.dependencies, ...packageJson.devDependencies };

            // Check for Playwright
            if (deps['@playwright/test']) {
                if (this.hasTypeScript(packageJsonPath)) {
                    return FRAMEWORKS['playwright-typescript'];
                }
                return FRAMEWORKS['playwright-javascript'];
            }

            // Check for Cypress
            if (deps['cypress']) {
                if (this.hasTypeScript(packageJsonPath)) {
                    return FRAMEWORKS['cypress-typescript'];
                }
                return FRAMEWORKS['cypress-javascript'];
            }

            // Check for WebdriverIO
            if (deps['webdriverio'] || deps['@wdio/cli']) {
                return FRAMEWORKS['webdriverio-javascript'];
            }

            return null;
        } catch (error) {
            return null;
        }
    }

    // Detect from requirements.txt
    private async detectFromRequirements(requirementsPath: string): Promise<FrameworkInfo | null> {
        try {
            const content = fs.readFileSync(requirementsPath, 'utf-8');
            if (content.includes('pytest') && content.includes('selenium')) {
                return FRAMEWORKS['pytest-python'];
            }
            return null;
        } catch (error) {
            return null;
        }
    }

    // Check if project uses TypeScript
    private hasTypeScript(packageJsonPath: string): boolean {
        const dir = path.dirname(packageJsonPath);
        const tsconfigPath = path.join(dir, 'tsconfig.json');
        return fs.existsSync(tsconfigPath);
    }

    // Get all frameworks
    getAllFrameworks(): FrameworkInfo[] {
        return Object.values(FRAMEWORKS);
    }

    // Get framework by ID
    getFramework(id: string): FrameworkInfo | null {
        return FRAMEWORKS[id] || null;
    }

    // Get free frameworks
    getFreeFrameworks(): FrameworkInfo[] {
        return Object.values(FRAMEWORKS).filter(f => !f.requiresPro);
    }

    // Get pro frameworks
    getProFrameworks(): FrameworkInfo[] {
        return Object.values(FRAMEWORKS).filter(f => f.requiresPro);
    }
}
