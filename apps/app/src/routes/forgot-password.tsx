import { isClerkEnabled, useSignIn } from "@repo/auth";
import { ForgotPasswordForm } from "@repo/design-system/components/auth/forgot-password-form";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { AuthScreen } from "../components/auth-screen";
import { extractClerkError } from "../lib/clerk-error";

export const Route = createFileRoute("/forgot-password")({
  component: ForgotPasswordRoute,
});

function ForgotPasswordRoute() {
  if (isClerkEnabled()) {
    return <ClerkForgotPasswordScreen />;
  }
  return <LocalForgotPasswordNotice />;
}

/** Clerk-powered forgot password (existing behaviour). */
function ClerkForgotPasswordScreen() {
  const { signIn, isLoaded } = useSignIn();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  return (
    <AuthScreen>
      <ForgotPasswordForm
        error={error}
        isLoading={loading}
        onBackToSignIn={() => navigate({ to: "/sign-in" })}
        onSubmit={async ({ email }) => {
          if (!isLoaded) {
            return;
          }
          setError(null);
          setLoading(true);
          try {
            await signIn.create({
              strategy: "reset_password_email_code",
              identifier: email,
            });
            setSuccess(true);
          } catch (err) {
            setError(extractClerkError(err));
          } finally {
            setLoading(false);
          }
        }}
        success={success}
      />
    </AuthScreen>
  );
}

/**
 * LocalForgotPasswordNotice is shown when local auth is active — password
 * reset requires email infrastructure (Clerk/Resend) which isn't available
 * in local auth mode.
 */
function LocalForgotPasswordNotice() {
  return (
    <AuthScreen>
      <div className="max-w-sm space-y-3 text-center">
        <h1 className="font-semibold text-xl">Password reset unavailable</h1>
        <p className="text-muted-foreground text-sm">
          Password reset is not available with local authentication. Contact
          your administrator to reset your password, or enable Clerk for
          full-featured auth with email-based password reset.
        </p>
        <Link className="text-sm underline underline-offset-4" to="/sign-in">
          Back to sign in
        </Link>
      </div>
    </AuthScreen>
  );
}
