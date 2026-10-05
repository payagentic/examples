import { PayagenticClient } from '@payagentic.ai/sdk';

// Synthetic fixture values only. This transport never delegates to network fetch.
export function createFixture({ status = 200, malformed = false, hasMore = false } = {}) {
  const requests = [];
  const client = new PayagenticClient({
    apiKey: 'synthetic-example-not-a-credential',
    agentId: 'synthetic-agent',
    baseUrl: 'https://payagentic.example',
    retryPolicy: { maxRetries: 0, baseDelay: 0, maxDelay: 0, jitter: false },
    fetch: async (request) => {
      const url = new URL(request.url);
      if (request.method !== 'GET' || url.origin !== 'https://payagentic.example'
        || url.pathname !== '/v1/wallets') throw new Error('Fixture only allows wallet reads');
      requests.push({ method: request.method, path: url.pathname, limit: url.searchParams.get('limit') });
      const wallet = {
        id: '00000000-0000-7000-8000-000000000001',
        organizationId: '00000000-0000-7000-8000-000000000002',
        address: 'synthetic-address', balance: '12.34', chain: 'base', currency: 'USDC',
        label: 'synthetic-private-label', status: 'active', type: 'managed_account',
        createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z',
      };
      const body = status !== 200 ? { status, title: 'synthetic-sensitive-error', detail: wallet.label }
        : malformed ? { unexpected: wallet } : { items: [wallet], has_more: hasMore };
      return Response.json(body, { status });
    },
  });
  return { client, requests };
}
