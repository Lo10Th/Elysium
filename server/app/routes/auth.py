from fastapi import APIRouter, HTTPException, Depends, Request
from fastapi.responses import RedirectResponse
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel, EmailStr, field_validator
from app.database import get_supabase
from app.models import User
from slowapi import Limiter
from slowapi.util import get_remote_address
import urllib.parse
import secrets
import re

router = APIRouter()
security = HTTPBearer()
limiter = Limiter(key_func=get_remote_address)

oauth_states: dict[str, str] = {}

_USERNAME_RE = re.compile(r"^[a-zA-Z0-9_-]{3,30}$")


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

        return User(id=response.user.id, email=response.user.email or "", username=None)
    except Exception as e:
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

        return User(id=response.user.id, email=response.user.email or "", username=None)
    except:
        return None


@router.post("/register", response_model=AuthResponse)
@limiter.limit("5/minute")
async def register(request: Request, body: RegisterRequest):
    try:
        supabase = get_supabase()
        response = supabase.auth.sign_up(
            {
                "email": body.email,
                "password": body.password,
                "options": {"data": {"username": body.username}},
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
                username=body.username,
            ),
        )
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/login", response_model=AuthResponse)
@limiter.limit("10/minute")
async def login(request: Request, body: LoginRequest):
    try:
        supabase = get_supabase()
        response = supabase.auth.sign_in_with_password(
            {"email": body.email, "password": body.password}
        )

        if not response.user or not response.session:
            raise HTTPException(status_code=401, detail="Invalid credentials")

        return AuthResponse(
            access_token=response.session.access_token,
            refresh_token=response.session.refresh_token,
            user=User(
                id=response.user.id, email=response.user.email or "", username=None
            ),
        )
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=401, detail="Invalid credentials")


@router.post("/logout")
async def logout(user: User = Depends(get_current_user)):
    try:
        supabase = get_supabase()
        supabase.auth.sign_out()
        return {"message": "Logged out successfully"}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/refresh", response_model=AuthResponse)
async def refresh_token(request: TokenRefreshRequest):
    try:
        supabase = get_supabase()
        response = supabase.auth.refresh_session(request.refresh_token)

        if not response.session:
            raise HTTPException(status_code=401, detail="Invalid refresh token")

        return AuthResponse(
            access_token=response.session.access_token,
            refresh_token=response.session.refresh_token,
            user=User(
                id=response.user.id if response.user else "",
                email=response.user.email if response.user else "",
                username=None,
            ),
        )
    except Exception as e:
        raise HTTPException(status_code=401, detail="Invalid refresh token")


@router.get("/me", response_model=User)
async def get_me(user: User = Depends(get_current_user)):
    return user


@router.get("/oauth/start")
async def oauth_start(redirect_uri: str, request: Request):
    raise HTTPException(
        status_code=501,
        detail="OAuth login is not implemented yet. Please use email/password login via 'ely login' or POST /api/auth/login",
    )


@router.get("/oauth/callback")
async def oauth_callback(state: str, code: str = "", error: str = ""):
    raise HTTPException(
        status_code=501,
        detail="OAuth login is not implemented yet. Please use email/password login via 'ely login' or POST /api/auth/login",
    )
