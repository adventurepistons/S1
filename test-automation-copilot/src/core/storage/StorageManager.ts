import * as fs from 'fs';
import * as path from 'path';
import Database from 'better-sqlite3';
import { FrameworkAnalysis } from '../analyzers/WorkspaceAnalyzer';
import { ParsedClass } from '../parsers/JavaParser';

export interface ChatMessage {
    id: string;
    role: 'user' | 'assistant';
    content: string;
    timestamp: number;
    context?: any;
}

export interface GeneratedCode {
    id: string;
    type: 'pageObject' | 'test' | 'utility';
    fileName: string;
    code: string;
    timestamp: number;
    applied: boolean;
}

export class StorageManager {
    private db: Database.Database;
    private storagePath: string;
    private vectorStore: Map<string, number[]>; // Simple in-memory vector store for now

    constructor(storagePath: string) {
        this.storagePath = storagePath;

        // Ensure storage directory exists
        if (!fs.existsSync(storagePath)) {
            fs.mkdirSync(storagePath, { recursive: true });
        }

        // Initialize SQLite database
        const dbPath = path.join(storagePath, 'copilot.db');
        this.db = new Database(dbPath);

        // Initialize vector store (in-memory for MVP, can integrate ChromaDB later)
        this.vectorStore = new Map();
    }

    public async initialize(): Promise<void> {
        // Create tables
        this.db.exec(`
            CREATE TABLE IF NOT EXISTS framework_analysis (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                workspace_path TEXT NOT NULL,
                framework TEXT,
                test_runner TEXT,
                analysis_data TEXT,
                created_at INTEGER,
                updated_at INTEGER
            );

            CREATE TABLE IF NOT EXISTS parsed_classes (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                workspace_path TEXT,
                class_name TEXT,
                file_path TEXT,
                class_type TEXT,
                class_data TEXT,
                created_at INTEGER
            );

            CREATE TABLE IF NOT EXISTS chat_history (
                id TEXT PRIMARY KEY,
                role TEXT NOT NULL,
                content TEXT NOT NULL,
                context TEXT,
                timestamp INTEGER NOT NULL
            );

            CREATE TABLE IF NOT EXISTS generated_code (
                id TEXT PRIMARY KEY,
                type TEXT NOT NULL,
                file_name TEXT NOT NULL,
                code TEXT NOT NULL,
                timestamp INTEGER NOT NULL,
                applied INTEGER DEFAULT 0
            );

            CREATE TABLE IF NOT EXISTS settings (
                key TEXT PRIMARY KEY,
                value TEXT
            );

            CREATE INDEX IF NOT EXISTS idx_parsed_classes_workspace ON parsed_classes(workspace_path);
            CREATE INDEX IF NOT EXISTS idx_parsed_classes_type ON parsed_classes(class_type);
            CREATE INDEX IF NOT EXISTS idx_chat_history_timestamp ON chat_history(timestamp);
        `);

        console.log('Storage manager initialized');
    }

    // ===== Framework Analysis Storage =====

    public async storeAnalysis(analysis: FrameworkAnalysis): Promise<void> {
        const stmt = this.db.prepare(`
            INSERT OR REPLACE INTO framework_analysis (workspace_path, framework, test_runner, analysis_data, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?)
        `);

        stmt.run(
            analysis.structure.rootPath,
            analysis.framework,
            analysis.testRunner,
            JSON.stringify(analysis),
            Date.now(),
            Date.now()
        );

        // Store individual parsed classes for semantic search
        for (const po of analysis.pageObjects) {
            await this.storeParsedClass(analysis.structure.rootPath, po.className, po.filePath, 'pageObject', po);
        }

        for (const tc of analysis.testCases) {
            await this.storeParsedClass(analysis.structure.rootPath, tc.className, tc.filePath, 'test', tc);
        }

        console.log(`Stored analysis for ${analysis.structure.rootPath}`);
    }

    public getAnalysis(workspacePath: string): FrameworkAnalysis | null {
        const stmt = this.db.prepare(`
            SELECT analysis_data FROM framework_analysis
            WHERE workspace_path = ?
            ORDER BY updated_at DESC
            LIMIT 1
        `);

        const row = stmt.get(workspacePath) as any;
        return row ? JSON.parse(row.analysis_data) : null;
    }

    private async storeParsedClass(
        workspacePath: string,
        className: string,
        filePath: string,
        type: string,
        data: any
    ): Promise<void> {
        const stmt = this.db.prepare(`
            INSERT OR REPLACE INTO parsed_classes (workspace_path, class_name, file_path, class_type, class_data, created_at)
            VALUES (?, ?, ?, ?, ?, ?)
        `);

        stmt.run(
            workspacePath,
            className,
            filePath,
            type,
            JSON.stringify(data),
            Date.now()
        );

        // Create embedding for semantic search (simplified version)
        const embedding = this.createSimpleEmbedding(className, type, data);
        this.vectorStore.set(`${workspacePath}:${className}`, embedding);
    }

    // ===== Chat History =====

