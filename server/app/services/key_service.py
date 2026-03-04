"""Service layer for API key business logic.

All DB operations and key-management rules live here.
Routes call these methods and handle HTTP concerns only.

Services accept the Supabase client as a parameter so that
route-level mocks continue to work in tests without conftest changes.
"""

import hashlib
import logging
import secrets
from datetime import datetime, timedelta
from typing import List, Optional

from fastapi import HTTPException
from supabase import Client

from app.models import KeyCreate, KeyListItem, KeyResponse

logger = logging.getLogger(__name__)


def _generate_api_key() -> str:
    """Generate a new random API key with the ``ely_`` prefix."""
    return f"ely_{secrets.token_urlsafe(32)}"


def _hash_key(key: str) -> str:
    """Return the SHA-256 hex digest of *key*."""
    return hashlib.sha256(key.encode()).hexdigest()


def _parse_dt(value: Optional[str]) -> Optional[datetime]:
    """Convert an ISO-8601 string (possibly ending with Z) to a datetime."""
    if not value:
        return None
    return datetime.fromisoformat(value.replace("Z", "+00:00"))


class KeyService:
    """Business logic for API key management."""

    @staticmethod
    def list_keys(supabase: Client, user_id: str) -> List[KeyListItem]:
        """Return all API keys (without secret values) for the given user."""
        try:
            response = (
                supabase.table("api_keys")
                .select("id, name, created_at, expires_at")
                .eq("user_id", user_id)
                .execute()
            )
            return [
                KeyListItem(
                    id=row["id"],
                    name=row["name"],
                    created_at=_parse_dt(row["created_at"]),
                    expires_at=_parse_dt(row.get("expires_at")),
                )
                for row in response.data
            ]
        except Exception as exc:
            logger.error("Failed to list keys for user '%s': %s", user_id, exc)
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def create_key(
        supabase: Client,
        user_id: str,
        request_body: KeyCreate,
    ) -> KeyResponse:
        """Create a new API key and return it (raw value shown once only)."""
        existing = (
            supabase.table("api_keys")
            .select("id")
            .eq("user_id", user_id)
            .eq("name", request_body.name)
            .execute()
        )
        if existing.data:
            raise HTTPException(
                status_code=400, detail="A key with this name already exists"
            )

        raw_key = _generate_api_key()
        key_hash = _hash_key(raw_key)

        expires_at: Optional[str] = None
        if request_body.expires_days:
            from datetime import timezone as _tz
            expires_at = (
                datetime.now(_tz.utc) + timedelta(days=request_body.expires_days)
            ).isoformat()

        try:
            response = (
                supabase.table("api_keys")
                .insert(
                    {
                        "user_id": user_id,
                        "name": request_body.name,
                        "key_hash": key_hash,
                        "expires_at": expires_at,
                    }
                )
                .execute()
            )
            if not response.data:
                raise HTTPException(status_code=500, detail="Failed to create key")

            row = response.data[0]
            return KeyResponse(
                id=row["id"],
                name=request_body.name,
                key=raw_key,
                created_at=_parse_dt(row["created_at"]),
                expires_at=_parse_dt(row.get("expires_at")),
            )
        except HTTPException:
            raise
        except Exception as exc:
            logger.error(
                "Failed to create key '%s' for user '%s': %s",
                request_body.name,
                user_id,
                exc,
            )
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def get_key(supabase: Client, user_id: str, key_id: str) -> KeyListItem:
        """Return key metadata (no secret value) for the given user and key ID."""
        try:
            response = (
                supabase.table("api_keys")
                .select("id, name, created_at, expires_at")
                .eq("id", key_id)
                .eq("user_id", user_id)
                .single()
                .execute()
            )
            if not response.data:
                raise HTTPException(status_code=404, detail="Key not found")

            row = response.data
            return KeyListItem(
                id=row["id"],
                name=row["name"],
                created_at=_parse_dt(row["created_at"]),
                expires_at=_parse_dt(row.get("expires_at")),
            )
        except HTTPException:
            raise
        except Exception as exc:
            logger.error(
                "Failed to get key '%s' for user '%s': %s", key_id, user_id, exc
            )
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def delete_key(supabase: Client, user_id: str, key_id: str) -> None:
        """Delete the specified API key if it belongs to the given user."""
        try:
            response = (
                supabase.table("api_keys")
                .select("id")
                .eq("id", key_id)
                .eq("user_id", user_id)
                .single()
                .execute()
            )
            if not response.data:
                raise HTTPException(status_code=404, detail="Key not found")

            supabase.table("api_keys").delete().eq("id", key_id).eq(
                "user_id", user_id
            ).execute()
        except HTTPException:
            raise
        except Exception as exc:
            logger.error(
                "Failed to delete key '%s' for user '%s': %s", key_id, user_id, exc
            )
            raise HTTPException(status_code=500, detail="Internal server error")
