import { isClerkEnabled, useAuth, useLocalAuth } from "@repo/auth";
import { Navigate } from "@tanstack/react-router";
import type { ReactNode } from "react";

/**
 * RequireAuth gates its children behind the active auth provider. Auth is always
 * enabled (Clerk or local), so the only choice is which hook to read. While the
 * session is loading it shows a centred message; when signed out it redirects to
 * the sign-in route.
 */
export function RequireAuth({ children }: { children: ReactNode }) {
  if (isClerkEnabled()) {
    return <ClerkGuard>{children}</ClerkGuard>;
  }
  return <LocalGuard>{children}</LocalGuard>;
}

function ClerkGuard({ children }: { children: ReactNode }) {
  const { isLoaded, isSignedIn } = useAuth();
  return renderGuard(isLoaded, isSignedIn, children);
}

function LocalGuard({ children }: { children: ReactNode }) {
  const { isLoaded, isSignedIn } = useLocalAuth();
  return renderGuard(isLoaded, isSignedIn, children);
}

function renderGuard(
  isLoaded: boolean,
  isSignedIn: boolean,
  children: ReactNode
) {
  if (!isLoaded) {
    return (
      <div className="flex flex-1 items-center justify-center p-6 text-muted-foreground">
        Loading…
      </div>
    );
  }
  if (!isSignedIn) {
    return <Navigate to="/sign-in" />;
  }
  return <>{children}</>;
}
