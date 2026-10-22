// Cloudflare Worker for the public Disaster Network demo.
// Deploy this in the disaster-network-proxy Worker, not in the Nuxt app.
const ORIGIN = 'https://01a16309-eb76-7c3d-88c1-dfbf49a2c465-8080.eur-1.aiven.app';

export default {
  async fetch(request) {
    const incoming = new URL(request.url);
    const upstream = new URL(incoming.pathname + incoming.search, ORIGIN);
    const headers = new Headers(request.headers);
    headers.set('Host', upstream.host);
    headers.set('X-Forwarded-Host', incoming.host);
    headers.set('X-Forwarded-Proto', incoming.protocol.slice(0, -1));

    const response = await fetch(new Request(upstream.toString(), {
      method: request.method,
      headers,
      body: ['GET', 'HEAD'].includes(request.method) ? undefined : request.body,
      redirect: 'manual',
    }));

    if (response.status === 101) return response;

    const outputHeaders = new Headers(response.headers);
    const location = outputHeaders.get('Location');
    if (location) {
      const target = new URL(location, upstream);
      if (target.origin === ORIGIN) {
        target.host = incoming.host;
        target.protocol = incoming.protocol;
        outputHeaders.set('Location', target.href);
      }
    }
    return new Response(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers: outputHeaders,
    });
  },
};
