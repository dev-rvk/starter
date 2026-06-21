import { ForgotPasswordForm } from "@repo/design-system/components/auth/forgot-password-form";
import type { Meta, StoryObj } from "@storybook/react";

const meta: Meta<typeof ForgotPasswordForm> = {
  title: "Auth/ForgotPasswordForm",
  component: ForgotPasswordForm,
  parameters: {
    layout: "centered",
  },
  args: {
    onSubmit: async (values) => {
      console.log("Submit:", values);
      await new Promise((resolve) => setTimeout(resolve, 1000));
    },
    onBackToSignIn: () => console.log("Back to sign in clicked"),
  },
};

export default meta;
type Story = StoryObj<typeof ForgotPasswordForm>;

export const Default: Story = {};

export const Loading: Story = {
  args: {
    isLoading: true,
  },
};

export const WithError: Story = {
  args: {
    error: "No account found with that email address.",
  },
};

export const Success: Story = {
  args: {
    success: true,
  },
};
