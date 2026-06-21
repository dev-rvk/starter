/**
 * Central feature-toggle map. A feature is enabled only when its key(s) are
 * present in the environment, mirroring the backend's toggle convention. Import
 * `features` anywhere to branch on availability; mount providers conditionally.
 *
 * Auth is always available: when `clerk` is true, Clerk handles auth; when
 * false, local username/password auth is used against the Go backend.
 */
const env = import.meta.env;

export const features = {
  /** When true, Clerk is the auth provider. When false, local auth is used. */
  clerk: Boolean(env.VITE_CLERK_PUBLISHABLE_KEY),
  googleAnalytics: Boolean(env.VITE_GA_MEASUREMENT_ID),
  posthog: Boolean(env.VITE_POSTHOG_KEY),
  sentry: Boolean(env.VITE_SENTRY_DSN),
} as const;

export type FeatureName = keyof typeof features;

export const apiUrl: string = env.VITE_API_URL ?? "http://localhost:3002";
