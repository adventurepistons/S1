# 🧠 GPT-Powered Intelligent Analysis Approach

## 🎯 Core Principle

**User gives us:** Recording (pages + elements + interactions)
**GPT figures out:** Everything else!

---

## 📊 Information Flow

```
┌─────────────────────────────────┐
│   User Records Page             │
│   (Browser Extension)           │
└────────────┬────────────────────┘
             │
             │ Recording JSON
             ▼
┌─────────────────────────────────┐
│   GPT Analysis Layer            │
│   "Think and Infer"             │
└────────────┬────────────────────┘
             │
             ├─► Page Type (login, checkout, etc.)
             ├─► Validation Rules (email format, required fields)
             ├─► Test Scenarios (positive, negative, edge cases)
             ├─► Success Criteria (what indicates success/failure)
             ├─► Test Data (valid/invalid examples)
             └─► Framework Structure (which Page Objects needed)
             │
             ▼
┌─────────────────────────────────┐
│   Code Generation               │
│   (Selenium Framework)          │
└─────────────────────────────────┘
```

---

## 🔍 Example: Login Page Recording

### **What User Provides (Recording JSON):**

```json
{
  "url": "https://app.example.com/login",
  "title": "Login - MyApp",
  "elements": [
    {
      "id": "email",
      "type": "email",
      "placeholder": "Enter your email",
      "required": true,
      "name": "email"
    },
    {
      "id": "password",
      "type": "password",
      "placeholder": "Password",
      "required": true,
      "minLength": 8
    },
    {
      "id": "remember-me",
      "type": "checkbox",
      "label": "Remember me"
    },
    {
      "id": "login-btn",
      "type": "submit",
      "text": "Sign In"
    },
    {
      "id": "forgot-password",
      "type": "link",
      "text": "Forgot Password?",
      "href": "/reset-password"
    },
    {
      "id": "signup-link",
      "type": "link",
      "text": "Create Account",
      "href": "/register"
    }
  ],
  "interactions": [
    {
      "action": "input",
      "element": "email",
      "value": "test@example.com"
    },
    {
      "action": "input",
      "element": "password",
      "value": "Test1234!"
    },
    {
      "action": "click",
      "element": "login-btn"
    },
    {
      "action": "navigate",
      "url": "https://app.example.com/dashboard"
    }
  ]
}
```

---

### **GPT Analysis Prompt:**

```
You are a senior QA engineer analyzing a recorded web page interaction.

RECORDING DATA:
URL: https://app.example.com/login
Title: Login - MyApp

Elements Found:
1. Email field (id: email, type: email, required: true)
2. Password field (id: password, type: password, required: true, minLength: 8)
3. Remember me checkbox (id: remember-me, optional)
4. Login button (id: login-btn, submit)
5. Forgot password link (id: forgot-password, href: /reset-password)
6. Sign up link (id: signup-link, href: /register)

User Interaction:
1. Entered email: test@example.com
2. Entered password: Test1234!
3. Clicked login button
4. Navigated to: https://app.example.com/dashboard

YOUR TASK:
Based ONLY on this recording data, infer and generate:

1. PAGE TYPE & PURPOSE
   - What type of page is this?
   - What is the primary functionality?

2. VALIDATION RULES
   - What validation rules can you infer from element attributes?
   - What error scenarios are likely?

3. SUCCESS CRITERIA
   - How do we know login succeeded?
   - What page/elements indicate success?

4. FAILURE CRITERIA
   - How do we know login failed?
   - What should we look for?

5. TEST SCENARIOS
   - Positive: All valid inputs that should work
   - Negative: Invalid inputs that should fail
   - Edge cases: Boundary conditions
   - Security: Injection, XSS, etc.

6. TEST DATA
   - Generate realistic test data for each scenario
   - Include valid and invalid examples

7. VALIDATION ELEMENTS TO CHECK
   - What elements should exist on success?
   - What elements indicate failure?

Output as structured JSON.
```

---

### **GPT Response (Inferred Intelligence):**

