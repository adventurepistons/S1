import { Injectable, Logger, BadRequestException } from '@nestjs/common';
import { DatabaseService } from '../shared/database.service';
import axios, { AxiosRequestConfig } from 'axios';

export interface ApiTestDefinition {
  name: string;
  type: 'REST' | 'GraphQL';
  method?: string; // GET, POST, etc. for REST
  url: string;
  headers?: Record<string, string>;
  body?: any;
  query?: string; // GraphQL query
  variables?: Record<string, any>; // GraphQL variables
  expectedStatus?: number;
  expectedBody?: any;
  expectedHeaders?: Record<string, string>;
  timeout?: number;
}

export interface ApiTestResult {
  name: string;
  passed: boolean;
  duration: number;
  request: {
    method: string;
    url: string;
    headers: Record<string, string>;
    body?: any;
  };
  response: {
    status: number;
    statusText: string;
    headers: Record<string, string>;
    body: any;
    duration: number;
  };
  assertions: {
    statusCode: { expected: number; actual: number; passed: boolean };
    body?: { passed: boolean; errors?: string[] };
    headers?: { passed: boolean; errors?: string[] };
  };
  error?: string;
}

// URL validation for SSRF prevention
const BLOCKED_HOSTS = [
  'localhost', '127.0.0.1', '0.0.0.0', '::1',
  '169.254.169.254',  // AWS/GCP metadata
  'metadata.google.internal',
  /^10\./,
  /^172\.(1[6-9]|2[0-9]|3[01])\./,
  /^192\.168\./,
  /^fd[0-9a-f]{2}:/i  // IPv6 private
];

function validateApiTestUrl(url: string): void {
  let parsed: URL;

  try {
    parsed = new URL(url);
  } catch (e) {
    throw new BadRequestException('Invalid URL format');
  }

  // Block non-HTTP(S) protocols
  if (!['http:', 'https:'].includes(parsed.protocol)) {
    throw new BadRequestException('Only HTTP and HTTPS protocols are allowed');
  }

  // Block private IPs and localhost
  const hostname = parsed.hostname.toLowerCase();
  for (const blocked of BLOCKED_HOSTS) {
    if (typeof blocked === 'string') {
      if (hostname === blocked) {
        throw new BadRequestException(`Access to ${hostname} is forbidden (SSRF protection)`);
      }
    } else if (blocked.test(hostname)) {
      throw new BadRequestException(`Access to private IP ranges is forbidden (SSRF protection)`);
    }
  }

  // Require HTTPS in production
  if (process.env.NODE_ENV === 'production' && parsed.protocol !== 'https:') {
    throw new BadRequestException('Only HTTPS URLs are allowed in production');
  }
}

@Injectable()
export class ApiTestingService {
  private readonly logger = new Logger(ApiTestingService.name);

  constructor(private db: DatabaseService) {}

  /**
   * Execute REST API test
   */
  async executeRestTest(test: ApiTestDefinition): Promise<ApiTestResult> {
    const startTime = Date.now();

    // ✅ VALIDATE URL BEFORE MAKING REQUEST
    validateApiTestUrl(test.url);

    try {
      // Build request config
      const config: AxiosRequestConfig = {
        method: (test.method || 'GET') as any,
        url: test.url,
        headers: test.headers || {},
        timeout: test.timeout || 30000,
      };

      if (test.body) {
        config.data = test.body;
      }

      // Execute request
      const response = await axios(config);
      const duration = Date.now() - startTime;

      // Run assertions
      const assertions = this.runAssertions(test, response);

      return {
        name: test.name,
        passed: Object.values(assertions).every((a: any) => a.passed),
        duration,
        request: {
          method: config.method!,
          url: config.url!,
          headers: config.headers as Record<string, string>,
          body: config.data,
        },
        response: {
          status: response.status,
          statusText: response.statusText,
          headers: response.headers as Record<string, string>,
          body: response.data,
          duration,
        },
        assertions,
      };
    } catch (error: any) {
      const duration = Date.now() - startTime;

      return {
        name: test.name,
        passed: false,
        duration,
        request: {
          method: test.method || 'GET',
          url: test.url,
          headers: test.headers || {},
          body: test.body,
        },
        response: {
          status: error.response?.status || 0,
          statusText: error.response?.statusText || 'Error',
          headers: error.response?.headers || {},
          body: error.response?.data || null,
          duration,
        },
        assertions: {
          statusCode: {
            expected: test.expectedStatus || 200,
            actual: error.response?.status || 0,
            passed: false,
          },
        },
        error: error.message,
      };
    }
  }

