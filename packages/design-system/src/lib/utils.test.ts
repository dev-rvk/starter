import { describe, it, expect } from "vitest";
import { cn } from "./utils";

describe("cn utility", () => {
  it("merges tailwind classes correctly", () => {
    expect(cn("bg-red-500", "text-center", "bg-blue-500")).toBe("text-center bg-blue-500");
  });

  it("handles conditional classes correctly", () => {
    expect(cn("text-sm", true && "text-lg")).toBe("text-lg");
  });
});