```json
{
  "pageAnalysis": {
    "type": "login",
    "purpose": "User authentication",
    "confidence": 0.95
  },

  "inferredValidationRules": {
    "email": {
      "required": true,
      "format": "email",
      "reasoning": "type='email' attribute indicates email validation"
    },
    "password": {
      "required": true,
      "minLength": 8,
      "reasoning": "minLength attribute explicitly set to 8"
    },
    "rememberMe": {
      "required": false,
      "reasoning": "checkbox, no required attribute"
    }
  },

  "successCriteria": {
    "navigation": {
      "expectedUrl": "https://app.example.com/dashboard",
      "reasoning": "User navigated here after successful login"
    },
    "elementsToCheck": [
      {
        "selector": "id=user-profile",
        "expectedState": "visible",
        "reasoning": "Dashboard typically shows user profile"
      },
      {
        "selector": "id=logout-btn",
        "expectedState": "visible",
        "reasoning": "Logout button indicates authenticated state"
      },
      {
        "text": "Welcome",
        "expectedState": "present",
        "reasoning": "Welcome message common on post-login page"
      }
    ]
  },

  "failureCriteria": {
    "navigation": {
      "expectedUrl": "https://app.example.com/login",
      "reasoning": "Stay on login page if authentication fails"
    },
    "elementsToCheck": [
      {
        "selector": "class=error-message",
        "expectedState": "visible",
        "reasoning": "Error message should appear"
      },
      {
        "text": "Invalid credentials",
        "expectedState": "present",
        "reasoning": "Common error message pattern"
      }
    ]
  },

  "testScenarios": [
    {
      "category": "positive",
      "name": "Valid login with correct credentials",
      "priority": "high",
      "testData": {
        "email": "valid.user@example.com",
        "password": "ValidPass123!",
        "rememberMe": false
      },
      "expectedResult": "success"
    },
    {
      "category": "positive",
      "name": "Valid login with remember me checked",
      "priority": "medium",
      "testData": {
        "email": "valid.user@example.com",
        "password": "ValidPass123!",
        "rememberMe": true
      },
      "expectedResult": "success"
    },
    {
      "category": "negative",
      "name": "Invalid password",
      "priority": "high",
      "testData": {
        "email": "valid.user@example.com",
        "password": "WrongPassword",
        "rememberMe": false
      },
      "expectedResult": "failure",
      "reasoning": "Wrong password should be rejected"
    },
    {
      "category": "negative",
      "name": "Empty email",
      "priority": "high",
      "testData": {
        "email": "",
        "password": "ValidPass123!",
        "rememberMe": false
      },
      "expectedResult": "failure",
      "reasoning": "Email is required (required attribute present)"
    },
    {
      "category": "negative",
      "name": "Empty password",
      "priority": "high",
      "testData": {
        "email": "valid.user@example.com",
        "password": "",
        "rememberMe": false
      },
      "expectedResult": "failure",
      "reasoning": "Password is required (required attribute present)"
    },
    {
      "category": "negative",
      "name": "Invalid email format",
      "priority": "medium",
      "testData": {
        "email": "not-an-email",
        "password": "ValidPass123!",
        "rememberMe": false
      },
      "expectedResult": "failure",
      "reasoning": "type='email' requires valid email format"
    },
    {
      "category": "edgeCase",
      "name": "Password less than minimum length",
      "priority": "medium",
      "testData": {
        "email": "valid.user@example.com",
        "password": "Short1!",
        "rememberMe": false
      },
      "expectedResult": "failure",
      "reasoning": "minLength=8, this password has only 7 chars"
    },
    {
      "category": "edgeCase",
      "name": "Very long email (boundary test)",
      "priority": "low",
      "testData": {
        "email": "verylongemailaddressthatexceedstypicallimits@example.com",
        "password": "ValidPass123!",
        "rememberMe": false
      },
      "expectedResult": "success",
      "reasoning": "Valid format, should work unless length limit exists"
    },
    {
      "category": "edgeCase",
      "name": "Special characters in password",
      "priority": "low",
      "testData": {
        "email": "valid.user@example.com",
        "password": "P@$$w0rd!#%&*",
        "rememberMe": false
      },
      "expectedResult": "success",
      "reasoning": "Special characters should be allowed in passwords"
    },
    {
      "category": "security",
      "name": "SQL Injection in email",
      "priority": "high",
      "testData": {
        "email": "admin' OR '1'='1",
        "password": "anypassword",
        "rememberMe": false
      },
      "expectedResult": "failure",
      "reasoning": "SQL injection should be prevented"
    },
    {
      "category": "security",
      "name": "XSS in email",
      "priority": "high",
      "testData": {
        "email": "<script>alert('XSS')</script>@example.com",
        "password": "ValidPass123!",
        "rememberMe": false
      },
      "expectedResult": "failure",
      "reasoning": "XSS payload should be sanitized/rejected"
    },
    {
      "category": "security",
      "name": "XSS in password",
      "priority": "medium",
      "testData": {
        "email": "valid.user@example.com",
        "password": "<script>alert('XSS')</script>",
        "rememberMe": false
      },
      "expectedResult": "failure",
      "reasoning": "XSS in password should fail authentication"
    }
  ],

  "relatedFeatures": {
    "forgotPassword": {
      "detected": true,
      "link": "/reset-password",
      "reasoning": "Forgot password link found, should test this flow separately"
    },
    "registration": {
      "detected": true,
      "link": "/register",
      "reasoning": "Sign up link found, registration flow exists"
    }
  },

  "recommendedPageObjects": [
    {
      "name": "LoginPage",
      "elements": ["email", "password", "rememberMe", "loginBtn", "forgotPasswordLink"],
      "methods": ["login", "clickForgotPassword", "isLoginButtonDisplayed"]
    },
    {
      "name": "DashboardPage",
      "elements": ["userProfile", "logoutBtn", "welcomeMessage"],
      "methods": ["isLoggedIn", "getWelcomeText", "logout"],
      "reasoning": "Needed to validate successful login"
    }
  ]
}
```

