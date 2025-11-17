import * as vscode from 'vscode';
import { auth, db, initializeFirebase } from './firebase';
import { signInWithPopup, GoogleAuthProvider, User, signOut } from 'firebase/auth';
import { doc, getDoc, setDoc, updateDoc } from 'firebase/firestore';

export interface UserData {
    uid: string;
    email: string;
    displayName: string;
    tier: 'free' | 'pro' | 'team';
    apiKeyMode: 'byok' | 'local' | 'managed';
    frameworks: string[];
    monthlyUsage: number;
    usageLimit: number;
    createdAt: string;
    lastLoginAt: string;
}

export class AuthService {
    private context: vscode.ExtensionContext;
    private currentUser: User | null = null;
    private userData: UserData | null = null;

    constructor(context: vscode.ExtensionContext) {
        this.context = context;
        initializeFirebase();
    }

    // Sign in with Google
    async signIn(): Promise<UserData | null> {
        try {
            const provider = new GoogleAuthProvider();
            const result = await signInWithPopup(auth, provider);
            const user = result.user;

            // Store user
            this.currentUser = user;

            // Get or create user data
            const userData = await this.getOrCreateUserData(user);
            this.userData = userData;

            // Store token securely
            const token = await user.getIdToken();
            await this.context.secrets.store('userToken', token);
            await this.context.secrets.store('userId', user.uid);

            vscode.window.showInformationMessage(`Welcome ${user.displayName}!`);

            return userData;
        } catch (error) {
            vscode.window.showErrorMessage(`Sign in failed: ${error}`);
            return null;
        }
    }

    // Sign out
    async signOutUser(): Promise<void> {
        try {
            await signOut(auth);
            this.currentUser = null;
            this.userData = null;

            await this.context.secrets.delete('userToken');
            await this.context.secrets.delete('userId');

            vscode.window.showInformationMessage('Signed out successfully');
        } catch (error) {
            vscode.window.showErrorMessage(`Sign out failed: ${error}`);
        }
    }

    // Get current user
    getCurrentUser(): User | null {
        return this.currentUser;
    }

    // Get user data
    getUserData(): UserData | null {
        return this.userData;
    }

    // Check if signed in
    async isSignedIn(): Promise<boolean> {
        const token = await this.context.secrets.get('userToken');
        return !!token;
    }

    // Get stored token
    async getToken(): Promise<string | null> {
        return await this.context.secrets.get('userToken') || null;
    }

    // Get or create user data in Firestore
    private async getOrCreateUserData(user: User): Promise<UserData> {
        const userRef = doc(db, 'users', user.uid);
        const userSnap = await getDoc(userRef);

        if (userSnap.exists()) {
            // Update last login
            await updateDoc(userRef, {
                lastLoginAt: new Date().toISOString()
            });

            return userSnap.data() as UserData;
        } else {
            // Create new user
            const newUserData: UserData = {
                uid: user.uid,
                email: user.email || '',
                displayName: user.displayName || 'User',
                tier: 'free',
                apiKeyMode: 'byok',
                frameworks: ['selenium-java'],  // Free tier: only Selenium
                monthlyUsage: 0,
                usageLimit: -1,  // Unlimited for BYOK
                createdAt: new Date().toISOString(),
                lastLoginAt: new Date().toISOString()
            };

            await setDoc(userRef, newUserData);

            return newUserData;
        }
    }

    // Refresh user data from Firestore
    async refreshUserData(): Promise<UserData | null> {
        if (!this.currentUser) {
            return null;
        }

        const userRef = doc(db, 'users', this.currentUser.uid);
        const userSnap = await getDoc(userRef);

        if (userSnap.exists()) {
            this.userData = userSnap.data() as UserData;
            return this.userData;
        }

        return null;
    }

    // Check if framework is allowed for user's tier
    isFrameworkAllowed(framework: string): boolean {
        if (!this.userData) {
            return false;
        }

        return this.userData.frameworks.includes(framework);
    }

    // Check if usage limit reached
    isUsageLimitReached(): boolean {
        if (!this.userData) {
            return true;
        }

        // BYOK mode has no limit
        if (this.userData.apiKeyMode === 'byok' || this.userData.apiKeyMode === 'local') {
            return false;
        }

        // Check managed API limit
        if (this.userData.usageLimit === -1) {
            return false;  // Unlimited
        }

        return this.userData.monthlyUsage >= this.userData.usageLimit;
    }

    // Get tier display info
    getTierInfo(): { name: string; color: string; features: string[] } {
        const tier = this.userData?.tier || 'free';

        switch (tier) {
            case 'free':
                return {
                    name: 'FREE',
                    color: '#888888',
                    features: [
                        'Selenium/Java support',
                        'BYOK or Local LLM',
                        'Smart RAG system (BYOK)',
                        'Unlimited usage (BYOK)'
                    ]
                };
            case 'pro':
                return {
                    name: 'PRO',
                    color: '#4CAF50',
                    features: [
                        '5+ frameworks',
                        'Managed API (no key needed)',
                        '500 generations/month',
                        'Priority support'
                    ]
                };
            case 'team':
                return {
                    name: 'TEAM',
                    color: '#2196F3',
                    features: [
                        'Everything in Pro',
                        'Unlimited generations',
                        'Team workspace',
                        'Shared templates',
                        'Admin dashboard'
                    ]
                };
        }
    }

    // Show upgrade dialog
    async showUpgradeDialog(reason: string): Promise<void> {
        const action = await vscode.window.showWarningMessage(
            reason,
            'Upgrade to Pro',
            'Learn More',
            'Cancel'
        );

        if (action === 'Upgrade to Pro') {
            vscode.env.openExternal(vscode.Uri.parse('https://testcopilot.dev/upgrade'));
        } else if (action === 'Learn More') {
            vscode.env.openExternal(vscode.Uri.parse('https://testcopilot.dev/pricing'));
        }
    }
}
