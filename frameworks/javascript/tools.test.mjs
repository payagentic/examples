import assert from 'node:assert/strict';
import test from 'node:test';
import { AIMessage } from '@langchain/core/messages';
import { ToolNode } from '@langchain/langgraph/prebuilt';
import { generateText } from 'ai';
import { MockLanguageModelV4 } from 'ai/test';
import { createLangChainTool, createVercelTools, toolName } from './tools.mjs';
import { createFixture } from './fixture.mjs';

process.env.LANGSMITH_TRACING = 'false';
process.env.LANGCHAIN_TRACING_V2 = 'false';

const expected = { wallets_on_page: 1, has_more: false };

test('LangChain native tool reads through the released SDK and projects only summary fields', async () => {
  const { client, requests } = createFixture();
  assert.deepEqual(await createLangChainTool(client).invoke({ limit: 7 }), expected);
  assert.deepEqual(requests, [{ method: 'GET', path: '/v1/wallets', limit: '7' }]);
});

test('LangGraph ToolNode produces a successful ToolMessage with the same safe projection', async () => {
  const { client } = createFixture();
  const node = new ToolNode([createLangChainTool(client)]);
  const result = await node.invoke({ messages: [new AIMessage({ content: '', tool_calls: [
    { name: toolName, args: { limit: 3 }, id: 'synthetic-call' },
  ] })] });
  assert.equal(result.messages[0].tool_call_id, 'synthetic-call');
  assert.notEqual(result.messages[0].status, 'error');
  assert.deepEqual(JSON.parse(result.messages[0].content), expected);
});

test('Vercel tool has the current inputSchema and executes with explicit pagination semantics', async () => {
  const { client } = createFixture({ hasMore: true });
  const tools = createVercelTools(client);
  assert.deepEqual(Object.keys(tools), [toolName]);
  assert.ok(tools[toolName].inputSchema);
  assert.deepEqual(await tools[toolName].execute({}), { wallets_on_page: 1, has_more: true });
});

test('Vercel generateText dispatches a synthetic model tool call without a provider request', async () => {
  const { client, requests } = createFixture();
  const model = new MockLanguageModelV4({ doGenerate: {
    content: [{ type: 'tool-call', toolCallId: 'synthetic-call', toolName, input: '{"limit":4}' }],
    finishReason: { unified: 'tool-calls', raw: 'tool_calls' },
    usage: { inputTokens: { total: 0 }, outputTokens: { total: 0 } }, warnings: [],
  } });
  const result = await generateText({ model, tools: createVercelTools(client), prompt: 'Synthetic test' });
  assert.deepEqual(result.toolResults[0].output, expected);
  assert.equal(requests[0].limit, '4');
  assert.equal(model.doGenerateCalls.length, 1);
});

test('invalid limits and unknown fields fail before any SDK request in both direct wrappers', async () => {
  const { client, requests } = createFixture();
  const langchain = createLangChainTool(client);
  const tools = [langchain.invoke.bind(langchain),
    createVercelTools(client)[toolName].execute];
  for (const run of tools) {
    for (const input of [{ limit: 0 }, { limit: 101 }, { limit: 1.5 }, { limit: '10' }, { secret: 'value' }]) {
      await assert.rejects(() => run(input));
    }
  }
  assert.equal(requests.length, 0);
});

for (const options of [{ status: 401 }, { status: 503 }, { malformed: true }]) {
  test(`upstream failure is sanitized for both wrappers: ${JSON.stringify(options)}`, async () => {
    const { client } = createFixture(options);
    for (const run of [() => createLangChainTool(client).invoke({}),
      () => createVercelTools(client)[toolName].execute({})]) {
      await assert.rejects(run, { message: 'PayAgentic wallet summary unavailable. Check access privately and retry.' });
    }
  });
}
