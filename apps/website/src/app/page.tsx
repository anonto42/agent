import {
  MessageSquare,
  ShieldCheck,
  MousePointerClick,
  Download,
  Puzzle,
  PanelRight,
  Sparkles,
} from 'lucide-react';
import { Button, Card, CardContent, CardHeader, CardTitle, CardDescription } from '@shared/ui';

const EXTENSION_ZIP = '/charli-extension.zip';

const features = [
  {
    icon: MessageSquare,
    title: 'Talk to any page',
    description: 'Open the side panel and ask. Charli understands the page you are on and helps you act.',
  },
  {
    icon: ShieldCheck,
    title: 'Safe by design',
    description: 'The model proposes; the backend decides. Risky actions always ask you first — nothing happens silently.',
  },
  {
    icon: MousePointerClick,
    title: 'Works where you already are',
    description: 'No new app to learn. Charli lives in your browser, right next to whatever you are doing.',
  },
];

const steps = [
  { icon: Download, title: 'Download', text: 'Grab the extension and unzip it.' },
  { icon: Puzzle, title: 'Load it', text: 'Add it to your browser in developer mode.' },
  { icon: PanelRight, title: 'Ask Charli', text: 'Open the side panel and start chatting.' },
];

export default function Home() {
  return (
    <div className="min-h-screen">
      {/* Nav */}
      <header className="sticky top-0 z-10 border-b bg-background/80 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2 font-semibold">
            <span className="flex size-7 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <Sparkles className="size-4" />
            </span>
            Charli
          </div>
          <nav className="flex items-center gap-4 text-sm">
            <a href="#how" className="text-muted-foreground hover:text-foreground">
              How it works
            </a>
            <Button size="sm" asChild>
              <a href={EXTENSION_ZIP} download>
                <Download className="size-4" /> Download
              </a>
            </Button>
          </nav>
        </div>
      </header>

      {/* Hero */}
      <section className="mx-auto grid max-w-5xl items-center gap-12 px-6 py-20 md:grid-cols-2">
        <div className="flex flex-col items-start gap-6">
          <span className="rounded-full border px-3 py-1 text-xs font-medium text-muted-foreground">
            Browser agent · v1
          </span>
          <h1 className="text-4xl font-semibold leading-tight tracking-tight sm:text-5xl">
            The agent that works <span className="text-primary">inside your browser.</span>
          </h1>
          <p className="max-w-md text-lg text-muted-foreground">
            Charli is a lightweight browser extension. Ask it anything about the page you are on and
            it helps you get things done — right where you already work.
          </p>
          <div className="flex flex-wrap gap-3">
            <Button size="lg" asChild>
              <a href={EXTENSION_ZIP} download>
                <Download className="size-4" /> Download for Chrome
              </a>
            </Button>
            <Button size="lg" variant="secondary" asChild>
              <a href="#how">See how it works</a>
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">Free · developer preview</p>
        </div>

        {/* Mock side panel — shows what the extension looks like */}
        <div className="mx-auto w-full max-w-xs">
          <div className="overflow-hidden rounded-2xl border bg-card shadow-xl">
            <div className="flex items-center gap-2 border-b px-4 py-3">
              <span className="flex size-6 items-center justify-center rounded-md bg-primary text-primary-foreground">
                <Sparkles className="size-3.5" />
              </span>
              <span className="text-sm font-semibold">Charli</span>
            </div>
            <div className="flex flex-col gap-3 p-4 text-sm">
              <div className="max-w-[80%] self-end rounded-lg bg-primary px-3 py-2 text-primary-foreground">
                What is this page about?
              </div>
              <div className="max-w-[85%] self-start rounded-lg bg-secondary px-3 py-2">
                It’s the Charli landing page — it explains the extension and lets you download it.
              </div>
              <div className="max-w-[80%] self-end rounded-lg bg-primary px-3 py-2 text-primary-foreground">
                Nice. Summarize it in one line.
              </div>
            </div>
            <div className="flex items-center gap-2 border-t p-3">
              <div className="flex-1 rounded-lg border px-3 py-2 text-sm text-muted-foreground">
                Ask Charli…
              </div>
              <span className="flex size-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <MessageSquare className="size-4" />
              </span>
            </div>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="mx-auto max-w-5xl px-6 pb-20">
        <div className="grid gap-6 sm:grid-cols-3">
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
        </div>
      </section>

      {/* How it works */}
      <section id="how" className="border-t bg-secondary/30">
        <div className="mx-auto max-w-5xl px-6 py-20">
          <h2 className="text-2xl font-semibold tracking-tight">How it works</h2>
          <p className="mt-2 text-muted-foreground">Three steps to get Charli in your browser.</p>
          <div className="mt-8 grid gap-6 sm:grid-cols-3">
            {steps.map(({ icon: Icon, title, text }, i) => (
              <div key={title} className="flex flex-col gap-3 rounded-xl border bg-card p-6">
                <div className="flex items-center gap-3">
                  <span className="flex size-8 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">
                    {i + 1}
                  </span>
                  <Icon className="size-4 text-muted-foreground" />
                </div>
                <h3 className="font-medium">{title}</h3>
                <p className="text-sm text-muted-foreground">{text}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Download / install */}
      <section className="mx-auto max-w-5xl px-6 py-20">
        <div className="rounded-2xl border bg-card p-8 md:p-12">
          <h2 className="text-2xl font-semibold tracking-tight">Get Charli</h2>
          <p className="mt-2 max-w-lg text-muted-foreground">
            Developer preview — installs in about a minute. A one-click Chrome Web Store version is
            coming later.
          </p>
          <div className="mt-6">
            <Button size="lg" asChild>
              <a href={EXTENSION_ZIP} download>
                <Download className="size-4" /> Download extension (.zip)
              </a>
            </Button>
          </div>
          <ol className="mt-8 grid gap-3 text-sm text-muted-foreground md:grid-cols-2">
            <li>1. Download and unzip the file above.</li>
            <li>
              2. Open <code className="rounded bg-secondary px-1.5 py-0.5">chrome://extensions</code>.
            </li>
            <li>3. Turn on “Developer mode” (top right).</li>
            <li>4. Click “Load unpacked” and pick the unzipped folder.</li>
          </ol>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-8 text-sm text-muted-foreground">
          <span>© Charli</span>
          <span>Your flexible browser agent.</span>
        </div>
      </footer>
    </div>
  );
}
