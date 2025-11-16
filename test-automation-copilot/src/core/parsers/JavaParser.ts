import * as fs from 'fs';
import * as path from 'path';
import Parser from 'tree-sitter';
import Java from 'tree-sitter-java';

export interface ParsedClass {
    className: string;
    filePath: string;
    package: string;
    imports: string[];
    fields: FieldInfo[];
    methods: MethodInfo[];
    annotations: string[];
    superClass?: string;
    type: 'pageObject' | 'test' | 'step' | 'utility' | 'unknown';
}

export interface FieldInfo {
    name: string;
    type: string;
    locatorType?: string; // By.id, @FindBy, etc.
    locatorValue?: string;
    annotations: string[];
}

export interface MethodInfo {
    name: string;
    returnType: string;
    parameters: ParameterInfo[];
    annotations: string[];
    body: string;
}

export interface ParameterInfo {
    name: string;
    type: string;
}

export class JavaParser {
    private parser: Parser;

    constructor() {
        this.parser = new Parser();
        this.parser.setLanguage(Java);
    }

    /**
     * Parse a single Java file
     */
    public parseFile(filePath: string): ParsedClass | null {
        try {
            const content = fs.readFileSync(filePath, 'utf8');
            const tree = this.parser.parse(content);

            return this.extractClassInfo(tree.rootNode, content, filePath);
        } catch (error) {
            console.error(`Error parsing ${filePath}:`, error);
            return null;
        }
    }

    /**
     * Parse all Java files in a directory
     */
    public parseDirectory(dirPath: string): ParsedClass[] {
        const results: ParsedClass[] = [];

        const walk = (dir: string) => {
            const files = fs.readdirSync(dir);

            for (const file of files) {
                const filePath = path.join(dir, file);
                const stat = fs.statSync(filePath);

                if (stat.isDirectory()) {
                    walk(filePath);
                } else if (file.endsWith('.java')) {
                    const parsed = this.parseFile(filePath);
                    if (parsed) {
                        results.push(parsed);
                    }
                }
            }
        };

        walk(dirPath);
        return results;
    }

    private extractClassInfo(node: Parser.SyntaxNode, content: string, filePath: string): ParsedClass {
        const result: ParsedClass = {
            className: '',
            filePath,
            package: '',
            imports: [],
            fields: [],
            methods: [],
            annotations: [],
            type: 'unknown'
        };

        // Extract package
        const packageNode = this.findNode(node, 'package_declaration');
        if (packageNode) {
            result.package = this.getNodeText(packageNode, content).replace('package ', '').replace(';', '').trim();
        }

        // Extract imports
        const importNodes = this.findAllNodes(node, 'import_declaration');
        result.imports = importNodes.map(n =>
            this.getNodeText(n, content).replace('import ', '').replace(';', '').trim()
        );

        // Extract class declaration
        const classNode = this.findNode(node, 'class_declaration');
        if (classNode) {
            // Class name
            const identifierNode = this.findNode(classNode, 'identifier');
            if (identifierNode) {
                result.className = this.getNodeText(identifierNode, content);
            }

            // Annotations
            result.annotations = this.extractAnnotations(classNode, content);

            // Super class
            const superclassNode = this.findNode(classNode, 'superclass');
            if (superclassNode) {
                const typeNode = this.findNode(superclassNode, 'type_identifier');
                if (typeNode) {
                    result.superClass = this.getNodeText(typeNode, content);
                }
            }

            // Class body
            const classBody = this.findNode(classNode, 'class_body');
            if (classBody) {
                // Extract fields
                result.fields = this.extractFields(classBody, content);

                // Extract methods
                result.methods = this.extractMethods(classBody, content);
            }
        }

        // Determine class type
        result.type = this.determineClassType(result);

        return result;
    }

    private extractFields(classBody: Parser.SyntaxNode, content: string): FieldInfo[] {
        const fields: FieldInfo[] = [];
        const fieldNodes = this.findAllNodes(classBody, 'field_declaration');

        for (const fieldNode of fieldNodes) {
            const typeNode = this.findNode(fieldNode, 'type');
            const declaratorNode = this.findNode(fieldNode, 'variable_declarator');

            if (typeNode && declaratorNode) {
                const identifierNode = this.findNode(declaratorNode, 'identifier');

                if (identifierNode) {
                    const field: FieldInfo = {
                        name: this.getNodeText(identifierNode, content),
                        type: this.getNodeText(typeNode, content),
                        annotations: this.extractAnnotations(fieldNode, content)
                    };

                    // Extract locator information
                    const locatorInfo = this.extractLocatorInfo(fieldNode, content);
                    if (locatorInfo) {
                        field.locatorType = locatorInfo.type;
                        field.locatorValue = locatorInfo.value;
                    }

                    fields.push(field);
                }
            }
        }

        return fields;
    }

