import { Button } from "@repo/design-system/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@repo/design-system/components/ui/card";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { RequireAuth } from "../components/require-auth";
import { apiUrl } from "../features";
import { useApiClient } from "../lib/api";

export const Route = createFileRoute("/")({
  component: DashboardPage,
});

function DashboardPage() {
  return (
    <RequireAuth>
      <Dashboard />
    </RequireAuth>
  );
}

function Dashboard() {
  const api = useApiClient();
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["users"],
    queryFn: async () => {
      const { data: users, error: apiError } = await api.GET("/users");
      if (apiError) {
        throw new Error("API error");
      }
      return users;
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
