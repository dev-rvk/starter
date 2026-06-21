"use client";

export {
  SignedIn,
  SignedOut,
  UserButton,
  useAuth,
  useClerk,
  useSignIn,
  useSignUp,
  useUser,
} from "@clerk/clerk-react";
export type { LocalAuthContextValue, LocalUser } from "./local-auth";

export { LocalAuthProvider, useLocalAuth } from "./local-auth";
/**
 * Authentication package — supports two modes based on feature toggles:
 *
 * 1. **Clerk mode** (VITE_CLERK_PUBLISHABLE_KEY set): Full Clerk auth with
 *    OAuth, MFA, etc. {@link isClerkEnabled} reports true.
 *
 * 2. **Local mode** (no Clerk key): Username/password auth against the Go
 *    backend. Uses JWT tokens stored in localStorage. {@link isClerkEnabled}
 *    reports false.
 *
 * {@link isAuthEnabled} always reports true — auth is always available.
 */
export {
  AuthProvider,
  isAuthEnabled,
  isClerkEnabled,
  publishableKey,
  useIsAuthEnabled,
  useIsClerkEnabled,
} from "./provider";
