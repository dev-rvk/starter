import { SignUpForm } from "@repo/design-system/components/auth/sign-up-form";
import type { Meta, StoryObj } from "@storybook/react";

const meta: Meta<typeof SignUpForm> = {
  title: "Auth/SignUpForm",
  component: SignUpForm,
  parameters: {
    layout: "centered",
  },
  args: {
    onSubmit: async (values) => {
      console.log("Submit:", values);
      await new Promise((resolve) => setTimeout(resolve, 1000));
    },
    onSignIn: () => console.log("Sign in clicked"),
  },
};

export default meta;
type Story = StoryObj<typeof SignUpForm>;

export const Default: Story = {};

export const Loading: Story = {
  args: {
    isLoading: true,
  },
};

export const WithError: Story = {
  args: {
    error: "Username is already taken.",
  },
};