    private extractMethods(classBody: Parser.SyntaxNode, content: string): MethodInfo[] {
        const methods: MethodInfo[] = [];
        const methodNodes = this.findAllNodes(classBody, 'method_declaration');

        for (const methodNode of methodNodes) {
            const nameNode = this.findNode(methodNode, 'identifier');
            const typeNode = this.findNode(methodNode, 'type') || this.findNode(methodNode, 'void_type');

            if (nameNode && typeNode) {
                const method: MethodInfo = {
                    name: this.getNodeText(nameNode, content),
                    returnType: this.getNodeText(typeNode, content),
                    parameters: this.extractParameters(methodNode, content),
                    annotations: this.extractAnnotations(methodNode, content),
                    body: this.getMethodBody(methodNode, content)
                };

                methods.push(method);
            }
        }

        return methods;
    }

    private extractParameters(methodNode: Parser.SyntaxNode, content: string): ParameterInfo[] {
        const parameters: ParameterInfo[] = [];
        const paramsNode = this.findNode(methodNode, 'formal_parameters');

        if (paramsNode) {
            const paramNodes = this.findAllNodes(paramsNode, 'formal_parameter');

            for (const paramNode of paramNodes) {
                const typeNode = this.findNode(paramNode, 'type');
                const identifierNode = this.findNode(paramNode, 'identifier');

                if (typeNode && identifierNode) {
                    parameters.push({
                        name: this.getNodeText(identifierNode, content),
                        type: this.getNodeText(typeNode, content)
                    });
                }
            }
        }

        return parameters;
    }

    private extractAnnotations(node: Parser.SyntaxNode, content: string): string[] {
        const annotations: string[] = [];

        for (const child of node.children) {
            if (child.type === 'modifiers') {
                const annotationNodes = this.findAllNodes(child, 'annotation');
                annotations.push(...annotationNodes.map(n => this.getNodeText(n, content)));
            } else if (child.type === 'annotation') {
                annotations.push(this.getNodeText(child, content));
            }
        }

        return annotations;
    }

    private extractLocatorInfo(fieldNode: Parser.SyntaxNode, content: string): { type: string, value: string } | null {
        const fieldText = this.getNodeText(fieldNode, content);

        // Check for @FindBy annotation
        const findByMatch = fieldText.match(/@FindBy\((.*?)\)/);
        if (findByMatch) {
            const params = findByMatch[1];
            const typeMatch = params.match(/(id|name|css|xpath|className|tagName|linkText|partialLinkText)\s*=\s*"([^"]+)"/);
            if (typeMatch) {
                return { type: `@FindBy(${typeMatch[1]})`, value: typeMatch[2] };
            }
        }

        // Check for By.xxx() pattern
        const byMatch = fieldText.match(/By\.(id|name|css|xpath|className|tagName|linkText|partialLinkText)\("([^"]+)"\)/);
        if (byMatch) {
            return { type: `By.${byMatch[1]}`, value: byMatch[2] };
        }

        return null;
    }

    private getMethodBody(methodNode: Parser.SyntaxNode, content: string): string {
        const bodyNode = this.findNode(methodNode, 'block');
        return bodyNode ? this.getNodeText(bodyNode, content) : '';
    }

    private determineClassType(classInfo: ParsedClass): 'pageObject' | 'test' | 'step' | 'utility' | 'unknown' {
        const className = classInfo.className.toLowerCase();
        const filePath = classInfo.filePath.toLowerCase();

        // Check for Page Object
        if (className.includes('page') || filePath.includes('/pages/')) {
            return 'pageObject';
        }

        // Check for Test class
        if (className.includes('test') || filePath.includes('/tests/') || filePath.includes('/test/')) {
            return 'test';
        }

        // Check for Step definition (Cucumber)
        if (className.includes('step') ||
            filePath.includes('/steps/') ||
            classInfo.annotations.some(a => a.includes('@Given') || a.includes('@When') || a.includes('@Then'))) {
            return 'step';
        }

        // Check for utility class
        if (className.includes('util') ||
            className.includes('helper') ||
            className.includes('factory') ||
            filePath.includes('/utils/') ||
            filePath.includes('/utilities/')) {
            return 'utility';
        }

        return 'unknown';
    }

    private findNode(node: Parser.SyntaxNode, type: string): Parser.SyntaxNode | null {
        if (node.type === type) {
            return node;
        }

        for (const child of node.children) {
            const found = this.findNode(child, type);
            if (found) {
                return found;
            }
        }

        return null;
    }

    private findAllNodes(node: Parser.SyntaxNode, type: string): Parser.SyntaxNode[] {
        const results: Parser.SyntaxNode[] = [];

        if (node.type === type) {
            results.push(node);
        }

        for (const child of node.children) {
            results.push(...this.findAllNodes(child, type));
        }

        return results;
    }

    private getNodeText(node: Parser.SyntaxNode, content: string): string {
        return content.substring(node.startIndex, node.endIndex);
    }
}