---

## 🔑 Key Intelligence Patterns

### **1. Element Attribute Analysis**

GPT infers from HTML attributes:
```javascript
{
  "type": "email"     → Email format validation required
  "required": true    → Field cannot be empty
  "minLength": 8      → Minimum 8 characters
  "maxLength": 50     → Maximum 50 characters
  "pattern": "regex"  → Custom validation pattern
}
```

### **2. Navigation Pattern Analysis**

```
Login page → Dashboard page = SUCCESS
Login page → Login page + error = FAILURE
```

GPT learns:
- Success = URL changes to /dashboard
- Failure = Stays on /login with error message

### **3. Context Inference**

From page title "Login - MyApp" and elements:
- This is a login page (high confidence)
- Primary action: Authentication
- Related features: Password reset, Registration

### **4. Security Test Generation**

GPT knows common vulnerabilities:
- SQL Injection: `admin' OR '1'='1`
- XSS: `<script>alert('XSS')</script>`
- CSRF: Test without proper tokens
- Brute force: Multiple failed attempts

### **5. Validation Element Prediction**

GPT predicts what to look for post-action:

**Success indicators:**
- Welcome message
- Logout button
- User profile
- Dashboard elements

**Failure indicators:**
- Error message div
- "Invalid credentials" text
- Red border on fields
- Login form still visible

---

## 🧪 Example: Checkout Page (More Complex)

### **Recording Data:**
```json
{
  "url": "https://shop.example.com/checkout",
  "elements": [
    {"id": "shipping-address", "type": "textarea", "required": true},
    {"id": "city", "type": "text", "required": true},
    {"id": "zipcode", "type": "text", "pattern": "[0-9]{5}"},
    {"id": "card-number", "type": "text", "pattern": "[0-9]{16}"},
    {"id": "cvv", "type": "password", "pattern": "[0-9]{3}"},
    {"id": "place-order-btn", "type": "submit", "text": "Place Order"}
  ]
}
```

### **GPT Infers:**

```json
{
  "pageType": "checkout",
  "complexity": "high",

  "testScenarios": [
    "Valid checkout with all fields",
    "Missing shipping address",
    "Invalid zip code format (letters instead of numbers)",
    "Invalid credit card (Luhn algorithm check)",
    "Invalid CVV (less than 3 digits)",
    "Card number with spaces",
    "Expired card date",
    "International address (if applicable)",
    "PO Box address (might be restricted)",
    "Special characters in address",
    "SQL injection in address field",
    "Card number: 0000000000000000 (invalid)",
    "Negative: incomplete form submission"
  ],

  "validationRules": {
    "zipcode": {
      "pattern": "5 digits only",
      "examples": {
        "valid": ["12345", "90210"],
        "invalid": ["1234", "abcde", "12-345"]
      }
    },
    "cardNumber": {
      "pattern": "16 digits",
      "additionalCheck": "Luhn algorithm",
      "examples": {
        "valid": ["4111111111111111", "5500000000000004"],
        "invalid": ["1111111111111111", "1234567812345678"]
      }
    }
  },

  "successCriteria": {
    "navigation": "/order-confirmation",
    "elements": [
      "Order number visible",
      "Confirmation message",
      "Email sent notice"
    ]
  }
}
```

