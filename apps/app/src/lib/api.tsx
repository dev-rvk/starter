import { isAuthEnabled, useAuth } from "@repo/auth";
import { type ApiClient, createApiClient } from "@repo/api-client";
import {
  createContext,
  type ReactNode,
  useContext,
  useMemo,
} from "react";
import { apiUrl } from "../features";

const ApiContext = createContext<ApiClient | null>(null);

/** AuthedApiProvider injects the Clerk session token into API requests. */
function AuthedApiProvider({ children }: { children: ReactNode }) {
  const { getToken } = useAuth();
  const client = useMemo(
    () => createApiClient({ baseUrl: `${apiUrl}/api/v1`, getToken: () => getToken() }),
    [getToken]
  );
  return <ApiContext.Provider value={client}>{children}</ApiContext.Provider>;
}

/** PlainApiProvider builds a client without auth (used when Clerk is disabled). */
function PlainApiProvider({ children }: { children: ReactNode }) {
  const client = useMemo(
    () => createApiClient({ baseUrl: `${apiUrl}/api/v1` }),
    []
  );
  return <ApiContext.Provider value={client}>{children}</ApiContext.Provider>;
}

/**
 * ApiProvider picks the right client variant. The choice is based on the
 * env-derived auth toggle, which is constant for the app's lifetime, so the
 * selected provider component is stable across renders.
 */
export function ApiProvider({ children }: { children: ReactNode }) {
  if (isAuthEnabled()) {
    return <AuthedApiProvider>{children}</AuthedApiProvider>;
  }
  return <PlainApiProvider>{children}</PlainApiProvider>;
}

/** useApiClient returns the typed API client from context. */
export function useApiClient(): ApiClient {
  const client = useContext(ApiContext);
  if (!client) {
    throw new Error("useApiClient must be used within <ApiProvider>");
  }
  return client;
}
