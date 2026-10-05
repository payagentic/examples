# Summarize PayAgentic wallets with HTTP Request and Edit Fields

## Who is this for?
Developers and operations teams who need a compact, read-only summary of their first PayAgentic wallet page.

## What it does
A manual trigger makes one authenticated GET request for up to ten wallets. Edit Fields returns only `wallets_on_page`, `has_more`, and `valid_response`. It excludes wallet identifiers, addresses, labels, and balances from the final output. It does not paginate; the count is not an account total. Malformed responses produce null counts and an invalid-response flag. HTTP failures stop execution.

## Setup and requirements
Import `account-read.json` into an approved private n8n instance. Create an HTTP Header Auth credential in the credential store: header `Authorization`, value `Bearer <your own scoped API key>`. Select it in **Read wallet page**, then run manually with a disposable test account. Never put a key in node parameters, exports, screenshots, or chat.

## Customize
Add a destination only after reviewing which summary fields it receives. Confirm instance retention first: intermediate responses can appear in the editor, and CLI execution may persist them despite workflow settings.

Only built-in nodes are used. Synthetic loopback execution passed with n8n 2.41.7. Live PayAgentic acceptance remains pending. This workflow cannot initiate payments, approvals, funding, or policy changes.
