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
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@repo/design-system/components/ui/form";
import { Input } from "@repo/design-system/components/ui/input";
import { useForm } from "react-hook-form";
import { z } from "zod";

/**
 * Field rules mirror the backend domain constraints (see apps/api domain
 * value objects). Keep these in sync with the Go validators.
 */
export const signUpSchema = z.object({
  username: z
    .string()
    .min(2, "Username must be at least 2 characters.")
    .max(6, "Username must be at most 6 characters.")
    .regex(/^[a-zA-Z0-9_]+$/, "Letters, numbers and underscores only."),
  email: z.string().email("Enter a valid email address."),
  password: z.string().min(8, "Password must be at least 8 characters."),
});

export type SignUpFormValues = z.infer<typeof signUpSchema>;

export interface SignUpFormProps {
  error?: string | null;
  isLoading?: boolean;
  onSignIn?: () => void;
  onSubmit: (values: SignUpFormValues) => void | Promise<void>;
}

export function SignUpForm({
  onSubmit,
  isLoading = false,
  error,
  onSignIn,
}: SignUpFormProps) {
  const form = useForm<SignUpFormValues>({
    resolver: zodResolver(signUpSchema),
    defaultValues: { username: "", email: "", password: "" },
  });

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle className="text-2xl">Create your account</CardTitle>
        <CardDescription>Start your free account in seconds.</CardDescription>
      </CardHeader>
      <Form {...form}>
        <form noValidate onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent className="grid gap-4">
            <FormField
              control={form.control}
              name="username"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Username</FormLabel>
                  <FormControl>
                    <Input
                      autoComplete="username"
                      placeholder="jdoe"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>2–6 characters.</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
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
            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Password</FormLabel>
                  <FormControl>
                    <Input
                      autoComplete="new-password"
                      type="password"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
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
            <Button className="w-full" disabled={isLoading} type="submit">
              {isLoading ? "Creating account…" : "Create account"}
            </Button>
            {onSignIn ? (
              <p className="text-center text-muted-foreground text-sm">
                Already have an account?{" "}
                <button
                  className="underline underline-offset-4"
                  onClick={onSignIn}
                  type="button"
                >
                  Sign in
                </button>
              </p>
            ) : null}
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}
