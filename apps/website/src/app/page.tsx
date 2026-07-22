import { MessageSquare, ShieldCheck, ScrollText } from 'lucide-react';
import { Button, Card, CardContent, CardHeader, CardTitle, CardDescription } from '@shared/ui';

const features = [
  {
    icon: MessageSquare,
    title: 'Talk to any page',
    description: 'Ask Charli in the side panel; it understands the page and helps you act.',
  },
  {
    icon: ShieldCheck,
    title: 'Safe by design',
    description: 'The model proposes; the backend decides. Risky actions always confirm.',
  },
  {
    icon: ScrollText,
    title: 'Full audit log',
    description: 'Every action Charli takes is recorded — nothing happens silently.',
  },
];

export default function Home() {
  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-16 px-6 py-20">
      <section className="flex flex-col items-start gap-6">
        <span className="rounded-full border px-3 py-1 text-xs font-medium text-muted-foreground">
          Browser agent · beta
        </span>
        <h1 className="text-4xl font-semibold tracking-tight sm:text-5xl">
          Meet <span className="text-primary">Charli</span>.
        </h1>
        <p className="max-w-xl text-lg text-muted-foreground">
          Your flexible browser agent. From rewriting a sentence to multi-step tasks — Charli works
          right where you already are.
        </p>
        <div className="flex flex-wrap gap-3">
          <Button size="lg">Add to browser</Button>
          <Button size="lg" variant="secondary">
            View console
          </Button>
        </div>
      </section>

      <section className="grid gap-6 sm:grid-cols-3">
        {features.map(({ icon: Icon, title, description }) => (
          <Card key={title}>
            <CardHeader>
              <div className="flex size-9 items-center justify-center rounded-lg bg-secondary text-primary">
                <Icon className="size-4" />
              </div>
              <CardTitle className="mt-2">{title}</CardTitle>
              <CardDescription>{description}</CardDescription>
            </CardHeader>
            <CardContent />
          </Card>
        ))}
      </section>
    </main>
  );
}
