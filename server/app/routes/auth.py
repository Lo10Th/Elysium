from fastapi import APIRouter, HTTPException, Depends, Request
from fastapi.responses import RedirectResponse
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel, EmailStr, field_validator
from app.database import get_supabase
from app.limiter import (
    limiter,
    PUBLIC_LIMIT,
    STRICT_LIMIT,
    REGISTER_LIMIT,
    REFRESH_LIMIT,
)
from app.models import User
from app.config import get_settings
import urllib.parse
import secrets
import re

router = APIRouter()
security = HTTPBearer()

FRONTEND_URL = get_settings().FRONTEND_URL


class LoginRequest(BaseModel):
    email: EmailStr
    password: str


class RegisterRequest(BaseModel):
    email: EmailStr
    password: str
    username: str

    @field_validator("password")
    @classmethod
    def password_min_length(cls, v: str) -> str:
        if len(v) < 8:
            raise ValueError("password must be at least 8 characters")
        return v

    @field_validator("username")
    @classmethod
    def username_format(cls, v: str) -> str:
        if not _USERNAME_RE.match(v):
            raise ValueError(
                "username must be 3-30 characters and contain only letters, digits, underscores, or hyphens"
            )
        return v


class TokenRefreshRequest(BaseModel):
    refresh_token: str


class ForgotPasswordRequest(BaseModel):
    email: EmailStr


class ResetPasswordRequest(BaseModel):
    token: str
    password: str

    @field_validator("password")
    @classmethod
    def password_min_length(cls, v: str) -> str:
        if len(v) < 8:
            raise ValueError("password must be at least 8 characters")
        return v


class UpdateProfileRequest(BaseModel):
    username: str | None = None
    bio: str | None = None
    avatar_url: str | None = None

    @field_validator("username")
    @classmethod
    def username_format(cls, v: str | None) -> str | None:
        if v and not _USERNAME_RE.match(v):
            raise ValueError(
                "username must be 3-30 characters and contain only letters, digits, underscores, or hyphens"
            )
        return v

    @field_validator("bio")
    @classmethod
    def bio_length(cls, v: str | None) -> str | None:
        if v and len(v) > 200:
            raise ValueError("bio must be at most 200 characters")
        return v


class ProfileResponse(BaseModel):
    id: str
    email: str
    username: str | None
    bio: str | None
    avatar_url: str | None
    created_at: str
    updated_at: str


class AuthResponse(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str = "bearer"
    user: User


class OAuthStartRequest(BaseModel):
    redirect_uri: str


async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
) -> User:
    try:
        supabase = get_supabase()
        token = credentials.credentials
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
            id=response.user.id, email=response.user.email or "", username=username
        )
    except HTTPException:
        raise
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid token")


async def get_current_user_optional(
    credentials: HTTPAuthorizationCredentials = Depends(security),
) -> User | None:
    try:
        supabase = get_supabase()
        token = credentials.credentials
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
            id=response.user.id, email=response.user.email or "", username=username
        )
    except:
        return None


@router.post("/register", response_model=AuthResponse)
@limiter.limit(REGISTER_LIMIT)
async def register(request: Request, request_body: RegisterRequest):
    try:
        supabase = get_supabase()

        existing = (
            supabase.table("profiles")
            .select("id")
            .eq("username", request_body.username)
            .maybe_single()
            .execute()
        )
        if existing.data:
            raise HTTPException(status_code=400, detail="Username already taken")

        response = supabase.auth.sign_up(
            {
                "email": request_body.email,
                "password": request_body.password,
                "options": {
                    "data": {"username": request_body.username},
                    "email_redirect_to": f"{FRONTEND_URL}/auth/callback",
                },
            }
        )

        if not response.user:
            raise HTTPException(status_code=400, detail="Registration failed")

        return AuthResponse(
            access_token=response.session.access_token if response.session else "",
            refresh_token=response.session.refresh_token if response.session else "",
            user=User(
                id=response.user.id,
                email=response.user.email or "",
                username=request_body.username,
            ),
        )
    except HTTPException:
        raise
    except Exception as e:
        error_msg = str(e)
        if "already registered" in error_msg.lower():
            raise HTTPException(status_code=400, detail="Email already registered")
        raise HTTPException(status_code=400, detail=error_msg)


