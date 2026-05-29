"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@repo/design-system/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@repo/design-system/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@repo/design-system/components/ui/form";
import { Input } from "@repo/design-system/components/ui/input";

export const forgotPasswordSchema = z.object({
  email: z.string().email("Enter a valid email address."),
});

export type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;

export type ForgotPasswordFormProps = {
  onSubmit: (values: ForgotPasswordFormValues) => void | Promise<void>;
  isLoading?: boolean;
  error?: string | null;
  success?: boolean;
  onBackToSignIn?: () => void;
};

export function ForgotPasswordForm({
  onSubmit,
  isLoading = false,
  error,
  success = false,
  onBackToSignIn,
}: ForgotPasswordFormProps) {
  const form = useForm<ForgotPasswordFormValues>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: { email: "" },
  });

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle className="text-2xl">Reset your password</CardTitle>
        <CardDescription>
          We&apos;ll email you a link to reset your password.
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent className="grid gap-4">
            {success ? (
              <p className="text-sm" role="status">
                If an account exists for that email, a reset link is on its way.
              </p>
            ) : (
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input
                        autoComplete="email"
                        placeholder="you@example.com"
                        type="email"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            {error ? (
              <p className="text-destructive text-sm" role="alert">
                {error}
              </p>
            ) : null}
          </CardContent>
          <CardFooter className="flex flex-col gap-3">
            {!success && (
              <Button className="w-full" disabled={isLoading} type="submit">
                {isLoading ? "Sending…" : "Send reset link"}
              </Button>
            )}
            {onBackToSignIn ? (
              <button
                className="text-center text-muted-foreground text-sm underline underline-offset-4"
                onClick={onBackToSignIn}
                type="button"
              >
                Back to sign in
              </button>
            ) : null}
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}
