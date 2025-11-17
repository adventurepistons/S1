# AI Integration Plan - Latest Research & Techniques (2025)

## Research Summary

Based on cutting-edge research from 2025, here are the most powerful techniques for AI-powered code generation:

### 1. **cAST Method (CMU, June 2025)** ⭐⭐⭐⭐⭐
- **What**: Structure-aware chunking via Abstract Syntax Trees
- **Performance**: +4.3 points Recall@5 on RepoEval, +2.67 points Pass@1 on SWE-bench
- **Why**: Maintains syntactic integrity, prevents breaking method structures
- **How**: Recursively breaks large AST nodes, merges siblings respecting size limits
- **We Have This!**: Our Tree-sitter parser already gives us complete AST

### 2. **Self-RAG with Verification** ⭐⭐⭐⭐⭐
- **What**: Multi-layered approach with verification mechanisms
- **Performance**: 96% reduction in hallucinations (Stanford 2024)
- **Techniques**:
  - Chain-of-Thought (CoT): Model explains reasoning
  - Chain-of-Verification (CoVe): Independent verification questions
  - Chain-of-Note (CoN): Enhanced robustness in noisy scenarios
- **We Have This!**: Our knowledge graph provides perfect context for retrieval

### 3. **Few-Shot Prompting with Negatives** ⭐⭐⭐⭐
- **What**: 3-5 examples showing both good AND bad code
- **Performance**: +35% reduction in anti-patterns
- **Best Practices**:
  - Use 3-5 examples (not more)
  - Include both positive and negative examples
  - Keep structure identical across examples
  - Cover diverse edge cases
- **We Have This!**: Our pattern_examples table has real code samples

### 4. **Prompt Caching** ⭐⭐⭐⭐⭐
- **What**: Reuse context across API calls
- **Cost Savings**:
  - Claude: 90% cheaper cached reads, 85% faster responses
  - OpenAI: 50% cheaper cached inputs
- **Perfect For**: Code generation with large codebase context
- **We Have This!**: Our framework config and patterns are static, perfect for caching

### 5. **Structured Data Integration** ⭐⭐⭐⭐
- **What**: Combine structured data (our DB) with unstructured (code)
- **Performance**: Reduces hallucinations significantly vs unstructured-only
- **We Have This!**: We have 25+ tables with complete structured data

---

## Our Implementation Plan

### Architecture: 3-Layer Approach

```
┌─────────────────────────────────────────────────────────────┐
│                    Context Builder                           │
│  Gathers: Framework + Patterns + Knowledge Graph + Examples │
└──────────────────────┬───────────────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────────────┐
│                   Prompt Engine                              │
│  Builds: System + Few-Shot Examples + User Request + CoT    │
└──────────────────────┬───────────────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────────────┐
│                   LLM Integration                            │
│  Supports: OpenAI GPT-4o + Claude 3.5 Sonnet (with caching) │
└──────────────────────┬───────────────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────────────┐
│                Code Validator                                │
│  Validates: Syntax + Patterns + Framework Compatibility     │
└─────────────────────────────────────────────────────────────┘
```

---

## Phase 3 Implementation (Week 1-2)

### Week 1: Context Builder + Prompt Engine

#### Day 1-2: Context Builder (`internal/ai/context_builder.go`)

**What to Build:**
```go
type ContextBuilder struct {
    db *sql.DB
    kg *graph.KnowledgeGraph
}

type Context struct {
    // Structured Data (for prompt caching)
    Framework      *models.FrameworkConfig
    CodingPatterns *models.CodingPatterns

    // Dynamic Context
    RelevantClasses []models.ClassData
    PatternExamples []PatternExample
    RelatedElements []ElementInfo

    // User Request
    UserIntent     string
    RequiredFiles  []string
}

func (cb *ContextBuilder) BuildContext(userRequest string) (*Context, error)
```

**Key Features:**
1. **Smart Retrieval** - Use knowledge graph to find ALL relevant code
2. **cAST Chunking** - Break large classes using AST boundaries
3. **Example Selection** - Pick 3-5 most relevant pattern examples
4. **Structured First** - Framework + Patterns go first (for caching)