@router.post("/login", response_model=AuthResponse)
@limiter.limit(STRICT_LIMIT)
async def login(request: Request, request_body: LoginRequest):
    try:
        supabase = get_supabase()
        response = supabase.auth.sign_in_with_password(
            {"email": request_body.email, "password": request_body.password}
        )

        if not response.user or not response.session:
            raise HTTPException(status_code=401, detail="Invalid credentials")

        profile = (
            supabase.table("profiles")
            .select("username")
            .eq("id", response.user.id)
            .maybe_single()
            .execute()
        )
        username = profile.data.get("username") if profile.data else None

        return AuthResponse(
            access_token=response.session.access_token,
            refresh_token=response.session.refresh_token,
            user=User(
                id=response.user.id, email=response.user.email or "", username=username
            ),
        )
    except HTTPException:
        raise
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid credentials")


@router.post("/logout")
@limiter.limit(PUBLIC_LIMIT)
async def logout(request: Request, user: User = Depends(get_current_user)):
    try:
        supabase = get_supabase()
        supabase.auth.sign_out()
        return {"message": "Logged out successfully"}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/refresh", response_model=AuthResponse)
@limiter.limit(REFRESH_LIMIT)
async def refresh_token(request: Request, request_body: TokenRefreshRequest):
    try:
        supabase = get_supabase()
        response = supabase.auth.refresh_session(request_body.refresh_token)

        if not response.session:
            raise HTTPException(status_code=401, detail="Invalid refresh token")

        profile = (
            supabase.table("profiles")
            .select("username")
            .eq("id", response.user.id)
            .maybe_single()
            .execute()
            if response.user
            else None
        )
        username = profile.data.get("username") if profile and profile.data else None

        return AuthResponse(
            access_token=response.session.access_token,
            refresh_token=response.session.refresh_token,
            user=User(
                id=response.user.id if response.user else "",
                email=response.user.email if response.user else "",
                username=username,
            ),
        )
    except HTTPException:
        raise
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid refresh token")


@router.post("/forgot-password")
@limiter.limit(STRICT_LIMIT)
async def forgot_password(request: Request, request_body: ForgotPasswordRequest):
    try:
        supabase = get_supabase()

        supabase.auth.reset_password_for_email(
            request_body.email,
            options={
                "redirect_to": f"{FRONTEND_URL}/reset-password",
            },
        )

        return {
            "message": "If an account with that email exists, we've sent password reset instructions."
        }
    except Exception:
        return {
            "message": "If an account with that email exists, we've sent password reset instructions."
        }


@router.post("/reset-password")
@limiter.limit(STRICT_LIMIT)
async def reset_password(request: Request, request_body: ResetPasswordRequest):
    try:
        supabase = get_supabase()

        supabase.auth.verify_oauth_token(
            {
                "type": "recovery",
                "token": request_body.token,
            }
        )

        response = supabase.auth.update_user({"password": request_body.password})

        if not response.user:
            raise HTTPException(status_code=400, detail="Password reset failed")

        return {"message": "Password reset successfully"}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail="Invalid or expired reset token")