  /**
   * Execute GraphQL test
   */
  async executeGraphQLTest(test: ApiTestDefinition): Promise<ApiTestResult> {
    const startTime = Date.now();

    // ✅ VALIDATE URL BEFORE MAKING REQUEST
    validateApiTestUrl(test.url);

    try {
      // Build GraphQL request
      const response = await axios.post(
        test.url,
        {
          query: test.query,
          variables: test.variables || {},
        },
        {
          headers: {
            'Content-Type': 'application/json',
            ...test.headers,
          },
          timeout: test.timeout || 30000,
        },
      );

      const duration = Date.now() - startTime;

      // Check for GraphQL errors
      const hasErrors = response.data.errors && response.data.errors.length > 0;

      const assertions = {
        statusCode: {
          expected: test.expectedStatus || 200,
          actual: response.status,
          passed: response.status === (test.expectedStatus || 200),
        },
        body: {
          passed: !hasErrors,
          errors: hasErrors ? response.data.errors.map((e: any) => e.message) : [],
        },
      };

      return {
        name: test.name,
        passed: !hasErrors && assertions.statusCode.passed,
        duration,
        request: {
          method: 'POST',
          url: test.url,
          headers: { 'Content-Type': 'application/json', ...test.headers },
          body: { query: test.query, variables: test.variables },
        },
        response: {
          status: response.status,
          statusText: response.statusText,
          headers: response.headers as Record<string, string>,
          body: response.data,
          duration,
        },
        assertions,
      };
    } catch (error: any) {
      const duration = Date.now() - startTime;

      return {
        name: test.name,
        passed: false,
        duration,
        request: {
          method: 'POST',
          url: test.url,
          headers: { 'Content-Type': 'application/json', ...test.headers },
          body: { query: test.query, variables: test.variables },
        },
        response: {
          status: error.response?.status || 0,
          statusText: error.response?.statusText || 'Error',
          headers: error.response?.headers || {},
          body: error.response?.data || null,
          duration,
        },
        assertions: {
          statusCode: {
            expected: test.expectedStatus || 200,
            actual: error.response?.status || 0,
            passed: false,
          },
        },
        error: error.message,
      };
    }
  }

  /**
   * Execute multiple API tests
   */
  async executeTests(tests: ApiTestDefinition[]): Promise<ApiTestResult[]> {
    const results: ApiTestResult[] = [];

    for (const test of tests) {
      this.logger.log(`Executing API test: ${test.name}`);

      const result =
        test.type === 'GraphQL'
          ? await this.executeGraphQLTest(test)
          : await this.executeRestTest(test);

      results.push(result);
    }

    return results;
  }

  /**
   * Generate API tests from OpenAPI/Swagger spec
   */
  async generateTestsFromOpenAPI(specUrl: string): Promise<ApiTestDefinition[]> {
    // ✅ VALIDATE URL BEFORE FETCHING
    validateApiTestUrl(specUrl);

    try {
      // Fetch OpenAPI spec
      const response = await axios.get(specUrl);
      const spec = response.data;

      const tests: ApiTestDefinition[] = [];

      // Parse paths
      for (const [path, methods] of Object.entries(spec.paths || {})) {
        for (const [method, operation] of Object.entries(methods as any)) {
          if (['get', 'post', 'put', 'patch', 'delete'].includes(method)) {
            const baseUrl = spec.servers?.[0]?.url || '';
            const fullUrl = `${baseUrl}${path}`;

            tests.push({
              name: (operation as any).summary || `${method.toUpperCase()} ${path}`,
              type: 'REST',
              method: method.toUpperCase(),
              url: fullUrl,
              expectedStatus: 200,
            });
          }
        }
      }

      this.logger.log(`Generated ${tests.length} tests from OpenAPI spec`);

      return tests;
    } catch (error: any) {
      this.logger.error(`Failed to parse OpenAPI spec: ${error.message}`);
      throw new BadRequestException('Invalid OpenAPI specification');
    }
  }

  /**
   * Generate GraphQL tests from schema introspection
   */
  async generateTestsFromGraphQL(url: string): Promise<ApiTestDefinition[]> {
    // ✅ VALIDATE URL BEFORE FETCHING
    validateApiTestUrl(url);

    try {
      // Introspection query
      const introspectionQuery = `
        query IntrospectionQuery {
          __schema {
            queryType { name }
            mutationType { name }
            types {
              name
              kind
              fields {
                name
                type {
                  name
                  kind
                }
              }
            }
          }
        }
      `;

      const response = await axios.post(url, {
        query: introspectionQuery,
      });

      const schema = response.data.data.__schema;
      const tests: ApiTestDefinition[] = [];

      // Generate tests for queries
      const queryType = schema.types.find((t: any) => t.name === schema.queryType.name);
      if (queryType) {
        for (const field of queryType.fields || []) {
          tests.push({
            name: `Query: ${field.name}`,
            type: 'GraphQL',
            url,
            query: `query { ${field.name} }`,
            expectedStatus: 200,
          });
        }
      }

      this.logger.log(`Generated ${tests.length} tests from GraphQL schema`);

      return tests;
    } catch (error: any) {
      this.logger.error(`Failed to introspect GraphQL schema: ${error.message}`);
      throw new BadRequestException('Invalid GraphQL endpoint');
    }
  }

  /**
   * Run assertions on response
   */
  private runAssertions(test: ApiTestDefinition, response: any): any {
    const assertions: any = {
      statusCode: {
        expected: test.expectedStatus || 200,
        actual: response.status,
        passed: response.status === (test.expectedStatus || 200),
      },
    };

    // Body assertions
    if (test.expectedBody) {
      const bodyPassed = this.deepEqual(test.expectedBody, response.data);
      assertions.body = {
        passed: bodyPassed,
        errors: bodyPassed ? [] : ['Response body does not match expected'],
      };
    }

    // Header assertions
    if (test.expectedHeaders) {
      const headerErrors: string[] = [];
      for (const [key, value] of Object.entries(test.expectedHeaders)) {
        if (response.headers[key.toLowerCase()] !== value) {
          headerErrors.push(`Expected header ${key}: ${value}, got: ${response.headers[key.toLowerCase()]}`);
        }
      }
      assertions.headers = {
        passed: headerErrors.length === 0,
        errors: headerErrors,
      };
    }

    return assertions;
  }

  /**
   * Deep equality check
   */
  private deepEqual(obj1: any, obj2: any): boolean {
    return JSON.stringify(obj1) === JSON.stringify(obj2);
  }
}
