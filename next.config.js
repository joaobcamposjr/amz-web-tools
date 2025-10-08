/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  // Disable all headers for static files
  async headers() {
    return []
  },
  // Disable rewrites
  async rewrites() {
    return []
  },
}

module.exports = nextConfig
