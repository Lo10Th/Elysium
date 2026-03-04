"""Authentication routes — thin HTTP handlers.

All business logic lives in app.services.auth_service.AuthService.
Routes are responsible only for:
- Declaring FastAPI path operations and dependencies
- Request/response model definitions (Pydantic)
- Calling the service layer
- Returning the service result directly

The ``get_current_user`` / ``get_current_user_optional`` dependency functions
remain here because they are imported by emblems.py and keys.py.
"""

import re

from fastapi import APIRouter, Depends, HTTPException, Request
from fastapi.responses import RedirectResponse
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel, EmailStr, field_validator

from app.config import get_settings
from app.database import get_supabase
from app.limiter import (
    limiter,
    PUBLIC_LIMIT,
    STRICT_LIMIT,
    REGISTER_LIMIT,
    REFRESH_LIMIT,
)
from app.models import User
from app.services.auth_service import AuthService

router = APIRouter()
security = HTTPBearer()

FRONTEND_URL = get_settings().FRONTEND_URL

# Module-level state dict for in-process OAuth CSRF protection.
oauth_states: dict[str, str] = {}

_USERNAME_RE = re.compile(r"^[a-zA-Z0-9_-]{3,30}$")


# ---------------------------------------------------------------------------
# Request / response models
# ---------------------------------------------------------------------------


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
                "username must be 3-30 characters and contain only letters, "
                "digits, underscores, or hyphens"
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
                "username must be 3-30 characters and contain only letters, "
                "digits, underscores, or hyphens"
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


# Device-code models (kept here so CLI integration tests can import them
# without taking on an auth_service dependency).
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


# ---------------------------------------------------------------------------
# Auth dependency (imported by emblems.py and keys.py)
# ---------------------------------------------------------------------------


async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
) -> User:
    supabase = get_supabase()
    return AuthService.get_user_from_token(supabase, credentials.credentials)


async def get_current_user_optional(
    credentials: HTTPAuthorizationCredentials = Depends(security),
) -> User | None:
    supabase = get_supabase()
    return AuthService.get_user_from_token_optional(supabase, credentials.credentials)


# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------


@router.post("/register", response_model=AuthResponse)
@limiter.limit(REGISTER_LIMIT)
async def register(request: Request, request_body: RegisterRequest):
    supabase = get_supabase()
    return AuthService.register(
        supabase,
        request_body.email,
        request_body.password,
        request_body.username,
        FRONTEND_URL,
    )


@router.post("/login", response_model=AuthResponse)
@limiter.limit(STRICT_LIMIT)
async def login(request: Request, request_body: LoginRequest):
    supabase = get_supabase()
    return AuthService.login(supabase, request_body.email, request_body.password)


@router.post("/logout")
@limiter.limit(PUBLIC_LIMIT)
async def logout(request: Request, user: User = Depends(get_current_user)):
    supabase = get_supabase()
    return AuthService.logout(supabase)


@router.post("/refresh", response_model=AuthResponse)
@limiter.limit(REFRESH_LIMIT)
async def refresh_token(request: Request, request_body: TokenRefreshRequest):
    supabase = get_supabase()
    return AuthService.refresh_token(supabase, request_body.refresh_token)


@router.post("/forgot-password")
@limiter.limit(STRICT_LIMIT)
async def forgot_password(request: Request, request_body: ForgotPasswordRequest):
    supabase = get_supabase()
    return AuthService.forgot_password(supabase, request_body.email, FRONTEND_URL)


@router.post("/reset-password")
@limiter.limit(STRICT_LIMIT)
async def reset_password(request: Request, request_body: ResetPasswordRequest):
    supabase = get_supabase()
    return AuthService.reset_password(supabase, request_body.token, request_body.password)


@router.get("/me", response_model=ProfileResponse)
@limiter.limit(PUBLIC_LIMIT)
async def get_me(request: Request, user: User = Depends(get_current_user)):
    supabase = get_supabase()
    return AuthService.get_profile(supabase, user.id, user.email)


@router.patch("/profile", response_model=ProfileResponse)
@limiter.limit(PUBLIC_LIMIT)
async def update_profile(
    request: Request,
    request_body: UpdateProfileRequest,
    user: User = Depends(get_current_user),
):
    supabase = get_supabase()
    return AuthService.update_profile(
        supabase,
        user.id,
        user.email,
        request_body.username,
        request_body.bio,
        request_body.avatar_url,
    )


@router.get("/oauth/{provider}/start")
@limiter.limit(STRICT_LIMIT)
async def oauth_start(request: Request, provider: str, redirect_uri: str):
    if provider not in ["github", "google"]:
        raise HTTPException(
            status_code=400, detail=f"Unsupported OAuth provider: {provider}"
        )
    supabase = get_supabase()
    return AuthService.oauth_start(
        supabase, provider, redirect_uri, FRONTEND_URL, oauth_states
    )


@router.get("/oauth/{provider}/callback")
@limiter.limit(STRICT_LIMIT)
async def oauth_callback(
    request: Request,
    provider: str,
    code: str = "",
    state: str = "",
    error: str = "",
):
    supabase = get_supabase()
    return AuthService.oauth_callback(
        supabase, provider, code, state, error, FRONTEND_URL, oauth_states
    )


# ---------------------------------------------------------------------------
# Device-code flow
# ---------------------------------------------------------------------------


@router.post("/device/code", response_model=DeviceCodeResponse)
@limiter.limit(PUBLIC_LIMIT)
async def create_device_code(
    request: Request,
    req: DeviceAuthorizationRequest = DeviceAuthorizationRequest(),
):
    supabase = get_supabase()
    return AuthService.create_device_code(supabase, req.client_name, FRONTEND_URL)


@router.get("/device/status")
@limiter.limit(PUBLIC_LIMIT)
async def get_device_status(request: Request, user_code: str):
    supabase = get_supabase()
    return AuthService.get_device_status(supabase, user_code)


@router.post("/device/verify")
@limiter.limit(PUBLIC_LIMIT)
async def verify_device_code(
    request: Request,
    req: DeviceVerifyRequest,
    user: User = Depends(get_current_user),
):
    supabase = get_supabase()
    return AuthService.verify_device_code(supabase, req.user_code, user)


@router.post("/device/token", response_model=DeviceTokenResponse)
@limiter.limit(PUBLIC_LIMIT)
async def poll_device_token(request: Request, req: DeviceTokenRequest):
    supabase = get_supabase()
    return AuthService.poll_device_token(supabase, req.device_code)