**GPT even knows:**
- Luhn algorithm for credit card validation
- Standard zip code formats
- PO Box restrictions
- International address patterns

---

## 🎯 Implementation: Scenario Generator with GPT Intelligence

### **File:** `copilot-core/pkg/scenario/scenario_generator.go`

```go
package scenario

import (
    "encoding/json"
    "fmt"
    "github.com/yourusername/copilot-core/pkg/llm"
)

type ScenarioGenerator struct {
    llmClient *llm.Client
}

type Recording struct {
    URL          string                 `json:"url"`
    Title        string                 `json:"title"`
    Elements     []RecordedElement      `json:"elements"`
    Interactions []RecordedInteraction  `json:"interactions"`
}

type RecordedElement struct {
    ID          string            `json:"id"`
    Type        string            `json:"type"`
    Name        string            `json:"name,omitempty"`
    Placeholder string            `json:"placeholder,omitempty"`
    Required    bool              `json:"required,omitempty"`
    Pattern     string            `json:"pattern,omitempty"`
    MinLength   int               `json:"minLength,omitempty"`
    MaxLength   int               `json:"maxLength,omitempty"`
    Label       string            `json:"label,omitempty"`
    Text        string            `json:"text,omitempty"`
    Attributes  map[string]string `json:"attributes,omitempty"`
}

type RecordedInteraction struct {
    Action  string `json:"action"` // "input", "click", "navigate"
    Element string `json:"element"`
    Value   string `json:"value,omitempty"`
    URL     string `json:"url,omitempty"`
}

type IntelligentAnalysis struct {
    PageAnalysis         PageAnalysis         `json:"pageAnalysis"`
    ValidationRules      map[string]Rule      `json:"inferredValidationRules"`
    SuccessCriteria      Criteria             `json:"successCriteria"`
    FailureCriteria      Criteria             `json:"failureCriteria"`
    TestScenarios        []TestScenario       `json:"testScenarios"`
    RelatedFeatures      map[string]Feature   `json:"relatedFeatures"`
    RecommendedPageObjects []PageObjectSpec   `json:"recommendedPageObjects"`
}

type TestScenario struct {
    Category       string                 `json:"category"` // positive, negative, edgeCase, security
    Name           string                 `json:"name"`
    Priority       string                 `json:"priority"` // high, medium, low
    TestData       map[string]interface{} `json:"testData"`
    ExpectedResult string                 `json:"expectedResult"` // success, failure
    Reasoning      string                 `json:"reasoning"`
}

func (g *ScenarioGenerator) AnalyzeRecording(recording *Recording) (*IntelligentAnalysis, error) {
    // Build GPT prompt
    prompt := g.buildAnalysisPrompt(recording)

    // Call GPT for intelligent analysis
    response, err := g.llmClient.Chat(prompt)
    if err != nil {
        return nil, fmt.Errorf("GPT analysis failed: %w", err)
    }

    // Parse GPT response
    var analysis IntelligentAnalysis
    if err := json.Unmarshal([]byte(response), &analysis); err != nil {
        return nil, fmt.Errorf("failed to parse GPT response: %w", err)
    }

    return &analysis, nil
}

func (g *ScenarioGenerator) buildAnalysisPrompt(recording *Recording) string {
    prompt := `You are a senior QA engineer analyzing a recorded web page interaction.

RECORDING DATA:
URL: ` + recording.URL + `
Title: ` + recording.Title + `

Elements Found:
`

    for i, elem := range recording.Elements {
        prompt += fmt.Sprintf("%d. %s (id: %s, type: %s",
            i+1, elem.Label, elem.ID, elem.Type)

        if elem.Required {
            prompt += ", required: true"
        }
        if elem.MinLength > 0 {
            prompt += fmt.Sprintf(", minLength: %d", elem.MinLength)
        }
        if elem.Pattern != "" {
            prompt += fmt.Sprintf(", pattern: %s", elem.Pattern)
        }
        prompt += ")\n"
    }

    prompt += `
