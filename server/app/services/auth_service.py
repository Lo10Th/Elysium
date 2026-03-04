"""Service layer for authentication business logic.

All auth-related DB operations and rules live here.
Routes call these methods and handle HTTP concerns only.

Services accept the Supabase client as a parameter so that
route-level mocks continue to work in tests without conftest changes.

Device-code constants and helpers are also centralised here so that
the route module only deals with HTTP-level concerns.
"""

import logging
import secrets
import string
import time
from datetime import datetime, timezone
from typing import Optional

from fastapi import HTTPException
from fastapi.responses import RedirectResponse
from supabase import Client

from app.models import User

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Device-code configuration
# ---------------------------------------------------------------------------

_DEVICE_CODE_LENGTH = 40
_USER_CODE_LENGTH = 8
_DEVICE_CODE_EXPIRY = 600  # seconds (10 minutes)
_POLL_INTERVAL = 5  # seconds


# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------


def _get_username(supabase: Client, user_id: str) -> Optional[str]:
    """Return the username for *user_id* from the profiles table, or None."""
    try:
        profile = (
            supabase.table("profiles")
            .select("username")
            .eq("id", user_id)
            .maybe_single()
            .execute()
        )
        return profile.data.get("username") if profile.data else None
    except Exception:
        return None


def _parse_expires(expires_at) -> Optional[datetime]:
    """Normalise a DB timestamp (str or datetime) to an aware datetime."""
    if not expires_at:
        return None
    if isinstance(expires_at, str):
        return datetime.fromisoformat(expires_at.replace("Z", "+00:00"))
    return expires_at


# ---------------------------------------------------------------------------
# User-facing response types (imported by the route module)
# ---------------------------------------------------------------------------

# Re-exported so route module only needs to import from auth_service.
DEVICE_CODE_EXPIRY = _DEVICE_CODE_EXPIRY
POLL_INTERVAL = _POLL_INTERVAL


# ---------------------------------------------------------------------------
# Public service methods
# ---------------------------------------------------------------------------


