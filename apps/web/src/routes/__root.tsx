import { ModeToggle } from "@repo/design-system";
import { Button } from "@repo/design-system/components/ui/button";
import { createRootRoute, Link, Outlet } from "@tanstack/react-router";

const appUrl: string =
  (import.meta.env.VITE_APP_URL as string | undefined) ??
  "http://localhost:3000";

export const Route = createRootRoute({
  component: MarketingLayout,
});

function MarketingLayout() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="flex items-center justify-between border-b px-6 py-3">
        <Link className="font-semibold" to="/">
          Starterpack
        </Link>
        <nav className="flex items-center gap-4 text-sm">
          <Link to="/pricing">Pricing</Link>
          <ModeToggle />
          <Button asChild size="sm">
            <a href={appUrl}>Sign in</a>
          </Button>
        </nav>
      </header>
      <main className="flex flex-1 flex-col">
        <Outlet />
      </main>
      <footer className="border-t px-6 py-4 text-muted-foreground text-sm">
        © {new Date().getFullYear()} Starterpack.
      </footer>
    </div>
  );
}