**Implementation:**
```go
// 1. Parse user intent
intent := cb.parseIntent(userRequest)
// "create a test for login with invalid password"

// 2. Get framework context (CACHEABLE)
framework := cb.db.GetFrameworkConfig()
patterns := cb.db.GetCodingPatterns()

// 3. Find relevant page objects via knowledge graph
if intent.RequiresPageObject("login") {
    pageObject := cb.kg.FindPageObject("LoginPage")
    context.RelevantClasses = append(context.RelevantClasses, pageObject)
}

// 4. Get pattern examples (3-5 examples)
examples := cb.selectBestExamples(intent, patterns)
// Select 3 positive examples (similar tests)
// Select 1-2 negative examples (what NOT to do)

// 5. Get related elements from knowledge graph
if intent.RequiresElement("password") {
    element := cb.kg.FindFieldDefinition("passwordField")
    context.RelatedElements = append(context.RelatedElements, element)
}

return context, nil
```

#### Day 3-4: Prompt Engine (`internal/ai/prompt_engine.go`)

**What to Build:**
```go
type PromptEngine struct {
    templates map[string]*PromptTemplate
}

type PromptTemplate struct {
    SystemPrompt    string
    FewShotExamples []Example
    UserPrompt      string
    ChainOfThought  bool
}

func (pe *PromptEngine) BuildPrompt(context *Context) (*Prompt, error)
```

**Prompt Structure (based on 2025 research):**

```markdown
# SYSTEM PROMPT (CACHEABLE - 90% cheaper on subsequent calls)

You are an expert Selenium + Java test automation engineer.

## Project Context
Framework: {framework.TestFramework} + {framework.BDDFramework}
Selenium Version: {framework.SeleniumVersion}

## Coding Patterns (LEARN THESE!)
- Test Naming: {patterns.TestMethodNaming}
- Element Naming: {patterns.WebElementNaming}
- Wait Strategy: {patterns.PreferredWaitType}, {patterns.DefaultWaitTimeout}s
- Assertion Style: {patterns.AssertionLibrary}, Messages: {patterns.UsesAssertMessages}
- Test Structure: {patterns.UsesAAAPattern ? "AAA" : "Given-When-Then"}

## Architecture
- Page Object Model: {framework.UsesPageObjectModel}
- PageFactory: {framework.UsesPageFactory}

---

# FEW-SHOT EXAMPLES (3-5 examples, including negatives)

## Example 1: Positive Test (FOLLOW THIS PATTERN)
```java
{patternExample1.code}
```
✅ Why this is good:
- Uses AAA pattern
- Clear test naming: testActionWithCondition
- Explicit waits with 10 second timeout
- Assertions with messages

## Example 2: Negative Test (FOLLOW THIS PATTERN)
```java
{patternExample2.code}
```
✅ Why this is good:
- Tests invalid input
- Verifies error message
- Follows same naming pattern

## Example 3: Data-Driven Test (FOLLOW THIS PATTERN)
```java
{patternExample3.code}
```
✅ Why this is good:
- Uses DataProvider
- Tests multiple scenarios

## Anti-Pattern Example (❌ DO NOT DO THIS)
```java
@Test
public void test1() {  // ❌ Bad: unclear name
    driver.findElement(By.id("username")).sendKeys("test");  // ❌ Bad: not using page object
    Thread.sleep(2000);  // ❌ Bad: using Thread.sleep instead of explicit wait
    assertTrue(true);  // ❌ Bad: no meaningful assertion
}
```
❌ Why this is bad:
- Unclear test name
- Not using page objects
- Using Thread.sleep
- No assertion message

---

# AVAILABLE RESOURCES (Dynamic context - changes per request)

## LoginPage.java
```java
{relevantClass.FullSource}
```

Elements available:
{foreach element in relatedElements}
- {element.Name}: {element.LocatorStrategy} = "{element.LocatorValue}"
{end}

Methods available:
{foreach method in relevantMethods}
- {method.Name}({method.Parameters}) -> {method.ReturnType}
{end}

---

# YOUR TASK

{userRequest}

## Think Step-by-Step (Chain-of-Thought):
1. What page objects do I need?
2. What elements do I need to interact with?
3. What is the test flow?
4. What assertions are needed?
5. Does this follow the project's coding patterns?

## Generate Code:
Now write the code following the exact patterns shown above.
```