class AuthService:
    """Business logic for authentication and user-profile management."""

    # ------------------------------------------------------------------
    # Core authentication
    # ------------------------------------------------------------------

    @staticmethod
    def get_user_from_token(supabase: Client, token: str) -> User:
        """Validate *token* and return the corresponding User.

        Raises HTTPException 401 if the token is invalid.
        """
        try:
            response = supabase.auth.get_user(token)
            if not response.user:
                raise HTTPException(status_code=401, detail="Invalid token")

            profile = (
                supabase.table("profiles")
                .select("username")
                .eq("id", response.user.id)
                .single()
                .execute()
            )
            username = profile.data.get("username") if profile.data else None

            return User(
                id=response.user.id,
                email=response.user.email or "",
                username=username,
            )
        except HTTPException:
            raise
        except Exception:
            raise HTTPException(status_code=401, detail="Invalid token")

    @staticmethod
    def get_user_from_token_optional(supabase: Client, token: str) -> Optional[User]:
        """Like ``get_user_from_token`` but returns None instead of raising."""
        try:
            response = supabase.auth.get_user(token)
            if not response.user:
                return None

            profile = (
                supabase.table("profiles")
                .select("username")
                .eq("id", response.user.id)
                .maybe_single()
                .execute()
            )
            username = profile.data.get("username") if profile.data else None

            return User(
                id=response.user.id,
                email=response.user.email or "",
                username=username,
            )
        except Exception:
            return None

    @staticmethod
    def register(
        supabase: Client,
        email: str,
        password: str,
        username: str,
        frontend_url: str,
    ) -> dict:
        """Register a new user.  Returns a dict matching AuthResponse schema."""
        try:
            existing = (
                supabase.table("profiles")
                .select("id")
                .eq("username", username)
                .maybe_single()
                .execute()
            )
            if existing.data:
                raise HTTPException(status_code=400, detail="Username already taken")

            response = supabase.auth.sign_up(
                {
                    "email": email,
                    "password": password,
                    "options": {
                        "data": {"username": username},
                        "email_redirect_to": f"{frontend_url}/auth/callback",
                    },
                }
            )

            if not response.user:
                raise HTTPException(status_code=400, detail="Registration failed")

            return {
                "access_token": response.session.access_token
                if response.session
                else "",
                "refresh_token": response.session.refresh_token
                if response.session
                else "",
                "token_type": "bearer",
                "user": User(
                    id=response.user.id,
                    email=response.user.email or "",
                    username=username,
                ),
            }
        except HTTPException:
            raise
        except Exception as exc:
            error_msg = str(exc)
            if "already registered" in error_msg.lower():
                raise HTTPException(
                    status_code=400, detail="Email already registered"
                )
            logger.error("Registration failed for email='%s': %s", email, exc)
            raise HTTPException(status_code=400, detail="Registration failed")

    @staticmethod
    def login(supabase: Client, email: str, password: str) -> dict:
        """Authenticate with email/password.  Returns a dict matching AuthResponse."""
        try:
            response = supabase.auth.sign_in_with_password(
                {"email": email, "password": password}
            )
            if not response.user or not response.session:
                raise HTTPException(status_code=401, detail="Invalid credentials")

            username = _get_username(supabase, response.user.id)

            return {
                "access_token": response.session.access_token,
                "refresh_token": response.session.refresh_token,
                "token_type": "bearer",
                "user": User(
                    id=response.user.id,
                    email=response.user.email or "",
                    username=username,
                ),
            }
        except HTTPException:
            raise
        except Exception:
            raise HTTPException(status_code=401, detail="Invalid credentials")

    @staticmethod
    def logout(supabase: Client) -> dict:
        """Sign the current session out."""
        try:
            supabase.auth.sign_out()
            return {"message": "Logged out successfully"}
        except Exception as exc:
            logger.error("Logout error: %s", exc)
            raise HTTPException(status_code=400, detail="Logout failed")

    @staticmethod
    def refresh_token(supabase: Client, refresh_token: str) -> dict:
        """Exchange a refresh token for a fresh session."""
        try:
            response = supabase.auth.refresh_session(refresh_token)
            if not response.session:
                raise HTTPException(
                    status_code=401, detail="Invalid refresh token"
                )

            username: Optional[str] = None
            if response.user:
                username = _get_username(supabase, response.user.id)

            return {
                "access_token": response.session.access_token,
                "refresh_token": response.session.refresh_token,
                "token_type": "bearer",
                "user": User(
                    id=response.user.id if response.user else "",
                    email=response.user.email if response.user else "",
                    username=username,
                ),
            }
        except HTTPException:
            raise
        except Exception:
            raise HTTPException(status_code=401, detail="Invalid refresh token")

    @staticmethod
    def forgot_password(supabase: Client, email: str, frontend_url: str) -> dict:
        """Trigger a password-reset email.  Always returns the same message."""
        _SAFE_MSG = (
            "If an account with that email exists, "
            "we've sent password reset instructions."
        )
        try:
            supabase.auth.reset_password_for_email(
                email,
                options={"redirect_to": f"{frontend_url}/reset-password"},
            )
        except Exception:
            pass  # Intentional: never reveal whether the email exists.
        return {"message": _SAFE_MSG}

    @staticmethod
    def reset_password(supabase: Client, token: str, password: str) -> dict:
        """Verify a reset token and update the user's password."""
        try:
            supabase.auth.verify_oauth_token(
                {"type": "recovery", "token": token}
            )
            response = supabase.auth.update_user({"password": password})
            if not response.user:
                raise HTTPException(status_code=400, detail="Password reset failed")
            return {"message": "Password reset successfully"}
        except HTTPException:
            raise
        except Exception:
            raise HTTPException(
                status_code=400, detail="Invalid or expired reset token"
            )

    # ------------------------------------------------------------------
    # Profile
    # ------------------------------------------------------------------

    @staticmethod
    def get_profile(supabase: Client, user_id: str, email: str) -> dict:
        """Return full profile data for *user_id*."""
        try:
            profile = (
                supabase.table("profiles")
                .select("*")
                .eq("id", user_id)
                .single()
                .execute()
            )
            if not profile.data:
                raise HTTPException(status_code=404, detail="Profile not found")

            return {
                "id": user_id,
                "email": email,
                "username": profile.data.get("username"),
                "bio": profile.data.get("bio"),
                "avatar_url": profile.data.get("avatar_url"),
                "created_at": profile.data.get("created_at"),
                "updated_at": profile.data.get("updated_at"),
            }
        except HTTPException:
            raise
        except Exception as exc:
            logger.error("Failed to get profile for '%s': %s", user_id, exc)
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def update_profile(
        supabase: Client,
        user_id: str,
        email: str,
        username: Optional[str],
        bio: Optional[str],
        avatar_url: Optional[str],
    ) -> dict:
        """Update profile fields and return the updated profile."""
        try:
            update_data: dict = {}

            if username is not None:
                existing = (
                    supabase.table("profiles")
                    .select("id")
                    .eq("username", username)
                    .neq("id", user_id)
                    .maybe_single()
                    .execute()
                )
                if existing.data:
                    raise HTTPException(
                        status_code=400, detail="Username already taken"
                    )
                update_data["username"] = username

            if bio is not None:
                update_data["bio"] = bio
            if avatar_url is not None:
                update_data["avatar_url"] = avatar_url

            if update_data:
                update_data["updated_at"] = "now()"
                profile = (
                    supabase.table("profiles")
                    .update(update_data)
                    .eq("id", user_id)
                    .execute()
                )
            else:
                profile = (
                    supabase.table("profiles")
                    .select("*")
                    .eq("id", user_id)
                    .single()
                    .execute()
                )

            if not profile.data:
                raise HTTPException(status_code=404, detail="Profile not found")

            profile_data = (
                profile.data[0]
                if isinstance(profile.data, list)
                else profile.data
            )

            return {
                "id": user_id,
                "email": email,
                "username": profile_data.get("username"),
                "bio": profile_data.get("bio"),
                "avatar_url": profile_data.get("avatar_url"),
                "created_at": profile_data.get("created_at"),
                "updated_at": profile_data.get("updated_at"),
            }
        except HTTPException:
            raise
        except Exception as exc:
            logger.error("Failed to update profile for '%s': %s", user_id, exc)
            raise HTTPException(status_code=500, detail="Internal server error")

    # ------------------------------------------------------------------
    # OAuth
    # ------------------------------------------------------------------

    @staticmethod
    def oauth_start(
        supabase: Client,
        provider: str,
        redirect_uri: str,
        frontend_url: str,
        oauth_states: dict,
    ) -> dict:
        """Initiate an OAuth flow and return the provider's redirect URL."""
        try:
            state = secrets.token_urlsafe(32)
            oauth_states[state] = redirect_uri

            response = supabase.auth.sign_in_with_oauth(
                {
                    "provider": provider,
                    "options": {
                        "redirect_to": (
                            f"{frontend_url}/auth/callback?state={state}"
                        ),
                        "scopes": (
                            "user:email" if provider == "github" else "email profile"
                        ),
                    },
                }
            )
            return {"url": response.url}
        except Exception as exc:
            logger.error("OAuth start failed for '%s': %s", provider, exc)
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def oauth_callback(
        supabase: Client,
        provider: str,
        code: str,
        state: str,
        error: str,
        frontend_url: str,
        oauth_states: dict,
    ) -> RedirectResponse:
        """Handle the OAuth provider callback and redirect the browser."""
        if error:
            redirect_uri = oauth_states.pop(state, frontend_url)
            return RedirectResponse(url=f"{redirect_uri}?error={error}")

        if not code or not state:
            raise HTTPException(
                status_code=400, detail="Missing code or state parameter"
            )

        redirect_uri = oauth_states.pop(state, frontend_url)

        try:
            response = supabase.auth.exchange_code_for_session({"auth_code": code})
            if not response.session:
                raise HTTPException(
                    status_code=401, detail="OAuth authentication failed"
                )

            return RedirectResponse(
                url=(
                    f"{redirect_uri}"
                    f"?access_token={response.session.access_token}"
                    f"&refresh_token={response.session.refresh_token}"
                )
            )
        except HTTPException:
            raise
        except Exception as exc:
            logger.error("OAuth callback failed for '%s': %s", provider, exc)
            return RedirectResponse(url=f"{redirect_uri}?error=oauth_failed")

    # ------------------------------------------------------------------
    # Device-code flow
    # ------------------------------------------------------------------

    @staticmethod
    def generate_device_code() -> str:
        """Generate a cryptographically random device code."""
        chars = string.ascii_letters + string.digits
        return "".join(secrets.choice(chars) for _ in range(_DEVICE_CODE_LENGTH))

    @staticmethod
    def generate_user_code() -> str:
        """Generate a human-friendly user code (e.g. ``ABCD-1234``)."""
        chars = string.ascii_uppercase + string.digits
        code = "".join(secrets.choice(chars) for _ in range(_USER_CODE_LENGTH))
        return f"{code[:4]}-{code[4:]}"

    @staticmethod
    def create_device_code(
        supabase: Client,
        client_name: str,
        frontend_url: str,
    ) -> dict:
        """Insert a new device-code record and return the response payload."""
        device_code = AuthService.generate_device_code()
        user_code = AuthService.generate_user_code()

        supabase.table("device_codes").insert(
            {
                "device_code": device_code,
                "user_code": user_code,
                "client_name": client_name,
                "expires_at": (
                    f"now() + interval '{_DEVICE_CODE_EXPIRY} seconds'"
                ),
            }
        ).execute()

        return {
            "device_code": device_code,
            "user_code": user_code,
            "verification_uri": f"{frontend_url}/device",
            "expires_in": _DEVICE_CODE_EXPIRY,
            "interval": _POLL_INTERVAL,
        }

    @staticmethod
    def get_device_status(supabase: Client, user_code: str) -> dict:
        """Return current status for a device-code authorisation."""
        result = (
            supabase.table("device_codes")
            .select("*")
            .eq("user_code", user_code.upper())
            .single()
            .execute()
        )
        if not result.data:
            raise HTTPException(status_code=404, detail="Device code not found")

        row = result.data
        return {
            "user_code": row["user_code"],
            "verified": row.get("verified_at") is not None,
            "client_name": row.get("client_name", "Elysium CLI"),
            "expires_at": row["expires_at"],
        }

    @staticmethod
    def verify_device_code(
        supabase: Client,
        user_code: str,
        user: User,
    ) -> dict:
        """Mark a device code as verified by the authenticated user."""
        result = (
            supabase.table("device_codes")
            .select("*")
            .eq("user_code", user_code.upper())
            .single()
            .execute()
        )
        if not result.data:
            raise HTTPException(status_code=404, detail="Invalid user code")

        row = result.data

        if row.get("verified_at"):
            raise HTTPException(
                status_code=400, detail="Device code already verified"
            )

        expires_at = _parse_expires(row.get("expires_at"))
        if expires_at and expires_at < datetime.now(timezone.utc):
            raise HTTPException(status_code=400, detail="Device code has expired")

        # NOTE: This endpoint's sole responsibility is to mark the device code
        # as verified so that poll_device_token can later issue tokens.  The
        # session / magic-link retrieval below was present in the original
        # implementation but its result is intentionally unused here — the
        # actual token exchange happens in poll_device_token which generates
        # its own magic link at that point.  Retained for side-effect
        # compatibility with any Supabase auth session state.
        session_resp = supabase.auth.get_session()
        if not session_resp or not getattr(session_resp, "access_token", None):
            # Warm up admin magic-link generation; result consumed by
            # poll_device_token, not here.
            supabase.auth.admin.generate_link(
                {"type": "magiclink", "email": user.email}
            )

        supabase.table("device_codes").update(
            {"user_id": user.id, "verified_at": "now()"}
        ).eq("user_code", user_code.upper()).execute()

        return {
            "message": "Device authorized successfully",
            "user_code": user_code.upper(),
        }

    @staticmethod
    def poll_device_token(supabase: Client, device_code: str) -> dict:
        """Poll for a completed device-code authorisation and return tokens."""
        result = (
            supabase.table("device_codes")
            .select("*")
            .eq("device_code", device_code)
            .single()
            .execute()
        )
        if not result.data:
            raise HTTPException(status_code=404, detail="Invalid device code")

        row = result.data

        expires_at = _parse_expires(row.get("expires_at"))
        if expires_at and expires_at < datetime.now(timezone.utc):
            raise HTTPException(status_code=400, detail="Device code has expired")

        if not row.get("verified_at"):
            raise HTTPException(status_code=400, detail="Authorization pending")

        if not row.get("user_id"):
            raise HTTPException(status_code=400, detail="Authorization pending")

        user_id: str = row["user_id"]

        user_result = (
            supabase.table("profiles")
            .select("*")
            .eq("id", user_id)
            .single()
            .execute()
        )
        profile = user_result.data if user_result.data else {}

        magic_link = supabase.auth.admin.generate_link(
            {"type": "magiclink", "email": profile.get("email") or user_id}
        )

        if not magic_link or not hasattr(magic_link, "properties"):
            raise HTTPException(
                status_code=500, detail="Failed to generate session"
            )

        supabase.table("device_codes").delete().eq(
            "device_code", device_code
        ).execute()

        return {
            "access_token": magic_link.properties.get("access_token", ""),
            "refresh_token": magic_link.properties.get("refresh_token", ""),
            "token_type": "bearer",
            "user": User(
                id=user_id,
                email=profile.get("email", ""),
                username=profile.get("username"),
            ),
        }
