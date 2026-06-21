import { isClerkEnabled, useLocalAuth, useSignUp } from "@repo/auth";
import { SignUpForm } from "@repo/design-system/components/auth/sign-up-form";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { AuthScreen } from "../components/auth-screen";
import { extractClerkError } from "../lib/clerk-error";

export const Route = createFileRoute("/sign-up")({ component: SignUpRoute });

function SignUpRoute() {
  if (isClerkEnabled()) {
    return <ClerkSignUpScreen />;
  }
  return <LocalSignUpScreen />;
}

/** Clerk-powered sign-up (existing behaviour). */
function ClerkSignUpScreen() {
  const { signUp, setActive, isLoaded } = useSignUp();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  return (
    <AuthScreen>
      <SignUpForm
        error={error}
        isLoading={loading}
        onSignIn={() => navigate({ to: "/sign-in" })}
        onSubmit={async ({ username, email, password }) => {
          if (!isLoaded) {
            return;
          }
          setError(null);
          setLoading(true);
          try {
            const result = await signUp.create({
              username,
              emailAddress: email,
              password,
            });
            if (result.status === "complete") {
              await setActive({ session: result.createdSessionId });
              navigate({ to: "/" });
            } else {
              await signUp.prepareEmailAddressVerification({
                strategy: "email_code",
              });
              setError(
                "Check your email to verify your address, then sign in."
              );
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

/** Local username/password sign-up. */
function LocalSignUpScreen() {
  const { signUp } = useLocalAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  return (
    <AuthScreen>
      <SignUpForm
        error={error}
        isLoading={loading}
        onSignIn={() => navigate({ to: "/sign-in" })}
        onSubmit={async ({ username, email, password }) => {
          setError(null);
          setLoading(true);
          try {
            await signUp(username, email, password);
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