@router.get("/me", response_model=ProfileResponse)
@limiter.limit(PUBLIC_LIMIT)
async def get_me(request: Request, user: User = Depends(get_current_user)):
    try:
        supabase = get_supabase()
        profile = (
            supabase.table("profiles").select("*").eq("id", user.id).single().execute()
        )

        if not profile.data:
            raise HTTPException(status_code=404, detail="Profile not found")

        return ProfileResponse(
            id=user.id,
            email=user.email,
            username=profile.data.get("username"),
            bio=profile.data.get("bio"),
            avatar_url=profile.data.get("avatar_url"),
            created_at=profile.data.get("created_at"),
            updated_at=profile.data.get("updated_at"),
        )
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.patch("/profile", response_model=ProfileResponse)
@limiter.limit(PUBLIC_LIMIT)
async def update_profile(
    request: Request,
    request_body: UpdateProfileRequest,
    user: User = Depends(get_current_user),
):
    try:
        supabase = get_supabase()

        update_data = {}
        if request_body.username is not None:
            existing = (
                supabase.table("profiles")
                .select("id")
                .eq("username", request_body.username)
                .neq("id", user.id)
                .maybe_single()
                .execute()
            )
            if existing.data:
                raise HTTPException(status_code=400, detail="Username already taken")
            update_data["username"] = request_body.username

        if request_body.bio is not None:
            update_data["bio"] = request_body.bio
        if request_body.avatar_url is not None:
            update_data["avatar_url"] = request_body.avatar_url

        if update_data:
            update_data["updated_at"] = "now()"
            profile = (
                supabase.table("profiles")
                .update(update_data)
                .eq("id", user.id)
                .execute()
            )
        else:
            profile = (
                supabase.table("profiles")
                .select("*")
                .eq("id", user.id)
                .single()
                .execute()
            )

        if not profile.data:
            raise HTTPException(status_code=404, detail="Profile not found")

        profile_data = (
            profile.data[0] if isinstance(profile.data, list) else profile.data
        )

        return ProfileResponse(
            id=user.id,
            email=user.email,
            username=profile_data.get("username"),
            bio=profile_data.get("bio"),
            avatar_url=profile_data.get("avatar_url"),
            created_at=profile_data.get("created_at"),
            updated_at=profile_data.get("updated_at"),
        )
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/oauth/{provider}/start")
@limiter.limit(STRICT_LIMIT)
async def oauth_start(request: Request, provider: str, redirect_uri: str):
    if provider not in ["github", "google"]:
        raise HTTPException(
            status_code=400, detail=f"Unsupported OAuth provider: {provider}"
        )

    try:
        supabase = get_supabase()

        state = secrets.token_urlsafe(32)
        oauth_states[state] = redirect_uri

        response = supabase.auth.sign_in_with_oauth(
            {
                "provider": provider,
                "options": {
                    "redirect_to": f"{FRONTEND_URL}/auth/callback?state={state}",
                    "scopes": "user:email" if provider == "github" else "email profile",
                },
            }
        )

        return {"url": response.url}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/oauth/{provider}/callback")
@limiter.limit(STRICT_LIMIT)
async def oauth_callback(
    request: Request, provider: str, code: str = "", state: str = "", error: str = ""
):
    if error:
        redirect_uri = oauth_states.pop(state, FRONTEND_URL)
        return RedirectResponse(url=f"{redirect_uri}?error={error}")

    if not code or not state:
        raise HTTPException(status_code=400, detail="Missing code or state parameter")

    redirect_uri = oauth_states.pop(state, FRONTEND_URL)

    try:
        supabase = get_supabase()

        response = supabase.auth.exchange_code_for_session(
            {
                "auth_code": code,
            }
        )

        if not response.session:
            raise HTTPException(status_code=401, detail="OAuth authentication failed")

        return RedirectResponse(
            url=f"{redirect_uri}?access_token={response.session.access_token}&refresh_token={response.session.refresh_token}"
        )
    except HTTPException:
        raise
    except Exception as e:
        return RedirectResponse(url=f"{redirect_uri}?error=oauth_failed")


# ============================================================================
# Device Code Flow for CLI Authentication
# ============================================================================

import string
import time

_DEVICE_CODE_LENGTH = 40
_USER_CODE_LENGTH = 8
_DEVICE_CODE_EXPIRY = 600  # 10 minutes
_POLL_INTERVAL = 5


def generate_device_code() -> str:
    chars = string.ascii_letters + string.digits
    return "".join(secrets.choice(chars) for _ in range(_DEVICE_CODE_LENGTH))


def generate_user_code() -> str:
    chars = string.ascii_uppercase + string.digits
    code = "".join(secrets.choice(chars) for _ in range(_USER_CODE_LENGTH))
    return f"{code[:4]}-{code[4:]}"


class DeviceCodeResponse(BaseModel):
    device_code: str
    user_code: str
    verification_uri: str
    expires_in: int
    interval: int


class DeviceVerifyRequest(BaseModel):
    user_code: str


class DeviceTokenRequest(BaseModel):
    device_code: str


