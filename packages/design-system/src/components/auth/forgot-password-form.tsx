"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
import { useForm } from "react-hook-form";
import { z } from "zod";

export const forgotPasswordSchema = z.object({
  email: z.string().email("Enter a valid email address."),
});

export type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;

export interface ForgotPasswordFormProps {
  error?: string | null;
  isLoading?: boolean;
  onBackToSignIn?: () => void;
  onSubmit: (values: ForgotPasswordFormValues) => void | Promise<void>;
  success?: boolean;
}

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
        <form noValidate onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent className="grid gap-4">
            {success ? (
              <output className="text-sm">
                If an account exists for that email, a reset link is on its way.
              </output>
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
          </CardContent>
          <CardFooter className="flex flex-col gap-3">
            {error ? (
              <div
                className="w-full rounded-md bg-destructive/15 p-3 text-center font-medium text-destructive text-sm"
                role="alert"
              >
                {error}
              </div>
            ) : null}
            {!success && (
              <Button className="w-full" disabled={isLoading} type="submit">
                {isLoading ? "Sending…" : "Send reset link"}
              </Button>
            )}
            {onBackToSignIn ? (
              <Button
                className="h-auto p-0 text-muted-foreground text-sm"
                onClick={onBackToSignIn}
                type="button"
                variant="link"
              >
                Back to sign in
              </Button>
            ) : null}
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}
