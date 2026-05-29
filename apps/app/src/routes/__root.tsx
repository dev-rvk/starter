import { isAuthEnabled, SignedIn, UserButton } from "@repo/auth";
import { ModeToggle } from "@repo/design-system";
import { createRootRoute, Link, Outlet } from "@tanstack/react-router";

export const Route = createRootRoute({
  component: RootLayout,
});

function RootLayout() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="flex items-center justify-between border-b px-6 py-3">
        <Link className="font-semibold" to="/">
          Starterpack
        </Link>
        <div className="flex items-center gap-3">
          <ModeToggle />
          {isAuthEnabled() ? (
            <SignedIn>
              <UserButton />
            </SignedIn>
          ) : null}
        </div>
      </header>
      <main className="flex flex-1 flex-col">
        <Outlet />
      </main>
    </div>
  );
}
