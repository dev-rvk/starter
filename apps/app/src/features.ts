/**
 * Central feature-toggle map. A feature is enabled only when its key(s) are
 * present in the environment, mirroring the backend's toggle convention. Import
 * `features` anywhere to branch on availability; mount providers conditionally.
 */
const env = import.meta.env;

export const features = {
  clerk: Boolean(env.VITE_CLERK_PUBLISHABLE_KEY),
  googleAnalytics: Boolean(env.VITE_GA_MEASUREMENT_ID),
  posthog: Boolean(env.VITE_POSTHOG_KEY),
  sentry: Boolean(env.VITE_SENTRY_DSN),
} as const;

export type FeatureName = keyof typeof features;

export const apiUrl: string = env.VITE_API_URL ?? "http://localhost:3002";
