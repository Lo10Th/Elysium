"""Service layer for Elysium server.

Business logic is separated from route handlers.
Each service module provides static methods that accept
the Supabase client as a parameter for testability.
"""

from app.services.auth_service import AuthService
from app.services.emblem_service import EmblemService
from app.services.key_service import KeyService

__all__ = ["AuthService", "EmblemService", "KeyService"]
