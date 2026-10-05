import socket
import unittest
from unittest.mock import patch
from offline_runtime import prepare_offline_runtime

unittest.addModuleCleanup(prepare_offline_runtime())

from fixture import fixture_client
from wallet_tool import WalletPageSummary


class ToolTests(unittest.TestCase):
    def setUp(self):
        original = socket.socket.connect

        def loopback_only(sock, address):
            if address[0] != "127.0.0.1":
                raise AssertionError("External network is forbidden in these tests")
            return original(sock, address)

        self.guard = patch.object(socket.socket, "connect", loopback_only)
        self.guard.start()
        self.addCleanup(self.guard.stop)

    def test_native_tool_and_released_sdk_project_only_summary(self):
        with fixture_client(has_more=True) as (client, requests):
            tool = WalletPageSummary(client)
            self.assertEqual(tool.run(limit=7), {"wallets_on_page": 1, "has_more": True})
            self.assertEqual(requests, [{"method": "GET", "path": "/v1/wallets", "limit": ["7"]}])
            self.assertNotIn("synthetic-example-not-a-credential", repr(tool))
            self.assertNotIn("_client", tool.model_dump())

    def test_invalid_input_never_reaches_sdk(self):
        with fixture_client() as (client, requests):
            tool = WalletPageSummary(client)
            for inputs in [{"limit": 0}, {"limit": 101}, {"limit": 1.5}, {"limit": "10"},
                           {"limit": True}, {"secret": "value"}]:
                with self.subTest(inputs=inputs), self.assertRaises(ValueError):
                    tool.run(**inputs)
            self.assertEqual(requests, [])

    def test_errors_do_not_expose_response_or_credentials(self):
        for options in [{"status": 401}, {"status": 503}, {"malformed": True}]:
            with self.subTest(options=options), fixture_client(**options) as (client, _requests):
                with self.assertRaisesRegex(RuntimeError, "^PayAgentic wallet summary unavailable. Check access privately and retry.$"):
                    WalletPageSummary(client).run()


if __name__ == "__main__":
    unittest.main()