User Interaction:
`

    for i, interaction := range recording.Interactions {
        switch interaction.Action {
        case "input":
            prompt += fmt.Sprintf("%d. Entered %s: %s\n",
                i+1, interaction.Element, interaction.Value)
        case "click":
            prompt += fmt.Sprintf("%d. Clicked %s\n",
                i+1, interaction.Element)
        case "navigate":
            prompt += fmt.Sprintf("%d. Navigated to: %s\n",
                i+1, interaction.URL)
        }
    }

    prompt += `
YOUR TASK:
Based ONLY on this recording data, infer and generate:

1. PAGE TYPE & PURPOSE
   - What type of page is this? (login, registration, checkout, search, etc.)
   - What is the primary functionality?
   - Confidence level (0.0 to 1.0)

2. VALIDATION RULES
   - Infer validation rules from element attributes
   - Consider: required, type, pattern, minLength, maxLength
   - Generate realistic valid and invalid test data

3. SUCCESS CRITERIA
   - How do we know the action succeeded?
   - What URL should we navigate to?
   - What elements should be visible/present?

4. FAILURE CRITERIA
   - How do we know the action failed?
   - What error messages might appear?
   - What elements indicate failure?

5. TEST SCENARIOS
   Generate comprehensive test scenarios:
   - Positive: Valid inputs that should succeed
   - Negative: Invalid inputs that should fail
   - Edge cases: Boundary conditions, special chars
   - Security: SQL injection, XSS, CSRF

   For EACH scenario provide:
   - Category (positive/negative/edgeCase/security)
   - Name (descriptive)
   - Priority (high/medium/low)
   - Test data (actual values)
   - Expected result (success/failure)
   - Reasoning (why this scenario matters)

6. RELATED FEATURES
   - Detect links to related pages (forgot password, registration, etc.)
   - Suggest additional flows to test

7. PAGE OBJECTS
   - Recommend Page Object classes needed
   - List elements for each Page Object
   - List methods for each Page Object

Output ONLY valid JSON matching this exact schema:
{
  "pageAnalysis": {
    "type": "string",
    "purpose": "string",
    "confidence": 0.0
  },
  "inferredValidationRules": {
    "fieldName": {
      "required": bool,
      "format": "string",
      "minLength": int,
      "maxLength": int,
      "pattern": "string",
      "reasoning": "string"
    }
  },
  "successCriteria": {
    "navigation": {
      "expectedUrl": "string",
      "reasoning": "string"
    },
    "elementsToCheck": [
      {
        "selector": "string",
        "expectedState": "visible|present|hidden",
        "text": "string (optional)",
        "reasoning": "string"
      }
    ]
  },
  "failureCriteria": {
    "navigation": {
      "expectedUrl": "string",
      "reasoning": "string"
    },
    "elementsToCheck": [...]
  },
  "testScenarios": [
    {
      "category": "positive|negative|edgeCase|security",
      "name": "string",
      "priority": "high|medium|low",
      "testData": {
        "fieldName": "value"
      },
      "expectedResult": "success|failure",
      "reasoning": "string"
    }
  ],
  "relatedFeatures": {
    "featureName": {
      "detected": bool,
      "link": "string",
      "reasoning": "string"
    }
  },
  "recommendedPageObjects": [
    {
      "name": "string",
      "elements": ["string"],
      "methods": ["string"],
      "reasoning": "string (optional)"
    }
  ]
}

IMPORTANT:
- Base ALL inferences on the recording data provided
- Be thorough - generate at least 10-15 test scenarios
- Think like a senior QA engineer
- Consider security vulnerabilities
- Output ONLY the JSON, no markdown formatting
`

    return prompt
}
```

---

## ✅ **This Approach is PERFECT Because:**

1. **Zero User Effort**
   - User: Records page ✅
   - AI: Does ALL the thinking ✅

2. **Intelligent Inference**
   - Analyzes element attributes
   - Predicts validation rules
   - Generates realistic test data

3. **Comprehensive Coverage**
   - 10-15+ scenarios per recording
   - Positive + Negative + Edge + Security
   - Things humans forget!

4. **Self-Sufficient**
   - No config files needed
   - No manual scenario writing
   - Just record → AI → tests

---

## 🚀 **Next Step:**

Should I implement the `ScenarioGenerator` with this GPT intelligence approach?

This will be the core "brain" that makes everything else work! 🧠