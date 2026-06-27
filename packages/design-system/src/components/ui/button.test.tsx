/// <reference types="@testing-library/jest-dom" />
import * as React from "react";
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { Button } from "./button";

describe("Button component", () => {
  it("renders the children passed to it", () => {
    render(<Button>Click me</Button>);
    expect(screen.getByRole("button", { name: "Click me" })).toBeInTheDocument();
  });

  it("applies the default variant classes", () => {
    render(<Button>Default</Button>);
    const button = screen.getByRole("button", { name: "Default" });
    // This is just a generic check to ensure some tailwind base class is present.
    // The actual class might vary based on your shadcn configuration, 
    // so if this fails, it's a good test to assess!
    expect(button.className).toContain("inline-flex");
  });

  it("can be disabled", () => {
    render(<Button disabled>Disabled</Button>);
    const button = screen.getByRole("button", { name: "Disabled" });
    expect(button).toBeDisabled();
  });
});
