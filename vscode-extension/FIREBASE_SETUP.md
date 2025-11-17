# Firebase Setup Guide

This guide will help you set up Firebase for Test Automation Copilot's authentication and user management system.

## Step 1: Create Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Click "Add Project"
3. Enter project name: `test-automation-copilot`
4. Disable Google Analytics (optional)
5. Click "Create Project"

## Step 2: Enable Authentication

1. In Firebase Console, go to **Authentication**
2. Click "Get Started"
3. Click "Sign-in method" tab
4. Enable **Google** provider:
   - Click on "Google"
   - Toggle "Enable"
   - Set Project support email
   - Click "Save"

## Step 3: Create Firestore Database

1. In Firebase Console, go to **Firestore Database**
2. Click "Create database"
3. Select "Start in **production mode**"
4. Choose location (e.g., `us-central`)
5. Click "Enable"

## Step 4: Set Firestore Rules

Replace the default rules with:

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Users collection
    match /users/{userId} {
      // Users can read/write their own data
      allow read, write: if request.auth != null && request.auth.uid == userId;

      // Allow creation for any authenticated user
      allow create: if request.auth != null;
    }
  }
}
```

Click "Publish"

## Step 5: Get Firebase Configuration

1. In Firebase Console, go to **Project Settings** (gear icon)
2. Scroll to "Your apps"
3. Click **Web** icon (`</>`)
4. Register app:
   - App nickname: `vscode-extension`
   - Don't check "Firebase Hosting"
   - Click "Register app"
5. **Copy the configuration** - you'll need this!

It looks like:
```javascript
const firebaseConfig = {
  apiKey: "AIzaSy...",
  authDomain: "test-automation-copilot.firebaseapp.com",
  projectId: "test-automation-copilot",
  storageBucket: "test-automation-copilot.appspot.com",
  messagingSenderId: "123456789",
  appId: "1:123456789:web:abc123"
};
```

## Step 6: Update Extension Code

Open `src/firebase.ts` and replace the config:

```typescript
const firebaseConfig = {
    apiKey: "YOUR_API_KEY",              // Replace
    authDomain: "YOUR_PROJECT_ID.firebaseapp.com",  // Replace
    projectId: "YOUR_PROJECT_ID",        // Replace
    storageBucket: "YOUR_PROJECT_ID.appspot.com",  // Replace
    messagingSenderId: "YOUR_SENDER_ID", // Replace
    appId: "YOUR_APP_ID"                 // Replace
};
```

## Step 7: Install Dependencies

```bash
cd vscode-extension
npm install
```

This installs Firebase SDK (~10MB).

## Step 8: Test Authentication

1. Compile extension: `npm run compile`
2. Press F5 to launch Extension Development Host
3. Run command: `Test Copilot: Sign In`
4. Google sign-in popup should appear
5. Sign in with your Google account
6. Check Firestore - you should see a new user document!

## Firestore Schema

### Users Collection (`users/{uid}`)

```javascript
{
  uid: string,              // Firebase Auth UID
  email: string,            // User's email
  displayName: string,      // User's display name
  tier: 'free' | 'pro' | 'team',
  apiKeyMode: 'byok' | 'local' | 'managed',
  frameworks: string[],     // Allowed frameworks
  monthlyUsage: number,     // Number of generations this month
  usageLimit: number,       // Monthly limit (-1 = unlimited)
  createdAt: string,        // ISO timestamp
  lastLoginAt: string       // ISO timestamp
}
```

### Example Document

```javascript
{
  uid: "abc123...",
  email: "user@example.com",
  displayName: "John Doe",
  tier: "free",
  apiKeyMode: "byok",
  frameworks: ["selenium-java"],
  monthlyUsage: 0,
  usageLimit: -1,
  createdAt: "2025-01-15T10:30:00.000Z",
  lastLoginAt: "2025-01-15T10:30:00.000Z"
}
```

## Security Notes

1. **Never commit Firebase config to public repos**
   - Add `src/firebase.ts` to `.gitignore` if config has secrets
   - Or use environment variables

2. **Firestore rules are enforced**
   - Users can only read/write their own data
   - No one can read other users' data

3. **API key is safe in client code**
   - Firebase API keys are not secrets
   - They identify your project, not authenticate it
   - Security comes from Firestore rules

## Troubleshooting

### "Firebase not initialized"
- Make sure you called `initializeFirebase()` in extension.ts
- Check console for initialization errors

### "Auth popup blocked"
- Browser may block popups
- Try clicking status bar to trigger auth again

### "Permission denied" in Firestore
- Check Firestore rules are set correctly
- Make sure user is authenticated
- Verify UID matches document ID

### "Module not found: firebase"
- Run `npm install` in vscode-extension folder
- Check package.json has firebase dependency
- Run `npm run compile`

## Next Steps

Once Firebase is set up:

1. ✅ Users can sign in with Google
2. ✅ User data stored in Firestore
3. ✅ Tier system working
4. ✅ Framework restrictions working

Next phase: Build backend API for managed API mode!
