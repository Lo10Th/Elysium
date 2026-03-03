from fastapi import APIRouter, HTTPException, Depends
from typing import List, Optional
from datetime import datetime, timedelta
from pydantic import BaseModel, field_validator
import secrets
import hashlib
from app.database import get_supabase
from app.routes.auth import get_current_user
from app.models import User

router = APIRouter()


class KeyCreate(BaseModel):
    name: str
    expires_days: Optional[int] = None

    @field_validator("name")
    @classmethod
    def name_length(cls, v: str) -> str:
        v = v.strip()
        if not v:
            raise ValueError("name must not be empty")
        if len(v) > 64:
            raise ValueError("name must be at most 64 characters")
        return v

    @field_validator("expires_days")
    @classmethod
    def expires_days_range(cls, v: Optional[int]) -> Optional[int]:
        if v is not None and (v < 1 or v > 365):
            raise ValueError("expires_days must be between 1 and 365")
        return v


class KeyResponse(BaseModel):
    id: str
    name: str
    key: Optional[str] = None
    created_at: datetime
    expires_at: Optional[datetime] = None


class KeyListItem(BaseModel):
    id: str
    name: str
    created_at: datetime
    expires_at: Optional[datetime] = None


def generate_api_key() -> str:
    return f"ely_{secrets.token_urlsafe(32)}"


def hash_key(key: str) -> str:
    return hashlib.sha256(key.encode()).hexdigest()


@router.get("", response_model=List[KeyListItem])
async def list_keys(user: User = Depends(get_current_user)):
    try:
        supabase = get_supabase()
        response = (
            supabase.table("api_keys")
            .select("id, name, created_at, expires_at")
            .eq("user_id", user.id)
            .execute()
        )

        keys = []
        for row in response.data:
            keys.append(
                KeyListItem(
                    id=row["id"],
                    name=row["name"],
                    created_at=datetime.fromisoformat(
                        row["created_at"].replace("Z", "+00:00")
                    ),
                    expires_at=datetime.fromisoformat(
                        row["expires_at"].replace("Z", "+00:00")
                    )
                    if row.get("expires_at")
                    else None,
                )
            )

        return keys
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("", response_model=KeyResponse, status_code=201)
async def create_key(request: KeyCreate, user: User = Depends(get_current_user)):
    supabase = get_supabase()
    existing = (
        supabase.table("api_keys")
        .select("id")
        .eq("user_id", user.id)
        .eq("name", request.name)
        .execute()
    )

    if existing.data:
        raise HTTPException(
            status_code=400, detail="A key with this name already exists"
        )

    raw_key = generate_api_key()
    key_hash = hash_key(raw_key)

    expires_at = None
    if request.expires_days:
        expires_at = (
            datetime.utcnow() + timedelta(days=request.expires_days)
        ).isoformat() + "Z"

    try:
        response = (
            supabase.table("api_keys")
            .insert(
                {
                    "user_id": user.id,
                    "name": request.name,
                    "key_hash": key_hash,
                    "expires_at": expires_at,
                }
            )
            .execute()
        )

        if not response.data:
            raise HTTPException(status_code=500, detail="Failed to create key")

        row = response.data[0]
        created_at = datetime.fromisoformat(row["created_at"].replace("Z", "+00:00"))

        return KeyResponse(
            id=row["id"],
            name=request.name,
            key=raw_key,
            created_at=created_at,
            expires_at=datetime.fromisoformat(row["expires_at"].replace("Z", "+00:00"))
            if row.get("expires_at")
            else None,
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/{key_id}", response_model=KeyListItem)
async def get_key(key_id: str, user: User = Depends(get_current_user)):
    try:
        supabase = get_supabase()
        response = (
            supabase.table("api_keys")
            .select("id, name, created_at, expires_at")
            .eq("id", key_id)
            .eq("user_id", user.id)
            .single()
            .execute()
        )

        if not response.data:
            raise HTTPException(status_code=404, detail="Key not found")

        row = response.data
        return KeyListItem(
            id=row["id"],
            name=row["name"],
            created_at=datetime.fromisoformat(row["created_at"].replace("Z", "+00:00")),
            expires_at=datetime.fromisoformat(row["expires_at"].replace("Z", "+00:00"))
            if row.get("expires_at")
            else None,
        )
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.delete("/{key_id}", status_code=204)
async def delete_key(key_id: str, user: User = Depends(get_current_user)):
    try:
        supabase = get_supabase()
        response = (
            supabase.table("api_keys")
            .select("id")
            .eq("id", key_id)
            .eq("user_id", user.id)
            .single()
            .execute()
        )

        if not response.data:
            raise HTTPException(status_code=404, detail="Key not found")

        supabase.table("api_keys").delete().eq("id", key_id).eq(
            "user_id", user.id
        ).execute()

        return None
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