class DeviceTokenResponse(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str = "bearer"
    user: User


class DeviceStatusResponse(BaseModel):
    user_code: str
    verified: bool
    client_name: str
    expires_at: str


class DeviceAuthorizationRequest(BaseModel):
    client_name: str = "Elysium CLI"


@router.post("/device/code", response_model=DeviceCodeResponse)
@limiter.limit(PUBLIC_LIMIT)
async def create_device_code(
    request: Request, req: DeviceAuthorizationRequest = DeviceAuthorizationRequest()
):
    device_code = generate_device_code()
    user_code = generate_user_code()
    expires_at = int(time.time()) + _DEVICE_CODE_EXPIRY

    supabase = get_supabase()
    supabase.table("device_codes").insert(
        {
            "device_code": device_code,
            "user_code": user_code,
            "client_name": req.client_name,
            "expires_at": f"now() + interval '{_DEVICE_CODE_EXPIRY} seconds'",
        }
    ).execute()

    return DeviceCodeResponse(
        device_code=device_code,
        user_code=user_code,
        verification_uri=f"{FRONTEND_URL}/device",
        expires_in=_DEVICE_CODE_EXPIRY,
        interval=_POLL_INTERVAL,
    )


@router.get("/device/status")
@limiter.limit(PUBLIC_LIMIT)
async def get_device_status(request: Request, user_code: str):
    supabase = get_supabase()
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
    return DeviceStatusResponse(
        user_code=row["user_code"],
        verified=row.get("verified_at") is not None,
        client_name=row.get("client_name", "Elysium CLI"),
        expires_at=row["expires_at"],
    )


@router.post("/device/verify")
@limiter.limit(PUBLIC_LIMIT)
async def verify_device_code(
    request: Request, req: DeviceVerifyRequest, user: User = Depends(get_current_user)
):
    supabase = get_supabase()

    result = (
        supabase.table("device_codes")
        .select("*")
        .eq("user_code", req.user_code.upper())
        .single()
        .execute()
    )

    if not result.data:
        raise HTTPException(status_code=404, detail="Invalid user code")

    row = result.data

    if row.get("verified_at"):
        raise HTTPException(status_code=400, detail="Device code already verified")

    expires_at = row.get("expires_at")
    if expires_at:
        from datetime import datetime, timezone

        if isinstance(expires_at, str):
            expires_dt = datetime.fromisoformat(expires_at.replace("Z", "+00:00"))
        else:
            expires_dt = expires_at
        if expires_dt < datetime.now(timezone.utc):
            raise HTTPException(status_code=400, detail="Device code has expired")

    auth_response = supabase.auth.sign_in_with_password(
        {
            "email": user.email,
            "password": secrets.token_urlsafe(32),
        }
    )

    session_resp = supabase.auth.get_session()

    if not session_resp or not session_resp.access_token:
        auth_resp = supabase.auth.admin.generate_link(
            {
                "type": "magiclink",
                "email": user.email,
            }
        )

    update_data = {
        "user_id": user.id,
        "verified_at": "now()",
    }

    supabase.table("device_codes").update(update_data).eq(
        "user_code", req.user_code.upper()
    ).execute()

    return {
        "message": "Device authorized successfully",
        "user_code": req.user_code.upper(),
    }


@router.post("/device/token", response_model=DeviceTokenResponse)
@limiter.limit(PUBLIC_LIMIT)
async def poll_device_token(request: Request, req: DeviceTokenRequest):
    supabase = get_supabase()

    result = (
        supabase.table("device_codes")
        .select("*")
        .eq("device_code", req.device_code)
        .single()
        .execute()
    )

    if not result.data:
        raise HTTPException(status_code=404, detail="Invalid device code")

    row = result.data

    expires_at = row.get("expires_at")
    if expires_at:
        from datetime import datetime, timezone

        if isinstance(expires_at, str):
            expires_dt = datetime.fromisoformat(expires_at.replace("Z", "+00:00"))
        else:
            expires_dt = expires_at
        if expires_dt < datetime.now(timezone.utc):
            raise HTTPException(status_code=400, detail="Device code has expired")

    if not row.get("verified_at"):
        raise HTTPException(status_code=400, detail="Authorization pending")

    if not row.get("user_id"):
        raise HTTPException(status_code=400, detail="Authorization pending")

    user_id = row["user_id"]

    user_result = (
        supabase.table("profiles").select("*").eq("id", user_id).single().execute()
    )
    profile = user_result.data if user_result.data else {}

    magic_link = supabase.auth.admin.generate_link(
        {
            "type": "magiclink",
            "email": profile.get("email") or user_id,
        }
    )

    if not magic_link or not hasattr(magic_link, "properties"):
        raise HTTPException(status_code=500, detail="Failed to generate session")

    supabase.table("device_codes").delete().eq("device_code", req.device_code).execute()

    return DeviceTokenResponse(
        access_token=magic_link.properties.get("access_token", ""),
        refresh_token=magic_link.properties.get("refresh_token", ""),
        user=User(
            id=user_id,
            email=profile.get("email", ""),
            username=profile.get("username"),
        ),
    )
