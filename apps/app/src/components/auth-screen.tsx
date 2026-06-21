import type { ReactNode } from "react";

/** AuthScreen centres an auth card within the viewport. */
export function AuthScreen({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-1 items-center justify-center p-6">
      {children}
    </div>
  );
}
