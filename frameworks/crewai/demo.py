from offline_runtime import prepare_offline_runtime

cleanup = prepare_offline_runtime()

from fixture import fixture_client
from wallet_tool import WalletPageSummary

try:
    with fixture_client() as (client, _requests):
        print(WalletPageSummary(client).run(limit=10))
finally:
    cleanup()
