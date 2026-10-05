# Summarize a synthetic wallet page with Edit Fields

## Who is this for?
Developers and operations teams learning how to reduce an API response to a small reporting summary before connecting a real account.

## What it does
Run manually to create one fictional wallet, then use Edit Fields to keep only `wallets_on_page`, `has_more`, and `valid_response`. No network request or account credential is used. The result is a first-page count, not a balance or a total across pages. Invalid response shapes return null counts and `valid_response: false`.

## Setup and requirements
Import `synthetic.json` into n8n and run **Manual start**. Only built-in nodes are required. Expected output: one wallet, no additional page, valid response. Runtime compatibility was verified with n8n 2.41.7 using synthetic data.

## Customize
Change the fictional fixture to an empty page or set `has_more` to true. Keep the final node's field filtering enabled. Never replace the fixture with customer records in a shared workspace or public export.

This learning example does not connect PayAgentic, execute payments, approve requests, or demonstrate production access. The separate account-connected workflow requires your own scoped credential and live acceptance. Execution retention depends on your n8n instance; review its settings before processing any account data.
