from fastapi import APIRouter, HTTPException, Depends, Request
from fastapi.responses import RedirectResponse
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel, EmailStr
from app.database import get_supabase
from app.limiter import limiter, PUBLIC_LIMIT, STRICT_LIMIT
from app.models import User
import urllib.parse
import secrets

router = APIRouter()
security = HTTPBearer()

oauth_states: dict[str, str] = {}


class LoginRequest(BaseModel):
    email: str
    password: str


class RegisterRequest(BaseModel):
    email: EmailStr
    password: str
    username: str


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
async def register(request: Request, request_body: RegisterRequest):
    try:
        supabase = get_supabase()
        response = supabase.auth.sign_up(
            {
                "email": request_body.email,
                "password": request_body.password,
                "options": {"data": {"username": request_body.username}},
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
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


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

        return AuthResponse(
            access_token=response.session.access_token,
            refresh_token=response.session.refresh_token,
            user=User(
                id=response.user.id, email=response.user.email or "", username=None
            ),
        )
    except Exception as e:
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
@limiter.limit("20/minute")
async def refresh_token(request: Request, request_body: TokenRefreshRequest):
    try:
        supabase = get_supabase()
        response = supabase.auth.refresh_session(request_body.refresh_token)

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
@limiter.limit(PUBLIC_LIMIT)
async def get_me(request: Request, user: User = Depends(get_current_user)):
    return user


@router.get("/oauth/start")
@limiter.limit(STRICT_LIMIT)
async def oauth_start(request: Request, redirect_uri: str):
    raise HTTPException(
        status_code=501,
        detail="OAuth login is not implemented yet. Please use email/password login via 'ely login' or POST /api/auth/login",
    )


@router.get("/oauth/callback")
@limiter.limit(STRICT_LIMIT)
async def oauth_callback(request: Request, state: str, code: str = "", error: str = ""):
    raise HTTPException(
        status_code=501,
        detail="OAuth login is not implemented yet. Please use email/password login via 'ely login' or POST /api/auth/login",
    )
