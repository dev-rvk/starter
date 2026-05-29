import { Button } from "@repo/design-system/components/ui/button";
import { createFileRoute } from "@tanstack/react-router";

const appUrl: string =
  (import.meta.env.VITE_APP_URL as string | undefined) ??
  "http://localhost:3000";

export const Route = createFileRoute("/")({
  component: HomePage,
});

function HomePage() {
  return (
    <section className="mx-auto flex max-w-3xl flex-1 flex-col items-center justify-center gap-6 p-6 text-center">
      <h1 className="font-bold text-4xl tracking-tight sm:text-5xl">
        Ship your SaaS faster
      </h1>
      <p className="max-w-xl text-lg text-muted-foreground">
        A deployable Turborepo starter: Vite + TanStack Router frontend, a Go
        hexagonal backend, a shadcn design system, and feature-toggled
        integrations.
      </p>
      <div className="flex gap-3">
        <Button asChild size="lg">
          <a href={appUrl}>Get started</a>
        </Button>
        <Button asChild size="lg" variant="outline">
          <a href="https://github.com">View on GitHub</a>
        </Button>
      </div>
    </section>
  );
}
