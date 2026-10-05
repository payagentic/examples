# n8n template submission package

Prepared 5 October 2026. Neither workflow is an accepted n8n directory listing.

Use the official [Creator Hub](https://creators.n8n.io/hub) and [Creator Dashboard](https://creators.n8n.io/login). An n8n Cloud login and a template creator profile are distinct. Unverified creators can submit one template at a time; wait for its approval before submitting another.

The [submission guidelines](https://n8n.notion.site/Template-submission-guidelines-9959894476734da3b402c90b124b1f77) require explanatory sticky notes and clear setup descriptions. Both JSON files now include the complete matching listing description in a yellow note. They use only built-in nodes and contain no credential references or pinned execution data.

| Candidate | Copy | Submission condition |
|---|---|---|
| Synthetic learning example | [synthetic-listing.md](synthetic-listing.md) | Can be reviewed independently of production. Educational value and creator eligibility are subject to n8n review. |
| Account-connected summary | [account-read-listing.md](account-read-listing.md) | Hold until scoped live-account acceptance and instance retention are verified. |

## Submission steps

1. Sign in with the authorized creator profile; check for an existing submission.
2. Import the selected JSON into an isolated n8n workspace. Read the yellow note and confirm node layout and title.
3. Execute the synthetic workflow first. Capture only synthetic output; do not capture credential editing screens.
4. Paste its matching Markdown description in the Creator Dashboard and use the built-in node tags shown by the portal. Do not claim a verified PayAgentic community node.
5. Review the preview and any terms shown. Save the actual submission ID and review state after sending; a local file or GitHub PR is not submission evidence.
6. For the account workflow, first record a successful read, a denied/revoked credential result, and no unintended requests using an isolated scoped account. Keep raw account responses private. Confirm that the approved instance handles execution data as intended.

## Media

The [documentation-preview image](synthetic-guide.jpg) describes the synthetic flow; its [HTML source](preview.html) is included. It is not an n8n editor screenshot or live-account proof. A native canvas screenshot can be captured during the import step above if requested by the reviewer. Built-in-node templates can also use the directory's workflow preview.

## Verification

Run `node automation/n8n/validate.mjs`. Runtime validation uses synthetic data only. See [README.md](README.md) for the retention limitation and read-only request details.
