# PayAgentic wallet page summary for n8n

Two importable workflows use only built-in n8n nodes. They return the number of wallets on the first page, whether more pages exist, and whether the response shape is valid. They never initiate payments or approvals.

## Try the synthetic workflow

Import `synthetic.json` into an isolated n8n workspace, then run **Manual start**. It uses one synthetic wallet and makes no API requests. Expected final output:

```json
{"wallets_on_page":1,"has_more":false,"valid_response":true}
```

The intermediate wallet contains deliberately synthetic private fields so you can verify that **Wallet page summary** drops them. A malformed response returns null counts and `valid_response: false`, rather than reporting an incorrect zero balance or wallet count.

## Connect an account

`account-read.json` is a separate, inactive template. Its only HTTP request is `GET https://app.payagentic.ai/v1/wallets?limit=10`. Live authenticated acceptance remains required before using it with an account.

1. Import it into a private workspace on an approved n8n instance.
2. Create an **HTTP Header Auth** credential in n8n's credential store, with header name `Authorization` and value `Bearer <your own scoped API key>`. Use a disposable account for acceptance. Do not enter the key in a workflow node, JSON export, prompt, issue or screenshot.
3. Select that credential in **Read wallet page**. No credential ID, name or secret is embedded in the template.
4. Run manually and inspect the final summary privately. `has_more: true` means the count is incomplete; this workflow does not paginate or return balances.

The request has a 10-second timeout, follows no redirects and does not continue after an HTTP failure. The workflow has a 30-second execution timeout. No schedule, webhook or external notification is configured.

## Account-data handling

Successful, failed and manual execution persistence is disabled in both workflow settings, and there is no pinned execution data. Review those settings after import because instance-level configuration may affect retention. The intermediate HTTP response is still processed in n8n memory and can be visible in its editor while running. Only the final node removes wallet details. Use an approved private instance, restrict workspace access, and do not share live execution screenshots or raw error output.

The n8n 2.41.7 server CLI saved execution records during our synthetic tests despite these workflow settings. Use the CLI only with fixtures unless you have separately verified and approved its retention behavior. Do not interpret the template's settings as a guarantee that the instance stores no account data.

The JSON files are public templates, not an accepted n8n marketplace listing or a verified community node. They do not demonstrate production OAuth. Keep a live acceptance record privately before submitting the account-connected template to a directory.

## Validation

Run `node automation/n8n/validate.mjs` from the repository root for structural and projection checks. These check the read-only route, credential omission, redirect policy, retention settings and final field filtering. Both workflows passed runtime validation in n8n 2.41.7 on 5 October 2026. For the account template, only the URL was changed to a loopback fixture and a synthetic Header Auth credential was selected. The test verified one authenticated GET and the final filtered result. It did not use a live PayAgentic account.

References: [HTTP Request node](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.httprequest/), [HTTP Request credentials](https://docs.n8n.io/integrations/builtin/credentials/httprequest), [Edit Fields node](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.set).

## Directory submission materials

See [SUBMISSION.md](SUBMISSION.md) for the two listing descriptions, creator workflow and submission gates. Each JSON includes a yellow setup note.
