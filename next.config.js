/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  // Expose NEXT_PUBLIC_API_URL at runtime
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
  },
  // Disable rewrites in production
  async rewrites() {
    return []
  },
  // Simplified headers - exclude static files from security headers
  async headers() {
    return [
      {
        source: '/:path((?!_next/static|_next/image|favicon.ico).*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block',
          },
        ],
      },
    ]
  },
}

module.exports = nextConfig
