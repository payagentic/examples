"""Synthetic HTTP fixture bound only to loopback; it makes no external requests."""

import json
import threading
from contextlib import contextmanager
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

from payagentic import PayAgentic
from payagentic.middleware.retry import RetryPolicy


@contextmanager
def fixture_client(status=200, malformed=False, has_more=False):
    requests = []

    class Handler(BaseHTTPRequestHandler):
        def log_message(self, *_args):
            pass

        def do_GET(self):
            parsed = urlparse(self.path)
            if parsed.path != "/v1/wallets":
                self.send_error(404)
                return
            requests.append({"method": "GET", "path": parsed.path,
                             "limit": parse_qs(parsed.query).get("limit")})
            wallet = {
                "id": "00000000-0000-7000-8000-000000000001",
                "organizationId": "00000000-0000-7000-8000-000000000002",
                "address": "synthetic-address", "balance": "12.34", "chain": "base",
                "currency": "USDC", "label": "synthetic-private-label", "status": "active",
                "type": "managed_account", "createdAt": "2026-01-01T00:00:00Z",
                "updatedAt": "2026-01-01T00:00:00Z",
            }
            if status != 200:
                body = {"status": status, "title": "synthetic-sensitive-error", "detail": wallet["label"]}
            elif malformed:
                body = {"unexpected": wallet}
            else:
                body = {"items": [wallet], "has_more": has_more}
            payload = json.dumps(body).encode()
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)

    server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    worker = threading.Thread(target=server.serve_forever, daemon=True)
    worker.start()
    try:
        with PayAgentic(
            api_key="synthetic-example-not-a-credential", agent_id="synthetic-agent",
            base_url=f"http://127.0.0.1:{server.server_port}",
            retry_policy=RetryPolicy(max_retries=0),
        ) as client:
            yield client, requests
    finally:
        server.shutdown()
        server.server_close()
        worker.join()
