import { isAuthEnabled, useAuth } from "@repo/auth";
import { Button } from "@repo/design-system/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@repo/design-system/components/ui/card";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Navigate } from "@tanstack/react-router";
import { apiUrl } from "../features";

export const Route = createFileRoute("/")({
  component: DashboardPage,
});

type ApiUser = {
  id: string;
  username: string;
  email: string;
  createdAt: string;
};

function DashboardPage() {
  // When Clerk is configured, gate the dashboard behind authentication.
  if (isAuthEnabled()) {
    return <GuardedDashboard />;
  }
  return <Dashboard />;
}

function GuardedDashboard() {
  const { isLoaded, isSignedIn } = useAuth();
  if (!isLoaded) {
    return <CenteredMessage>Loading…</CenteredMessage>;
  }
  if (!isSignedIn) {
    return <Navigate to="/sign-in" />;
  }
  return <Dashboard />;
}

function Dashboard() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["users"],
    queryFn: async (): Promise<ApiUser[]> => {
      const res = await fetch(`${apiUrl}/api/v1/users`);
      if (!res.ok) {
        throw new Error(`API error ${res.status}`);
      }
      return res.json();
    },
  });

  return (
    <div className="mx-auto w-full max-w-3xl space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="font-bold text-2xl">Dashboard</h1>
        <Button onClick={() => refetch()} variant="outline">
          Refresh
        </Button>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Users</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? <p>Loading users…</p> : null}
          {error ? (
            <p className="text-destructive text-sm">
              Could not reach the API. Is it running on {apiUrl}?
            </p>
          ) : null}
          {data && data.length === 0 ? (
            <p className="text-muted-foreground text-sm">No users yet.</p>
          ) : null}
          {data && data.length > 0 ? (
            <ul className="divide-y">
              {data.map((u) => (
                <li className="flex justify-between py-2 text-sm" key={u.id}>
                  <span className="font-medium">{u.username}</span>
                  <span className="text-muted-foreground">{u.email}</span>
                </li>
              ))}
            </ul>
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}

function CenteredMessage({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-1 items-center justify-center p-6 text-muted-foreground">
      {children}
    </div>
  );
}
