import { PaymentsClient } from "@payagentic.ai/sdk";

export async function buyData(url, propose) {
  const resource = new URL(url);
  if (resource.protocol !== "http:" || resource.hostname !== "127.0.0.1") {
    throw new Error("This simulation only calls a merchant on 127.0.0.1.");
  }
  const buyer = new PaymentsClient({
    apiKey: "local-fixture-not-a-real-key",
    walletId: "local-fixture-not-a-real-wallet",
    baseUrl: "https://gateway.invalid",
    proposer: propose,
  });
  const response = await buyer.fetch(url);
  if (!response.ok) throw new Error(`Merchant returned HTTP ${response.status}`);
  return response.json();
}
