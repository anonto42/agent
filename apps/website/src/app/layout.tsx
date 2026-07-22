import type { ReactNode } from 'react';
import { Toaster } from 'sonner';
import { Providers } from './providers';
import './globals.css';

export const metadata = {
  title: 'Charli',
  description: 'Charli — your flexible browser agent.',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen antialiased">
        <Providers>{children}</Providers>
        <Toaster richColors position="top-right" />
      </body>
    </html>
  );
}
