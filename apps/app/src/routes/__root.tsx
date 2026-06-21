import { isClerkEnabled, SignedIn, UserButton, useLocalAuth } from "@repo/auth";
import { ModeToggle } from "@repo/design-system";
import { Button } from "@repo/design-system/components/ui/button";
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
          {isClerkEnabled() ? <ClerkUserControls /> : <LocalUserControls />}
        </div>
      </header>
      <main className="flex flex-1 flex-col">
        <Outlet />
      </main>
    </div>
  );
}

/** Clerk mode — uses Clerk's UserButton component. */
function ClerkUserControls() {
  return (
    <SignedIn>
      <UserButton />
    </SignedIn>
  );
}

/** Local auth mode — shows email + sign out button when signed in. */
function LocalUserControls() {
  const { isSignedIn, user, signOut } = useLocalAuth();

  if (!(isSignedIn && user)) {
    return null;
  }

  return (
    <div className="flex items-center gap-2">
      <span className="text-muted-foreground text-sm">{user.email}</span>
      <Button onClick={signOut} size="sm" variant="outline">
        Sign out
      </Button>
    </div>
  );
}
