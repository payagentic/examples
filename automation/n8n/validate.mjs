import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import vm from 'node:vm';

const workflows = ['synthetic', 'account-read'].map(name => JSON.parse(readFileSync(new URL(`${name}.json`, import.meta.url))));
for (const workflow of workflows) {
  assert.equal(workflow.active, false);
  assert.deepEqual(workflow.pinData, {});
  assert.equal(workflow.settings.saveDataSuccessExecution, 'none');
  assert.equal(workflow.settings.saveDataErrorExecution, 'none');
  assert.equal(workflow.settings.saveManualExecutions, false);
  assert.equal(workflow.settings.saveExecutionProgress, false);
  assert.equal(workflow.settings.executionTimeout, 30);
  assert.equal(workflow.nodes[0].type, 'n8n-nodes-base.manualTrigger');
  assert.ok(workflow.nodes.every(node => !node.credentials && !node.onError && !node.continueOnFail));
  const summary = workflow.nodes.find(node => node.name === 'Wallet page summary');
  assert.equal(summary.parameters.includeOtherFields, false);
  const expression = summary.parameters.jsonOutput.slice(3, -2);
  const project = $json => JSON.parse(vm.runInNewContext(expression, { $json }, { timeout: 100 }));
  assert.deepEqual(project({ items: [{ secret: 'synthetic-private-field' }], has_more: true }),
    { wallets_on_page: 1, has_more: true, valid_response: true });
  assert.deepEqual(project({ items: [], has_more: false }),
    { wallets_on_page: 0, has_more: false, valid_response: true });
  assert.deepEqual(project({ error: 'synthetic-error' }),
    { wallets_on_page: null, has_more: null, valid_response: false });
}
assert.ok(workflows[0].nodes.every(node => node.type !== 'n8n-nodes-base.httpRequest'));
const request = workflows[1].nodes.find(node => node.type === 'n8n-nodes-base.httpRequest');
assert.equal(request.parameters.method, 'GET');
assert.equal(request.parameters.url, 'https://app.payagentic.ai/v1/wallets');
assert.equal(request.parameters.genericAuthType, 'httpHeaderAuth');
assert.equal(request.parameters.options.redirect.redirect.followRedirects, false);
assert.equal(request.parameters.options.timeout, 10000);
assert.deepEqual(request.parameters.queryParameters.parameters, [{ name: 'limit', value: '10' }]);
console.log('Both templates pass structural and data-projection checks. This does not prove n8n runtime or live acceptance.');
