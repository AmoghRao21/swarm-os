/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/api/v1/:path*",
        destination: "http://swarm-core:8080/api/v1/:path*",
      },
    ];
  },
};

export default nextConfig;