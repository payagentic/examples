import { tool as langchainTool } from '@langchain/core/tools';
import { tool as vercelTool } from 'ai';
import { z } from 'zod';

export const inputSchema = z.object({
  limit: z.number().int().min(1).max(100).default(10)
    .describe('Maximum wallets to inspect on the first page; not an account-wide total.'),
}).strict();

const description = 'Return the number of wallets on the first page and whether more pages exist. '
  + 'Read-only. Does not return wallet identifiers, addresses, balances or labels.';
export const toolName = 'payagentic_wallet_page_summary';

// Project only explicitly allowed fields before returning anything to a model.
// Never return SDK response objects: they can contain requests, headers and account data.
export function makeSummary(client) {
  return async (input) => {
    const { limit } = inputSchema.parse(input);
    try {
      const result = await client.wallets.listWallets({ query: { limit } });
      if (result.error || !result.response?.ok || !Array.isArray(result.data?.items)
        || typeof result.data.has_more !== 'boolean') throw new Error('Invalid response');
      return { wallets_on_page: result.data.items.length, has_more: result.data.has_more };
    } catch {
      // Transport error text may contain upstream data; do not pass it to a model or console.
      throw new Error('PayAgentic wallet summary unavailable. Check access privately and retry.');
    }
  };
}

export function createLangChainTool(client) {
  return langchainTool(makeSummary(client), { name: toolName, description, schema: inputSchema });
}

export function createVercelTools(client) {
  return { [toolName]: vercelTool({ description, inputSchema, execute: makeSummary(client) }) };
}
