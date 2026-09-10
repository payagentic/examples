import { randomUUID } from "node:crypto";
import { payagentic } from "@payagentic.ai/vendor/express";
import express from "express";

// Teaching fixture only: this replaces payment signing and gateway verification.
// Never mount this verifier on a public server or use it to authorize real work.
export function createSimulation() {
  const issued = new Set();
  return {
    async propose(requirements) {
      if (requirements.amount !== "0.01" || requirements.currency !== "USDC") {
        throw new Error("The local demo only accepts its 0.01 USDC fixture.");
      }
      const intent = { id: randomUUID(), amount: "0.01" };
      const payment_signature = "SIMULATED-NOT-A-SIGNATURE";
      const credential = Buffer.from(
        JSON.stringify({ signature: payment_signature, intent_id: intent.id }),
      ).toString("base64");
      issued.add(credential);
      return { intent, payment_signature };
    },
    async verify(_input, init) {
      const body = JSON.parse(init.body);
      if (
        body.expected_amount !== "0.01" ||
        body.expected_currency !== "USDC" ||
        !issued.delete(body.credential)
      ) {
        return Response.json({ error: { message: "Invalid or reused demo credential" } }, { status: 403 });
      }
      return Response.json({
        payment_id: "simulation-only",
        agent_id: "local-demo",
        amount: "0.01",
        currency: "USDC",
        status: "simulated",
      });
    },
  };
}

export async function startMerchant(simulation) {
  const app = express();
  const pay = payagentic({
    apiKey: "local-fixture-not-a-real-key",
    baseUrl: "https://gateway.invalid",
    fetch: simulation.verify,
  });
  app.get("/data", pay({ price: "0.01", currency: "USDC" }), (_req, res) => {
    res.json({ source: "local-fixture", report: "Example API data", payment: "simulated" });
  });
  const server = await new Promise((resolve, reject) => {
    const listener = app.listen(0, "127.0.0.1", () => resolve(listener));
    listener.once("error", reject);
  });
  const address = server.address();
  if (!address || typeof address === "string") throw new Error("Missing local server address");
  return {
    url: `http://127.0.0.1:${address.port}/data`,
    close: () => new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve())),
  };
}
