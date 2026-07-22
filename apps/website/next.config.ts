import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  // @charli/shared is a workspace TS package, transpiled by Next.
  transpilePackages: ['@charli/shared'],
};

export default nextConfig;
