import assert from "node:assert/strict";
import { test } from "node:test";
import { PolicyDeniedError } from "@payagentic.ai/sdk";
import { buyData } from "./buyer.mjs";
import { createSimulation, startMerchant } from "./merchant.mjs";

test("packaged buyer and merchant exchange data only after fixture verification", async () => {
  const simulation = createSimulation();
  const merchant = await startMerchant(simulation);
  try {
    const unpaid = await fetch(merchant.url);
    assert.equal(unpaid.status, 402);
    assert.equal((await unpaid.json()).amount, "0.01");
    const invalid = await fetch(merchant.url, { headers: { Authorization: "Payment invalid" } });
    assert.equal(invalid.status, 402);
    assert.deepEqual(await buyData(merchant.url, simulation.propose), {
      source: "local-fixture", report: "Example API data", payment: "simulated",
    });
    await assert.rejects(buyData(merchant.url, async () => { throw new PolicyDeniedError("Demo policy denial"); }), PolicyDeniedError);
    await assert.rejects(buyData("https://merchant.example/data", simulation.propose), /127.0.0.1/);
  } finally {
    await merchant.close();
  }
});

test("fixture credentials are rejected on replay", async () => {
  const simulation = createSimulation();
  const proposal = await simulation.propose({ amount: "0.01", currency: "USDC" });
  const credential = Buffer.from(JSON.stringify({ signature: proposal.payment_signature, intent_id: proposal.intent.id })).toString("base64");
  const init = { body: JSON.stringify({ credential, expected_amount: "0.01", expected_currency: "USDC" }) };
  assert.equal((await simulation.verify("", init)).status, 200);
  assert.equal((await simulation.verify("", init)).status, 403);
});
