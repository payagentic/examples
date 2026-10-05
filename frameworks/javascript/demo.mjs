import { AIMessage } from '@langchain/core/messages';
import { ToolNode } from '@langchain/langgraph/prebuilt';
import { createLangChainTool, createVercelTools, toolName } from './tools.mjs';
import { createFixture } from './fixture.mjs';

process.env.LANGSMITH_TRACING = 'false';
process.env.LANGCHAIN_TRACING_V2 = 'false';

// No model provider, real credentials, external API requests or payments are used.
const { client } = createFixture();
const langchain = createLangChainTool(client);
console.log('LangChain:', await langchain.invoke({ limit: 10 }));

const graph = new ToolNode([langchain]);
const output = await graph.invoke({ messages: [new AIMessage({
  content: '', tool_calls: [{ name: toolName, args: { limit: 10 }, id: 'synthetic-call' }],
})] });
console.log('LangGraph:', output.messages[0].content);

const vercel = createVercelTools(client);
console.log('Vercel AI SDK:', await vercel[toolName].execute({ limit: 10 }));
