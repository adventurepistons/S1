import * as fs from 'fs';
import * as path from 'path';
import { JavaParser, ParsedClass } from '../parsers/JavaParser';
import { StorageManager } from '../storage/StorageManager';

export interface FrameworkAnalysis {
    framework: 'selenium-java' | 'selenium-python' | 'playwright-ts' | 'unknown';
    testRunner: 'testng' | 'junit' | 'cucumber' | 'unknown';
    pageObjects: PageObjectInfo[];
    testCases: TestCaseInfo[];
    stepDefinitions: StepDefinitionInfo[];
    utilities: UtilityInfo[];
    dependencies: string[];
    structure: ProjectStructure;
}

export interface PageObjectInfo {
    className: string;
    filePath: string;
    elements: ElementInfo[];
    methods: ActionInfo[];
}

export interface ElementInfo {
    name: string;
    locatorType: string;
    locatorValue: string;
}

export interface ActionInfo {
    name: string;
    description: string;
    parameters: string[];
}

export interface TestCaseInfo {
    className: string;
    filePath: string;
    testMethods: TestMethodInfo[];
    usesPageObjects: string[];
}

export interface TestMethodInfo {
    name: string;
    annotations: string[];
    description: string;
}

export interface StepDefinitionInfo {
    className: string;
    filePath: string;
    steps: StepInfo[];
}

export interface StepInfo {
    type: 'Given' | 'When' | 'Then' | 'And';
    pattern: string;
    method: string;
}

export interface UtilityInfo {
    className: string;
    filePath: string;
    purpose: string;
}

export interface ProjectStructure {
    rootPath: string;
    pagesDir?: string;
    testsDir?: string;
    stepsDir?: string;
    utilsDir?: string;
    resourcesDir?: string;
}

export class WorkspaceAnalyzer {
    private javaParser: JavaParser;
    private storageManager: StorageManager;

    constructor(storageManager: StorageManager) {
        this.javaParser = new JavaParser();
        this.storageManager = storageManager;
    }

    public async analyzeWorkspace(workspacePath: string): Promise<FrameworkAnalysis> {
        console.log(`Analyzing workspace: ${workspacePath}`);

        const analysis: FrameworkAnalysis = {
            framework: 'unknown',
            testRunner: 'unknown',
            pageObjects: [],
            testCases: [],
            stepDefinitions: [],
            utilities: [],
            dependencies: [],
            structure: {
                rootPath: workspacePath
            }
        };

        // Detect project structure
        analysis.structure = this.detectProjectStructure(workspacePath);

        // Parse dependencies (pom.xml or build.gradle)
        analysis.dependencies = this.parseDependencies(workspacePath);

        // Detect framework and test runner
        analysis.framework = this.detectFramework(analysis.dependencies, workspacePath);
        analysis.testRunner = this.detectTestRunner(analysis.dependencies, workspacePath);

        // Parse all Java files
        const srcPath = this.findSrcDirectory(workspacePath);
        if (srcPath) {
            const parsedClasses = this.javaParser.parseDirectory(srcPath);

            // Categorize classes
            for (const parsedClass of parsedClasses) {
                switch (parsedClass.type) {
                    case 'pageObject':
                        analysis.pageObjects.push(this.convertToPageObjectInfo(parsedClass));
                        break;
                    case 'test':
                        analysis.testCases.push(this.convertToTestCaseInfo(parsedClass));
                        break;
                    case 'step':
                        analysis.stepDefinitions.push(this.convertToStepDefinitionInfo(parsedClass));
                        break;
                    case 'utility':
                        analysis.utilities.push(this.convertToUtilityInfo(parsedClass));
                        break;
                }
            }
        }

        return analysis;
    }

    private detectProjectStructure(rootPath: string): ProjectStructure {
        const structure: ProjectStructure = { rootPath };

        const possibleDirs = [
            { key: 'pagesDir', paths: ['src/test/java/pages', 'src/main/java/pages', 'pages'] },
            { key: 'testsDir', paths: ['src/test/java/tests', 'src/test/java', 'tests'] },
            { key: 'stepsDir', paths: ['src/test/java/steps', 'src/test/java/stepDefinitions', 'steps'] },
            { key: 'utilsDir', paths: ['src/test/java/utils', 'src/main/java/utils', 'utils'] },
            { key: 'resourcesDir', paths: ['src/test/resources', 'resources'] }
        ];

        for (const { key, paths } of possibleDirs) {
            for (const p of paths) {
                const fullPath = path.join(rootPath, p);
                if (fs.existsSync(fullPath)) {
                    (structure as any)[key] = fullPath;
                    break;
                }
            }
        }

        return structure;
    }

