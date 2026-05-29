/**
 * Typed API client generated from the Go backend's OpenAPI spec.
 *
 * Regenerate types with `bun run generate` (swag -> swagger2openapi ->
 * openapi-typescript). The client is a thin wrapper over openapi-fetch that
 * pairs well with TanStack Query.
 */
import createClient, { type Middleware } from "openapi-fetch";
import type { paths } from "./schema";

export type { paths } from "./schema";

export type CreateApiClientOptions = {
  /** Base URL of the API, e.g. http://localhost:3002 */
  baseUrl: string;
  /** Optional async getter for a bearer token (e.g. Clerk session token). */
  getToken?: () => Promise<string | null | undefined>;
};

/** createApiClient builds a typed client, injecting an auth token when provided. */
export function createApiClient({ baseUrl, getToken }: CreateApiClientOptions) {
  const client = createClient<paths>({ baseUrl });

  if (getToken) {
    const authMiddleware: Middleware = {
      async onRequest({ request }) {
        const token = await getToken();
        if (token) {
          request.headers.set("Authorization", `Bearer ${token}`);
        }
        return request;
      },
    };
    client.use(authMiddleware);
  }

  return client;
}

export type ApiClient = ReturnType<typeof createApiClient>;
