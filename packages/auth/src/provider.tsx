"use client";

import { ClerkProvider } from "@clerk/clerk-react";
import type { ReactNode } from "react";
import { LocalAuthProvider } from "./local-auth";

/**
 * The Clerk publishable key, read from the Vite environment. Empty when auth is
 * not configured.
 */
export const publishableKey: string =
  (import.meta.env.VITE_CLERK_PUBLISHABLE_KEY as string | undefined) ?? "";

/**
 * The API URL, used by local auth to call the backend auth endpoints.
 */
const apiUrl: string =
  (import.meta.env.VITE_API_URL as string | undefined) ??
  "http://localhost:3002";

/** isAuthEnabled always reports true — auth is always available (Clerk or local). */
export const isAuthEnabled = (): boolean => true;

/** isClerkEnabled reports whether Clerk is configured (a key is present). */
export const isClerkEnabled = (): boolean => publishableKey.length > 0;

/** useIsAuthEnabled is the hook form of {@link isAuthEnabled}. */
export const useIsAuthEnabled = (): boolean => isAuthEnabled();

/** useIsClerkEnabled is the hook form of {@link isClerkEnabled}. */
export const useIsClerkEnabled = (): boolean => isClerkEnabled();

interface AuthProviderProps {
  /** Where Clerk should send users after sign-out, etc. */
  afterSignOutUrl?: string;
  children: ReactNode;
}

/**
 * AuthProvider mounts Clerk when a publishable key is present, or the local
 * auth provider otherwise. Auth is always available — the choice is between
 * Clerk (full-featured, OAuth, etc.) and local (username/password against the DB).
 */
export function AuthProvider({
  children,
  afterSignOutUrl = "/",
}: AuthProviderProps) {
  if (isClerkEnabled()) {
    return (
      <ClerkProvider
        afterSignOutUrl={afterSignOutUrl}
        publishableKey={publishableKey}
      >
        {children}
      </ClerkProvider>
    );
  }
  return (
    <LocalAuthProvider apiUrl={`${apiUrl}/api/v1`}>
      {children}
    </LocalAuthProvider>
  );
}