    public saveChatMessage(message: ChatMessage): void {
        const stmt = this.db.prepare(`
            INSERT INTO chat_history (id, role, content, context, timestamp)
            VALUES (?, ?, ?, ?, ?)
        `);

        stmt.run(
            message.id,
            message.role,
            message.content,
            message.context ? JSON.stringify(message.context) : null,
            message.timestamp
        );
    }

    public getChatHistory(limit: number = 50): ChatMessage[] {
        const stmt = this.db.prepare(`
            SELECT * FROM chat_history
            ORDER BY timestamp DESC
            LIMIT ?
        `);

        const rows = stmt.all(limit) as any[];
        return rows.reverse().map(row => ({
            id: row.id,
            role: row.role,
            content: row.content,
            context: row.context ? JSON.parse(row.context) : undefined,
            timestamp: row.timestamp
        }));
    }

    public clearChatHistory(): void {
        this.db.prepare('DELETE FROM chat_history').run();
    }

    // ===== Generated Code =====

    public saveGeneratedCode(code: GeneratedCode): void {
        const stmt = this.db.prepare(`
            INSERT INTO generated_code (id, type, file_name, code, timestamp, applied)
            VALUES (?, ?, ?, ?, ?, ?)
        `);

        stmt.run(
            code.id,
            code.type,
            code.fileName,
            code.code,
            code.timestamp,
            code.applied ? 1 : 0
        );
    }

    public markCodeAsApplied(id: string): void {
        this.db.prepare('UPDATE generated_code SET applied = 1 WHERE id = ?').run(id);
    }

    public getGeneratedCode(limit: number = 20): GeneratedCode[] {
        const stmt = this.db.prepare(`
            SELECT * FROM generated_code
            ORDER BY timestamp DESC
            LIMIT ?
        `);

        const rows = stmt.all(limit) as any[];
        return rows.map(row => ({
            id: row.id,
            type: row.type,
            fileName: row.file_name,
            code: row.code,
            timestamp: row.timestamp,
            applied: row.applied === 1
        }));
    }

    // ===== Semantic Search =====

    public async semanticSearch(query: string, workspacePath: string, topK: number = 5): Promise<any[]> {
        // Create query embedding
        const queryEmbedding = this.createSimpleEmbedding(query, 'query', { query });

        // Find similar items (cosine similarity)
        const results: { key: string, similarity: number }[] = [];

        for (const [key, embedding] of this.vectorStore.entries()) {
            if (key.startsWith(workspacePath)) {
                const similarity = this.cosineSimilarity(queryEmbedding, embedding);
                results.push({ key, similarity });
            }
        }

        // Sort by similarity and take top K
        results.sort((a, b) => b.similarity - a.similarity);
        const topResults = results.slice(0, topK);

        // Fetch full data from database
        const relevantClasses: any[] = [];
        for (const result of topResults) {
            const className = result.key.split(':')[1];
            const stmt = this.db.prepare(`
                SELECT class_data FROM parsed_classes
                WHERE workspace_path = ? AND class_name = ?
            `);
            const row = stmt.get(workspacePath, className) as any;
            if (row) {
                relevantClasses.push({
                    ...JSON.parse(row.class_data),
                    similarity: result.similarity
                });
            }
        }

        return relevantClasses;
    }

    // ===== Settings =====

    public setSetting(key: string, value: string): void {
        this.db.prepare('INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)').run(key, value);
    }

    public getSetting(key: string): string | null {
        const row = this.db.prepare('SELECT value FROM settings WHERE key = ?').get(key) as any;
        return row ? row.value : null;
    }

    // ===== Helper Methods =====

    private createSimpleEmbedding(text: string, type: string, data: any): number[] {
        // Simplified embedding: TF-IDF style feature vector
        // In production, use OpenAI embeddings API or a local model

        const features: number[] = [];
        const words = text.toLowerCase().split(/\W+/);

        // Feature 1-10: Common keywords
        const keywords = ['login', 'page', 'test', 'click', 'enter', 'verify', 'assert', 'driver', 'element', 'wait'];
        for (const keyword of keywords) {
            features.push(words.filter(w => w.includes(keyword)).length);
        }

        // Feature 11-15: Type indicators
        features.push(type === 'pageObject' ? 1 : 0);
        features.push(type === 'test' ? 1 : 0);
        features.push(type === 'step' ? 1 : 0);
        features.push(type === 'utility' ? 1 : 0);
        features.push(type === 'query' ? 1 : 0);

        // Feature 16-20: Data characteristics
        if (data.elements) features.push(data.elements.length);
        else features.push(0);

        if (data.methods) features.push(data.methods.length);
        else features.push(0);

        if (data.testMethods) features.push(data.testMethods.length);
        else features.push(0);

        features.push(JSON.stringify(data).length / 100); // Size indicator
        features.push(words.length); // Text length

        return features;
    }

    private cosineSimilarity(a: number[], b: number[]): number {
        if (a.length !== b.length) return 0;

        let dotProduct = 0;
        let normA = 0;
        let normB = 0;

        for (let i = 0; i < a.length; i++) {
            dotProduct += a[i] * b[i];
            normA += a[i] * a[i];
            normB += b[i] * b[i];
        }

        if (normA === 0 || normB === 0) return 0;

        return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB));
    }

    public close(): void {
        this.db.close();
    }
}
