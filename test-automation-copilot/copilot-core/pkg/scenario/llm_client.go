package scenario

// LLMClient defines the interface for LLM interactions
type LLMClient interface {
	Chat(prompt string) (string, error)
}

// MockLLMClient is a mock implementation for testing
type MockLLMClient struct {
	Response string
	Error    error
}

// Chat implements the LLMClient interface
func (m *MockLLMClient) Chat(prompt string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	if m.Response != "" {
		return m.Response, nil
	}

	// Default mock response for testing
	return `{
  "pageAnalysis": {
    "type": "login",
    "purpose": "User authentication",
    "confidence": 0.95
  },
  "inferredValidationRules": {
    "email": {
      "required": true,
      "format": "email",
      "reasoning": "Inferred from type='email' attribute"
    },
    "password": {
      "required": true,
      "minLength": 8,
      "reasoning": "Inferred from required and minLength attributes"
    }
  },
  "successCriteria": {
    "navigation": {
      "expectedUrl": "https://app.example.com/dashboard",
      "reasoning": "User navigated here after successful login"
    },
    "elementsToCheck": [
      {
        "selector": "id=welcome-message",
        "expectedState": "visible",
        "text": "Welcome",
        "reasoning": "Welcome message indicates successful login"
      }
    ]
  },
  "failureCriteria": {
    "navigation": {
      "expectedUrl": "https://app.example.com/login",
      "reasoning": "Stays on login page when authentication fails"
    },
    "elementsToCheck": [
      {
        "selector": "class=error-message",
        "expectedState": "visible",
        "reasoning": "Error message should appear on failure"
      }
    ]
  },
  "testScenarios": [
    {
      "category": "positive",
      "name": "Valid login with correct credentials",
      "priority": "high",
      "testData": {
        "email": "user@example.com",
        "password": "ValidPass123!"
      },
      "expectedResult": "success",
      "reasoning": "Primary happy path test"
    },
    {
      "category": "negative",
      "name": "Invalid password",
      "priority": "high",
      "testData": {
        "email": "user@example.com",
        "password": "wrongpassword"
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
        "password": "ValidPass123!"
      },
      "expectedResult": "failure",
      "reasoning": "Email is required"
    },
    {
      "category": "security",
      "name": "SQL injection in email",
      "priority": "high",
      "testData": {
        "email": "admin' OR '1'='1",
        "password": "anypass"
      },
      "expectedResult": "failure",
      "reasoning": "SQL injection should be prevented"
    }
  ],
  "relatedFeatures": {
    "forgotPassword": {
      "detected": true,
      "link": "/reset-password",
      "reasoning": "Forgot password link found"
    }
  },
  "recommendedPageObjects": [
    {
      "name": "LoginPage",
      "elements": ["emailField", "passwordField", "loginButton"],
      "methods": ["login", "isLoginButtonVisible", "getErrorMessage"]
    },
    {
      "name": "DashboardPage",
      "elements": ["welcomeMessage", "logoutButton"],
      "methods": ["isLoggedIn", "getWelcomeText"]
    }
  ]
}`, nil
}