    private parseDependencies(workspacePath: string): string[] {
        const dependencies: string[] = [];

        // Check for pom.xml (Maven)
        const pomPath = path.join(workspacePath, 'pom.xml');
        if (fs.existsSync(pomPath)) {
            const pomContent = fs.readFileSync(pomPath, 'utf8');

            // Extract dependency names
            const depMatches = pomContent.matchAll(/<artifactId>(.*?)<\/artifactId>/g);
            for (const match of depMatches) {
                dependencies.push(match[1]);
            }
        }

        // Check for build.gradle (Gradle)
        const gradlePath = path.join(workspacePath, 'build.gradle');
        if (fs.existsSync(gradlePath)) {
            const gradleContent = fs.readFileSync(gradlePath, 'utf8');

            const depMatches = gradleContent.matchAll(/implementation.*['"]([^'"]+)['"]/g);
            for (const match of depMatches) {
                dependencies.push(match[1]);
            }
        }

        return dependencies;
    }

    private detectFramework(dependencies: string[], workspacePath: string): FrameworkAnalysis['framework'] {
        if (dependencies.some(d => d.includes('selenium'))) {
            // Check language
            if (fs.existsSync(path.join(workspacePath, 'pom.xml'))) {
                return 'selenium-java';
            }
            if (fs.existsSync(path.join(workspacePath, 'requirements.txt'))) {
                return 'selenium-python';
            }
            return 'selenium-java'; // default
        }

        if (dependencies.some(d => d.includes('playwright'))) {
            if (fs.existsSync(path.join(workspacePath, 'package.json'))) {
                return 'playwright-ts';
            }
        }

        return 'unknown';
    }

    private detectTestRunner(dependencies: string[], workspacePath: string): FrameworkAnalysis['testRunner'] {
        if (dependencies.some(d => d.includes('testng'))) {
            return 'testng';
        }

        if (dependencies.some(d => d.includes('junit'))) {
            return 'junit';
        }

        if (dependencies.some(d => d.includes('cucumber'))) {
            return 'cucumber';
        }

        // Check for testng.xml
        if (fs.existsSync(path.join(workspacePath, 'testng.xml'))) {
            return 'testng';
        }

        return 'unknown';
    }

    private findSrcDirectory(workspacePath: string): string | null {
        const possiblePaths = [
            path.join(workspacePath, 'src'),
            path.join(workspacePath, 'test'),
            workspacePath
        ];

        for (const p of possiblePaths) {
            if (fs.existsSync(p)) {
                return p;
            }
        }

        return null;
    }

    private convertToPageObjectInfo(parsedClass: ParsedClass): PageObjectInfo {
        const elements: ElementInfo[] = parsedClass.fields
            .filter(f => f.locatorType && f.locatorValue)
            .map(f => ({
                name: f.name,
                locatorType: f.locatorType!,
                locatorValue: f.locatorValue!
            }));

        const methods: ActionInfo[] = parsedClass.methods
            .filter(m => !m.name.startsWith('get') && m.returnType === 'void')
            .map(m => ({
                name: m.name,
                description: this.inferMethodDescription(m.name),
                parameters: m.parameters.map(p => `${p.type} ${p.name}`)
            }));

        return {
            className: parsedClass.className,
            filePath: parsedClass.filePath,
            elements,
            methods
        };
    }

    private convertToTestCaseInfo(parsedClass: ParsedClass): TestCaseInfo {
        const testMethods: TestMethodInfo[] = parsedClass.methods
            .filter(m => m.annotations.some(a => a.includes('@Test')))
            .map(m => ({
                name: m.name,
                annotations: m.annotations,
                description: this.inferMethodDescription(m.name)
            }));

        const usesPageObjects: string[] = [];
        for (const field of parsedClass.fields) {
            if (field.type.includes('Page')) {
                usesPageObjects.push(field.type);
            }
        }

        return {
            className: parsedClass.className,
            filePath: parsedClass.filePath,
            testMethods,
            usesPageObjects
        };
    }

    private convertToStepDefinitionInfo(parsedClass: ParsedClass): StepDefinitionInfo {
        const steps: StepInfo[] = [];

        for (const method of parsedClass.methods) {
            for (const annotation of method.annotations) {
                const match = annotation.match(/@(Given|When|Then|And)\("(.+?)"\)/);
                if (match) {
                    steps.push({
                        type: match[1] as 'Given' | 'When' | 'Then' | 'And',
                        pattern: match[2],
                        method: method.name
                    });
                }
            }
        }

        return {
            className: parsedClass.className,
            filePath: parsedClass.filePath,
            steps
        };
    }

    private convertToUtilityInfo(parsedClass: ParsedClass): UtilityInfo {
        let purpose = 'Utility class';

        if (parsedClass.className.includes('Driver')) {
            purpose = 'WebDriver management';
        } else if (parsedClass.className.includes('Config')) {
            purpose = 'Configuration management';
        } else if (parsedClass.className.includes('Data')) {
            purpose = 'Test data management';
        } else if (parsedClass.className.includes('Wait')) {
            purpose = 'Wait utilities';
        }

        return {
            className: parsedClass.className,
            filePath: parsedClass.filePath,
            purpose
        };
    }

    private inferMethodDescription(methodName: string): string {
        // Convert camelCase to readable description
        const words = methodName.replace(/([A-Z])/g, ' $1').trim();
        return words.charAt(0).toUpperCase() + words.slice(1);
    }
}
