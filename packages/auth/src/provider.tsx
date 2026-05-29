"use client";

import { ClerkProvider } from "@clerk/clerk-react";
import type { ReactNode } from "react";

/**
 * The Clerk publishable key, read from the Vite environment. Empty when auth is
 * not configured.
 */
export const publishableKey: string =
  (import.meta.env.VITE_CLERK_PUBLISHABLE_KEY as string | undefined) ?? "";

/** isAuthEnabled reports whether Clerk is configured (a key is present). */
export const isAuthEnabled = (): boolean => publishableKey.length > 0;

/** useIsAuthEnabled is the hook form of {@link isAuthEnabled}. */
export const useIsAuthEnabled = (): boolean => isAuthEnabled();

type AuthProviderProps = {
  children: ReactNode;
  /** Where Clerk should send users after sign-out, etc. */
  afterSignOutUrl?: string;
};

/**
 * AuthProvider mounts Clerk only when a publishable key is present. Without a
 * key it renders children directly, leaving the app fully functional with auth
 * disabled (the feature toggle is "off").
 */
export function AuthProvider({
  children,
  afterSignOutUrl = "/",
}: AuthProviderProps) {
  if (!isAuthEnabled()) {
    return <>{children}</>;
  }
  return (
    <ClerkProvider
      afterSignOutUrl={afterSignOutUrl}
      publishableKey={publishableKey}
    >
      {children}
    </ClerkProvider>
  );
}
