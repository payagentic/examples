"""Isolation for demos/tests only; never import this in an application using CrewAI login."""

import os
from tempfile import TemporaryDirectory
from unittest.mock import patch


def prepare_offline_runtime():
    os.environ["OTEL_SDK_DISABLED"] = "true"
    os.environ["CREWAI_TELEMETRY_DISABLED"] = "true"
    os.environ["CREWAI_TRACING_ENABLED"] = "false"
    storage = TemporaryDirectory(prefix="payagentic-crewai-fixture-")
    os.environ["CREWAI_STORAGE_DIR"] = storage.name

    # CrewAI 1.15.23 checks its saved cloud login at import even with tracing off.
    # Simulate a logged-out user so the fixture never reads or creates user credentials.
    from crewai_core.auth.token import AuthError
    login = patch("crewai_core.auth.token.get_auth_token", side_effect=AuthError("Offline fixture"))
    login.start()

    def cleanup():
        login.stop()
        storage.cleanup()

    return cleanup
