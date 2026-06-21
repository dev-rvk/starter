import { type ApiClient, createApiClient } from "@repo/api-client";
import { isClerkEnabled, useAuth, useLocalAuth } from "@repo/auth";
import { createContext, type ReactNode, useContext, useMemo } from "react";
import { apiUrl } from "../features";

const ApiContext = createContext<ApiClient | null>(null);

/** AuthedApiProvider injects the Clerk session token into API requests. */
function ClerkApiProvider({ children }: { children: ReactNode }) {
  const { getToken } = useAuth();
  const client = useMemo(
    () =>
      createApiClient({
        baseUrl: `${apiUrl}/api/v1`,
        getToken: () => getToken(),
      }),
    [getToken]
  );
  return <ApiContext.Provider value={client}>{children}</ApiContext.Provider>;
}

/** LocalApiProvider injects the local JWT into API requests. */
function LocalApiProvider({ children }: { children: ReactNode }) {
  const { getToken } = useLocalAuth();
  const client = useMemo(
    () =>
      createApiClient({
        baseUrl: `${apiUrl}/api/v1`,
        getToken,
      }),
    [getToken]
  );
  return <ApiContext.Provider value={client}>{children}</ApiContext.Provider>;
}

/**
 * ApiProvider picks the right client variant. When Clerk is configured, inject
 * the Clerk session token. Otherwise, inject the local JWT. Both modes produce
 * a typed API client with bearer auth.
 */
export function ApiProvider({ children }: { children: ReactNode }) {
  if (isClerkEnabled()) {
    return <ClerkApiProvider>{children}</ClerkApiProvider>;
  }
  return <LocalApiProvider>{children}</LocalApiProvider>;
}

/** useApiClient returns the typed API client from context. */
export function useApiClient(): ApiClient {
  const client = useContext(ApiContext);
  if (!client) {
    throw new Error("useApiClient must be used within <ApiProvider>");
  }
  return client;
}
