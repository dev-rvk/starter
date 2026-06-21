import { isClerkEnabled, useLocalAuth, useSignIn } from "@repo/auth";
import { LoginForm } from "@repo/design-system/components/auth/login-form";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { AuthScreen } from "../components/auth-screen";
import { extractClerkError } from "../lib/clerk-error";

export const Route = createFileRoute("/sign-in")({ component: SignInRoute });

function SignInRoute() {
  if (isClerkEnabled()) {
    return <ClerkSignInScreen />;
  }
  return <LocalSignInScreen />;
}

/** Clerk-powered sign-in (existing behaviour). */
function ClerkSignInScreen() {
  const { signIn, setActive, isLoaded } = useSignIn();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  return (
    <AuthScreen>
      <LoginForm
        error={error}
        isLoading={loading}
        onForgotPassword={() => navigate({ to: "/forgot-password" })}
        onSignUp={() => navigate({ to: "/sign-up" })}
        onSubmit={async ({ email, password }) => {
          if (!isLoaded) {
            return;
          }
          setError(null);
          setLoading(true);
          try {
            const result = await signIn.create({
              identifier: email,
              password,
            });
            if (result.status === "complete") {
              await setActive({ session: result.createdSessionId });
              navigate({ to: "/" });
            } else {
              setError("Additional verification required.");
            }
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

/** Local username/password sign-in. */
function LocalSignInScreen() {
  const { signIn } = useLocalAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  return (
    <AuthScreen>
      <LoginForm
        error={error}
        isLoading={loading}
        onForgotPassword={() => navigate({ to: "/forgot-password" })}
        onSignUp={() => navigate({ to: "/sign-up" })}
        onSubmit={async ({ email, password }) => {
          setError(null);
          setLoading(true);
          try {
            await signIn(email, password);
            navigate({ to: "/" });
          } catch (err) {
            setError(
              err instanceof Error ? err.message : "Something went wrong."
            );
          } finally {
            setLoading(false);
          }
        }}
      />
    </AuthScreen>
  );
}
