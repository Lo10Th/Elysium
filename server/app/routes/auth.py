from fastapi import APIRouter, HTTPException, Depends, Request
from fastapi.responses import RedirectResponse
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel, EmailStr
from app.database import supabase
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


async def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security)) -> User:
    try:
        token = credentials.credentials
        response = supabase.auth.get_user(token)
        
        if not response.user:
            raise HTTPException(status_code=401, detail="Invalid token")
        
        return User(
            id=response.user.id,
            email=response.user.email or "",
            username=None
        )
    except Exception as e:
        raise HTTPException(status_code=401, detail="Invalid token")


async def get_current_user_optional(credentials: HTTPAuthorizationCredentials = Depends(security)) -> User | None:
    try:
        token = credentials.credentials
        response = supabase.auth.get_user(token)
        
        if not response.user:
            return None
        
        return User(
            id=response.user.id,
            email=response.user.email or "",
            username=None
        )
    except:
        return None


@router.post("/register", response_model=AuthResponse)
async def register(request: RegisterRequest):
    try:
        response = supabase.auth.sign_up({
            "email": request.email,
            "password": request.password,
            "options": {
                "data": {
                    "username": request.username
                }
            }
        })
        
        if not response.user:
            raise HTTPException(status_code=400, detail="Registration failed")
        
        return AuthResponse(
            access_token=response.session.access_token if response.session else "",
            refresh_token=response.session.refresh_token if response.session else "",
            user=User(
                id=response.user.id,
                email=response.user.email or "",
                username=request.username
            )
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/login", response_model=AuthResponse)
async def login(request: LoginRequest):
    try:
        response = supabase.auth.sign_in_with_password({
            "email": request.email,
            "password": request.password
        })
        
        if not response.user or not response.session:
            raise HTTPException(status_code=401, detail="Invalid credentials")
        
        return AuthResponse(
            access_token=response.session.access_token,
            refresh_token=response.session.refresh_token,
            user=User(
                id=response.user.id,
                email=response.user.email or "",
                username=None
            )
        )
    except Exception as e:
        raise HTTPException(status_code=401, detail="Invalid credentials")


@router.post("/logout")
async def logout(user: User = Depends(get_current_user)):
    try:
        supabase.auth.sign_out()
        return {"message": "Logged out successfully"}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/refresh", response_model=AuthResponse)
async def refresh_token(request: TokenRefreshRequest):
    try:
        response = supabase.auth.refresh_session(request.refresh_token)
        
        if not response.session:
            raise HTTPException(status_code=401, detail="Invalid refresh token")
        
        return AuthResponse(
            access_token=response.session.access_token,
            refresh_token=response.session.refresh_token,
            user=User(
                id=response.user.id if response.user else "",
                email=response.user.email if response.user else "",
                username=None
            )
        )
    except Exception as e:
        raise HTTPException(status_code=401, detail="Invalid refresh token")


@router.get("/me", response_model=User)
async def get_me(user: User = Depends(get_current_user)):
    return user


@router.get("/oauth/start")
async def oauth_start(redirect_uri: str, request: Request):
    state = secrets.token_urlsafe(32)
    oauth_states[state] = redirect_uri
    
    base_url = "https://supabase.example.com/auth/v1/authorize"
    callback_url = f"{request.url.scheme}://{request.url.netloc}/api/auth/oauth/callback?state={state}"
    params = {
        "provider": "github",
        "redirect_to": callback_url,
    }
    
    auth_url = f"{base_url}?{urllib.parse.urlencode(params)}"
    return RedirectResponse(url=auth_url)


@router.get("/oauth/callback")
async def oauth_callback(state: str, code: str = "", error: str = ""):
    if error:
        redirect_uri = oauth_states.pop(state, "http://localhost:8080/callback")
        error_params = urllib.parse.urlencode({"error": error})
        return RedirectResponse(url=f"{redirect_uri}?{error_params}")
    
    if state not in oauth_states:
        raise HTTPException(status_code=400, detail="Invalid or expired state")
    
    redirect_uri = oauth_states.pop(state)
    
    try:
        response = supabase.auth.exchange_code_for_session(code)
        
        if not response.session:
            raise HTTPException(status_code=401, detail="Authentication failed")
        
        token_params = urllib.parse.urlencode({
            "access_token": response.session.access_token,
            "refresh_token": response.session.refresh_token,
            "token_type": "bearer",
        })
        
        return RedirectResponse(url=f"{redirect_uri}?{token_params}")
    except Exception as e:
        error_params = urllib.parse.urlencode({"error": str(e)})
        return RedirectResponse(url=f"{redirect_uri}?{error_params}")