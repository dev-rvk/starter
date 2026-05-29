import { isAuthEnabled, useSignIn } from "@repo/auth";
import { ForgotPasswordForm } from "@repo/design-system/components/auth/forgot-password-form";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { AuthDisabledNotice, AuthScreen } from "../components/auth-screen";
import { extractClerkError } from "../lib/clerk-error";

export const Route = createFileRoute("/forgot-password")({
  component: ForgotPasswordRoute,
});

function ForgotPasswordRoute() {
  if (!isAuthEnabled()) {
    return <AuthDisabledNotice />;
  }
  return <ForgotPasswordScreen />;
}

function ForgotPasswordScreen() {
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
        success={success}
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
      />
    </AuthScreen>
  );
}
