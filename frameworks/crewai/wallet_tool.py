"""One native CrewAI tool, wired to the released PayAgentic Python SDK."""

from typing import Type

from crewai.tools import BaseTool
from payagentic import PayAgentic
from payagentic._generated.models.paginated_wallets import PaginatedWallets
from pydantic import BaseModel, ConfigDict, Field, PrivateAttr


class WalletPageInput(BaseModel):
    model_config = ConfigDict(extra="forbid", strict=True, hide_input_in_errors=True)
    limit: int = Field(default=10, ge=1, le=100, description="First-page size, not an account total")


class WalletPageSummary(BaseTool):
    name: str = "payagentic_wallet_page_summary"
    description: str = (
        "Count wallets on the first page and report whether more pages exist. Read-only. "
        "Does not return identifiers, addresses, balances or labels."
    )
    args_schema: Type[BaseModel] = WalletPageInput
    _client: PayAgentic = PrivateAttr()

    def __init__(self, client: PayAgentic):
        super().__init__()
        self._client = client

    def _run(self, limit: int = 10, **kwargs) -> dict:
        # Validate direct Python calls too, before touching the API.
        validated = WalletPageInput.model_validate({"limit": limit, **kwargs})
        try:
            page = self._client.wallets.list_wallets(limit=validated.limit)
            if not isinstance(page, PaginatedWallets) or type(page.has_more) is not bool:
                raise ValueError("Invalid response")
            return {"wallets_on_page": len(page.items), "has_more": page.has_more}
        except Exception:
            # Do not expose upstream response text, credentials or request objects.
            raise RuntimeError(
                "PayAgentic wallet summary unavailable. Check access privately and retry."
            ) from None
