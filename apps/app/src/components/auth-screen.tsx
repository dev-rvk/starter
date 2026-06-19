import { Link } from "@tanstack/react-router";
import type { ReactNode } from "react";

/** AuthScreen centres an auth card within the viewport. */
export function AuthScreen({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-1 items-center justify-center p-6">{children}</div>
  );
}

/**
 * AuthDisabledNotice is shown on auth routes when Clerk is not configured, so
 * the app still renders sensibly with the auth feature toggled off.
 */
export function AuthDisabledNotice() {
  return (
    <AuthScreen>
      <div className="max-w-sm space-y-3 text-center">
        <h1 className="font-semibold text-xl">Authentication is disabled</h1>
        <p className="text-muted-foreground text-sm">
          Set <code>VITE_CLERK_PUBLISHABLE_KEY</code> in{" "}
          <code>apps/app/.env.local</code> to enable sign-in. Until then the
          dashboard is open in development.
        </p>
        <Link className="text-sm underline underline-offset-4" to="/">
          Go to dashboard
        </Link>
      </div>
    </AuthScreen>
  );
}
