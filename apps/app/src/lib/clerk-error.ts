/** extractClerkError pulls a human-readable message out of a Clerk API error. */
export function extractClerkError(err: unknown): string {
  if (
    typeof err === "object" &&
    err !== null &&
    "errors" in err &&
    Array.isArray((err as { errors: unknown[] }).errors)
  ) {
    const first = (err as { errors: { message?: string }[] }).errors[0];
    if (first?.message) {
      return first.message;
    }
  }
  return "Something went wrong. Please try again.";
}
