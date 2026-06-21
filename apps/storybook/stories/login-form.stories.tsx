import type { Meta, StoryObj } from "@storybook/react";
import { LoginForm } from "@repo/design-system/components/auth/login-form";

const meta: Meta<typeof LoginForm> = {
  title: "Auth/LoginForm",
  component: LoginForm,
  parameters: {
    layout: "centered",
  },
  args: {
    onSubmit: async (values) => {
      console.log("Submit:", values);
      return new Promise((resolve) => setTimeout(resolve, 1000));
    },
    onForgotPassword: () => console.log("Forgot password clicked"),
    onSignUp: () => console.log("Sign up clicked"),
  },
};

export default meta;
type Story = StoryObj<typeof LoginForm>;

export const Default: Story = {};

export const Loading: Story = {
  args: {
    isLoading: true,
  },
};

export const WithError: Story = {
  args: {
    error: "Invalid email or password.",
  },
};
