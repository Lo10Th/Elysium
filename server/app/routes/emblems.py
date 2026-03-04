"""Emblem routes — thin HTTP handlers.

All business logic lives in app.services.emblem_service.EmblemService.
Routes are responsible only for:
- Declaring FastAPI path operations and dependencies
- Extracting / validating request data (via Pydantic models and Query params)
- Calling the service layer
- Returning the service result directly
"""

from typing import List, Optional

from fastapi import APIRouter, Depends, Query, Request

from app.database import get_supabase, run_sync
from app.limiter import limiter, PUBLIC_LIMIT, AUTH_LIMIT
from app.models import Emblem, EmblemCreate, EmblemUpdate, User
from app.routes.auth import get_current_user
from app.services.emblem_service import EmblemService

router = APIRouter()


@router.get("", response_model=List[Emblem])
@limiter.limit(PUBLIC_LIMIT)
async def list_emblems(
    request: Request,
    category: Optional[str] = Query(None),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
):
    supabase = get_supabase()
    return await run_sync(EmblemService.list_emblems, supabase, category, limit, offset)


@router.get("/search", response_model=List[Emblem])
@limiter.limit(PUBLIC_LIMIT)
async def search_emblems(
    request: Request,
    q: str = Query(..., max_length=200),
    category: Optional[str] = Query(None, max_length=100),
    sort: str = Query("downloads"),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
):
    supabase = get_supabase()
    return await run_sync(
        EmblemService.search_emblems, supabase, q, category, sort, limit, offset
    )


@router.get("/{name}", response_model=Emblem)
@limiter.limit(PUBLIC_LIMIT)
async def get_emblem(request: Request, name: str):
    supabase = get_supabase()
    return await run_sync(EmblemService.get_emblem, supabase, name)


@router.get("/{name}/{version}", response_model=dict)
@limiter.limit(PUBLIC_LIMIT)
async def get_emblem_version(request: Request, name: str, version: str):
    supabase = get_supabase()
    return await run_sync(EmblemService.get_emblem_version, supabase, name, version)


@router.post("", response_model=Emblem)
@limiter.limit(AUTH_LIMIT)
async def create_emblem(
    request: Request,
    request_body: EmblemCreate,
    user: User = Depends(get_current_user),
):
    EmblemService.validate_yaml(request_body.yaml_content)
    supabase = get_supabase()
    return await run_sync(EmblemService.create_emblem, supabase, request_body, user)


@router.put("/{name}", response_model=Emblem)
@limiter.limit(AUTH_LIMIT)
async def update_emblem(
    request: Request,
    name: str,
    request_body: EmblemUpdate,
    user: User = Depends(get_current_user),
):
    EmblemService.validate_yaml(request_body.yaml_content)
    supabase = get_supabase()
    return await run_sync(EmblemService.update_emblem, supabase, name, request_body, user)


@router.delete("/{name}")
@limiter.limit(AUTH_LIMIT)
async def delete_emblem(
    request: Request, name: str, user: User = Depends(get_current_user)
):
    supabase = get_supabase()
    return await run_sync(EmblemService.delete_emblem, supabase, name, user)
