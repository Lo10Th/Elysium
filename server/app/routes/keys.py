"""API key routes — thin HTTP handlers.

All business logic lives in app.services.key_service.KeyService.
Routes are responsible only for:
- Declaring FastAPI path operations and dependencies
- Calling the service layer
- Returning the service result directly
"""

from fastapi import APIRouter, Depends, Request

from app.database import get_supabase
from app.limiter import limiter, PUBLIC_LIMIT, AUTH_LIMIT
from app.models import KeyCreate, User
from app.routes.auth import get_current_user
from app.services.key_service import KeyService

router = APIRouter()


@router.get("")
@limiter.limit(PUBLIC_LIMIT)
async def list_keys(request: Request, user: User = Depends(get_current_user)):
    supabase = get_supabase()
    return KeyService.list_keys(supabase, user.id)


@router.post("", status_code=201)
@limiter.limit(AUTH_LIMIT)
async def create_key(
    request: Request,
    request_body: KeyCreate,
    user: User = Depends(get_current_user),
):
    supabase = get_supabase()
    return KeyService.create_key(supabase, user.id, request_body)


@router.get("/{key_id}")
@limiter.limit(PUBLIC_LIMIT)
async def get_key(
    request: Request, key_id: str, user: User = Depends(get_current_user)
):
    supabase = get_supabase()
    return KeyService.get_key(supabase, user.id, key_id)


@router.delete("/{key_id}", status_code=204)
@limiter.limit(AUTH_LIMIT)
async def delete_key(
    request: Request, key_id: str, user: User = Depends(get_current_user)
):
    supabase = get_supabase()
    KeyService.delete_key(supabase, user.id, key_id)
    return None
