"use client";

/**
 * Authentication package — a thin wrapper around Clerk that honours the
 * feature-toggle convention: when no publishable key is configured, Clerk is
 * not mounted and {@link useIsAuthEnabled} reports false, so apps can degrade
 * gracefully (e.g. bypass protection in local dev).
 */
export {
  AuthProvider,
  isAuthEnabled,
  useIsAuthEnabled,
  publishableKey,
} from "./provider";

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