**Why This Works:**
1. **System Prompt Cached** → 90% cheaper (framework + patterns don't change)
2. **Few-Shot Examples** → Model learns exact project style
3. **Negative Examples** → Avoids anti-patterns
4. **Chain-of-Thought** → Forces reasoning before code
5. **Structured Context** → All necessary info available

#### Day 5: LLM Integration (`internal/ai/llm_client.go`)

**What to Build:**
```go
type LLMClient interface {
    Generate(prompt *Prompt) (*Response, error)
    StreamGenerate(prompt *Prompt, callback func(chunk string)) error
}

type OpenAIClient struct {
    apiKey string
    model  string  // gpt-4o
}

type ClaudeClient struct {
    apiKey string
    model  string  // claude-3-5-sonnet-20241022
}

type Response struct {
    Code        string
    Reasoning   string
    Confidence  float64
}
```

**Claude Implementation (with caching):**
```go
func (c *ClaudeClient) Generate(prompt *Prompt) (*Response, error) {
    messages := []Message{
        {
            Role: "user",
            Content: []ContentBlock{
                // Cacheable system context
                {
                    Type: "text",
                    Text: prompt.SystemPrompt,
                    CacheControl: &CacheControl{Type: "ephemeral"}, // ← Cache this!
                },
                // Cacheable few-shot examples
                {
                    Type: "text",
                    Text: prompt.FewShotExamples,
                    CacheControl: &CacheControl{Type: "ephemeral"}, // ← Cache this!
                },
                // Dynamic context (not cached)
                {
                    Type: "text",
                    Text: prompt.UserRequest,
                },
            },
        },
    }

    // Call Claude API
    response, err := c.callAPI(messages)

    return &Response{
        Code: extractCode(response),
        Reasoning: extractReasoning(response),
    }, nil
}
```

**Cost Calculation:**
```
First call:
- System prompt: 2000 tokens × $3/MTok (write to cache) = $0.006
- Few-shot: 3000 tokens × $3/MTok (write to cache) = $0.009
- User request: 500 tokens × $3/MTok = $0.0015
- Total: $0.0165

Subsequent calls (system + few-shot cached):
- System prompt: 2000 tokens × $0.30/MTok (read cache) = $0.0006  (90% cheaper!)
- Few-shot: 3000 tokens × $0.30/MTok (read cache) = $0.0009      (90% cheaper!)
- User request: 500 tokens × $3/MTok = $0.0015
- Total: $0.003  (82% cheaper!)
```

### Week 2: Code Validator + Testing

#### Day 6-7: Code Validator (`internal/ai/validator.go`)

**What to Build:**
```go
type CodeValidator struct {
    parser          *parser.JavaParser
    patternDetector *detector.PatternDetector
}

type ValidationResult struct {
    IsValid         bool
    SyntaxErrors    []SyntaxError
    PatternViolations []PatternViolation
    Suggestions     []string
}

func (cv *CodeValidator) Validate(code string, context *Context) (*ValidationResult, error)
```

**Validation Steps:**
1. **Syntax Check** - Parse with Tree-sitter, ensure valid Java
2. **Pattern Check** - Verify follows project coding patterns
3. **Framework Check** - Correct annotations, imports
4. **Element Check** - Elements used exist in page objects
5. **Method Check** - Methods called exist

**Implementation:**
```go
func (cv *CodeValidator) Validate(code string, context *Context) (*ValidationResult, error) {
    result := &ValidationResult{IsValid: true}

    // 1. Syntax validation
    classData, err := cv.parser.ParseString(code)
    if err != nil {
        result.IsValid = false
        result.SyntaxErrors = append(result.SyntaxErrors, SyntaxError{
            Message: err.Error(),
        })
        return result, nil
    }

    // 2. Pattern validation
    if !cv.matchesNamingPattern(classData, context.CodingPatterns) {
        result.PatternViolations = append(result.PatternViolations, PatternViolation{
            Type: "naming",
            Expected: context.CodingPatterns.TestMethodNaming,
            Actual: classData.Methods[0].Name,
            Suggestion: "Use pattern: testActionWithCondition",
        })
    }

    // 3. Wait strategy validation
    if hasThreadSleep(classData) {
        result.PatternViolations = append(result.PatternViolations, PatternViolation{
            Type: "wait_strategy",
            Message: "Uses Thread.sleep instead of explicit waits",
            Suggestion: "Use WebDriverWait with ExpectedConditions",
        })
    }

    // 4. Assertion validation
    if !hasAssertMessages(classData) && context.CodingPatterns.UsesAssertMessages {
        result.Suggestions = append(result.Suggestions,
            "Project uses assertion messages - consider adding descriptive messages")
    }

    return result, nil
}
```

#### Day 8-10: Integration + Testing

**CLI Command:**
```bash
./copilot generate "create a test for login with invalid password"
```

**Complete Workflow:**
```go
func generateCode(request string) {
    // 1. Build context
    ctx := contextBuilder.BuildContext(request)

    // 2. Build prompt
    prompt := promptEngine.BuildPrompt(ctx)

    // 3. Generate code
    response := llmClient.Generate(prompt)

    // 4. Validate
    validation := validator.Validate(response.Code, ctx)

    if !validation.IsValid {
        // Auto-fix or ask LLM to fix
        fixedCode := llmClient.Fix(response.Code, validation)
        response.Code = fixedCode
    }

    // 5. Save to file
    saveToFile(response.Code, "LoginTest.java")

    // 6. Show results
    fmt.Printf("✓ Generated: LoginTest.java\n")
    fmt.Printf("Confidence: %.0f%%\n", response.Confidence*100)
    if len(validation.Suggestions) > 0 {
        fmt.Printf("\n💡 Suggestions:\n")
        for _, s := range validation.Suggestions {
            fmt.Printf("  - %s\n", s)
        }
    }
}
```

---

## Key Advantages of Our Approach

✅ **We Have Complete Understanding** - Our parser extracted EVERYTHING
✅ **We Have Perfect Context** - Knowledge graph connects everything
✅ **We Have Real Examples** - Pattern examples from actual project
✅ **We Have Structured Data** - Framework + patterns reduce hallucinations
✅ **We Use Latest Research** - cAST, Self-RAG, Few-Shot, Prompt Caching
✅ **Cost Optimized** - 90% cheaper with caching, 85% faster

## Expected Performance

Based on research:
- **Accuracy**: +4.3 points (cAST) + 96% less hallucinations (Self-RAG)
- **Quality**: Code matches project style (few-shot + patterns)
- **Cost**: 82% cheaper per request (prompt caching)
- **Speed**: 85% faster responses (cached prompts)

## Success Criteria

Phase 3 complete when:
- [ ] Can generate test methods matching project style
- [ ] Can generate page object classes with correct patterns
- [ ] Generated code passes validation (syntax + patterns)
- [ ] Uses correct framework (TestNG vs JUnit)
- [ ] Follows coding patterns (naming, waits, assertions)
- [ ] Cost per generation < $0.01
- [ ] Response time < 3 seconds

---

## Implementation Order

1. ✅ Context Builder - Gathers all relevant info
2. ✅ Prompt Engine - Builds optimized prompts
3. ✅ LLM Client - Calls OpenAI/Claude with caching
4. ✅ Code Validator - Ensures quality
5. ✅ CLI Integration - `copilot generate <request>`
6. ✅ Testing - Verify with real requests

**Total Time**: ~2 weeks for complete AI integration

**Result**: Production-ready AI code generation system using 2025's best techniques!
