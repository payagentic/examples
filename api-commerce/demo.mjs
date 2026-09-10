import assert from "node:assert/strict";
import { buyData } from "./buyer.mjs";
import { createSimulation, startMerchant } from "./merchant.mjs";

const simulation = createSimulation();
const merchant = await startMerchant(simulation);
try {
  const unpaid = await fetch(merchant.url);
  assert.equal(unpaid.status, 402);
  console.log("Merchant: HTTP 402; price 0.01 USDC (simulation).");
  const data = await buyData(merchant.url, simulation.propose);
  assert.equal(data.payment, "simulated");
  console.log("Buyer: HTTP 200; received Example API data.");
  console.log("No funds moved. No gateway, wallet, or blockchain was contacted.");
} finally {
  await merchant.close();
}
