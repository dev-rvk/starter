import { Button } from "@repo/design-system/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@repo/design-system/components/ui/card";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/pricing")({
  component: PricingPage,
});

const tiers = [
  {
    name: "Hobby",
    price: "$0",
    description: "For side projects.",
    features: ["1 project", "Community support"],
  },
  {
    name: "Pro",
    price: "$20",
    description: "For growing teams.",
    features: ["Unlimited projects", "Email support", "Analytics"],
  },
];

function PricingPage() {
  return (
    <section className="mx-auto w-full max-w-4xl space-y-8 p-6">
      <div className="space-y-2 text-center">
        <h1 className="font-bold text-3xl">Pricing</h1>
        <p className="text-muted-foreground">
          Simple, transparent pricing. (Stripe checkout wired later.)
        </p>
      </div>
      <div className="grid gap-6 sm:grid-cols-2">
        {tiers.map((tier) => (
          <Card key={tier.name}>
            <CardHeader>
              <CardTitle>{tier.name}</CardTitle>
              <CardDescription>{tier.description}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              <p className="font-bold text-3xl">
                {tier.price}
                <span className="text-base text-muted-foreground">/mo</span>
              </p>
              <ul className="text-muted-foreground text-sm">
                {tier.features.map((f) => (
                  <li key={f}>• {f}</li>
                ))}
              </ul>
            </CardContent>
            <CardFooter>
              <Button className="w-full">Choose {tier.name}</Button>
            </CardFooter>
          </Card>
        ))}
      </div>
    </section>
  );
}
