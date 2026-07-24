import type { ReactNode } from 'react';
import { Inter } from 'next/font/google';
import { Toaster } from 'sonner';
import { Providers } from './providers';
import './globals.css';

// Same body font as levelaxis/cv's web app, self-hosted at build time by
// next/font (no runtime request, unlike the extension panel's fallback-only
// approach — that one avoids fetching a webfont to stay dependency-light).
const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });

export const metadata = {
  title: 'Charli',
  description: 'Charli — your flexible browser agent.',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className={inter.variable}>
      <body className="min-h-screen font-sans antialiased">
        <Providers>{children}</Providers>
        <Toaster richColors position="top-right" />
      </body>
    </html>
  );
}
