from fastapi import APIRouter, HTTPException, Depends, Query, Request
from typing import Optional, List
from app.database import get_supabase
from app.limiter import limiter, PUBLIC_LIMIT, AUTH_LIMIT
from app.models import (
    Emblem,
    EmblemCreate,
    EmblemUpdate,
    EmblemVersion,
    EmblemWithVersion,
    User,
)
from app.routes.auth import get_current_user
import yaml
import json
from jsonschema import validate, ValidationError as JsonSchemaError
from pathlib import Path

router = APIRouter()

SCHEMA_PATH = (
    Path(__file__).parent.parent.parent.parent / "schemas" / "emblem.schema.json"
)


def load_schema() -> dict:
    with open(SCHEMA_PATH) as f:
        return json.load(f)


def validate_emblem_yaml(yaml_content: str) -> dict:
    try:
        data = yaml.safe_load(yaml_content)
        schema = load_schema()
        validate(instance=data, schema=schema)
        return data
    except yaml.YAMLError as e:
        raise HTTPException(status_code=400, detail=f"Invalid YAML: {str(e)}")
    except JsonSchemaError as e:
        raise HTTPException(
            status_code=400, detail=f"Schema validation failed: {e.message}"
        )


@router.get("", response_model=List[Emblem])
@limiter.limit(PUBLIC_LIMIT)
async def list_emblems(
    request: Request,
    category: Optional[str] = Query(None),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
):
    try:
        supabase = get_supabase()
        query = supabase.table("emblems").select("*, profiles(username)")

        if category:
            query = query.eq("category", category)

        response = (
            query.order("downloads_count", desc=True)
            .range(offset, offset + limit - 1)
            .execute()
        )

        emblems = []
        for row in response.data:
            emblem = Emblem(
                id=row["id"],
                name=row["name"],
                description=row["description"],
                author_id=row.get("author_id"),
                author_name=row.get("profiles", {}).get("username")
                if row.get("profiles")
                else None,
                category=row.get("category"),
                tags=row.get("tags"),
                license=row.get("license", "MIT"),
                repository_url=row.get("repository_url"),
                homepage_url=row.get("homepage_url"),
                latest_version=row.get("latest_version"),
                downloads_count=row.get("downloads_count", 0),
                created_at=row["created_at"],
                updated_at=row["updated_at"],
            )
            emblems.append(emblem)

        return emblems
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


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
    try:
        supabase = get_supabase()
        query = supabase.table("emblems").select("*, profiles(username)")

        # Escape LIKE metacharacters to prevent pattern injection
        safe_q = q.replace("\\", "\\\\").replace("%", "\\%").replace("_", "\\_")
        query = query.or_(f"name.ilike.%{safe_q}%,description.ilike.%{safe_q}%")

        if category:
            query = query.eq("category", category)

        if sort == "downloads":
            query = query.order("downloads_count", desc=True)
        elif sort == "recent":
            query = query.order("created_at", desc=True)
        elif sort == "name":
            query = query.order("name")

        response = query.range(offset, offset + limit - 1).execute()

        emblems = []
        for row in response.data:
            emblem = Emblem(
                id=row["id"],
                name=row["name"],
                description=row["description"],
                author_id=row.get("author_id"),
                author_name=row.get("profiles", {}).get("username")
                if row.get("profiles")
                else None,
                category=row.get("category"),
                tags=row.get("tags"),
                license=row.get("license", "MIT"),
                repository_url=row.get("repository_url"),
                homepage_url=row.get("homepage_url"),
                latest_version=row.get("latest_version"),
                downloads_count=row.get("downloads_count", 0),
                created_at=row["created_at"],
                updated_at=row["updated_at"],
            )
            emblems.append(emblem)

        return emblems
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/{name}", response_model=Emblem)
@limiter.limit(PUBLIC_LIMIT)
async def get_emblem(request: Request, name: str):
    try:
        supabase = get_supabase()
        response = (
            supabase.table("emblems")
            .select("*, profiles(username)")
            .eq("name", name)
            .single()
            .execute()
        )

        if not response.data:
            raise HTTPException(status_code=404, detail="Emblem not found")

        row = response.data
        return Emblem(
            id=row["id"],
            name=row["name"],
            description=row["description"],
            author_id=row.get("author_id"),
            author_name=row.get("profiles", {}).get("username")
            if row.get("profiles")
            else None,
            category=row.get("category"),
            tags=row.get("tags"),
            license=row.get("license", "MIT"),
            repository_url=row.get("repository_url"),
            homepage_url=row.get("homepage_url"),
            latest_version=row.get("latest_version"),
            downloads_count=row.get("downloads_count", 0),
            created_at=row["created_at"],
            updated_at=row["updated_at"],
        )
    except Exception as e:
        if "not found" in str(e).lower():
            raise HTTPException(status_code=404, detail="Emblem not found")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/{name}/{version}", response_model=dict)
@limiter.limit(PUBLIC_LIMIT)
async def get_emblem_version(request: Request, name: str, version: str):
    try:
        supabase = get_supabase()
        emblem_response = (
            supabase.table("emblems").select("id").eq("name", name).single().execute()
        )

        if not emblem_response.data:
            raise HTTPException(status_code=404, detail="Emblem not found")

        emblem_id = emblem_response.data["id"]

        version_response = (
            supabase.table("emblem_versions")
            .select("*")
            .eq("emblem_id", emblem_id)
            .eq("version", version)
            .single()
            .execute()
        )

        if not version_response.data:
            raise HTTPException(status_code=404, detail="Version not found")

        return {
            "name": name,
            "version": version,
            "yaml_content": version_response.data["yaml_content"],
            "changelog": version_response.data.get("changelog"),
            "published_at": version_response.data["published_at"],
        }
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("", response_model=Emblem)
@limiter.limit(AUTH_LIMIT)
async def create_emblem(request: Request, request_body: EmblemCreate, user: User = Depends(get_current_user)):
    validate_emblem_yaml(request_body.yaml_content)

    try:
        supabase = get_supabase()
        existing = (
            supabase.table("emblems").select("id").eq("name", request_body.name).execute()
        )

        if existing.data:
            raise HTTPException(
                status_code=400, detail="Emblem with this name already exists"
            )

        emblem_response = (
            supabase.table("emblems")
            .insert(
                {
                    "name": request_body.name,
                    "description": request_body.description,
                    "author_id": user.id,
                    "category": request_body.category,
                    "tags": request_body.tags,
                    "license": request_body.license,
                    "repository_url": request_body.repository_url,
                    "homepage_url": request_body.homepage_url,
                    "latest_version": request_body.version,
                }
            )
            .execute()
        )

        if not emblem_response.data:
            raise HTTPException(status_code=500, detail="Failed to create emblem")

        emblem_id = emblem_response.data[0]["id"]

        supabase.table("emblem_versions").insert(
            {
                "emblem_id": emblem_id,
                "version": request_body.version,
                "yaml_content": request_body.yaml_content,
                "published_by": user.id,
            }
        ).execute()

        supabase.table("emblem_pulls").insert(
            {"emblem_id": emblem_id, "version": request_body.version, "pulled_by": user.id}
        ).execute()

        return Emblem(
            id=emblem_id,
            name=request_body.name,
            description=request_body.description,
            author_id=user.id,
            author_name=user.username,
            category=request_body.category,
            tags=request_body.tags,
            license=request_body.license,
            repository_url=request_body.repository_url,
            homepage_url=request_body.homepage_url,
            latest_version=request_body.version,
            downloads_count=0,
            created_at=emblem_response.data[0]["created_at"],
            updated_at=emblem_response.data[0]["updated_at"],
        )
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.put("/{name}", response_model=Emblem)
@limiter.limit(AUTH_LIMIT)
async def update_emblem(
    request: Request, name: str, request_body: EmblemUpdate, user: User = Depends(get_current_user)
):
    validate_emblem_yaml(request_body.yaml_content)

    try:
        supabase = get_supabase()
        emblem_response = (
            supabase.table("emblems").select("*").eq("name", name).single().execute()
        )

        if not emblem_response.data:
            raise HTTPException(status_code=404, detail="Emblem not found")

        if emblem_response.data["author_id"] != user.id:
            raise HTTPException(
                status_code=403, detail="Not authorized to update this emblem"
            )

        emblem_id = emblem_response.data["id"]

        existing_version = (
            supabase.table("emblem_versions")
            .select("id")
            .eq("emblem_id", emblem_id)
            .eq("version", request_body.version)
            .execute()
        )

        if existing_version.data:
            raise HTTPException(status_code=400, detail="Version already exists")

        supabase.table("emblem_versions").insert(
            {
                "emblem_id": emblem_id,
                "version": request_body.version,
                "yaml_content": request_body.yaml_content,
                "changelog": request_body.description,
                "published_by": user.id,
            }
        ).execute()

        supabase.table("emblems").update(
            {
                "description": request_body.description
                or emblem_response.data["description"],
                "latest_version": request_body.version,
            }
        ).eq("id", emblem_id).execute()

        return Emblem(
            id=emblem_id,
            name=name,
            description=request_body.description or emblem_response.data["description"],
            author_id=user.id,
            author_name=user.username,
            category=emblem_response.data.get("category"),
            tags=emblem_response.data.get("tags"),
            license=emblem_response.data.get("license", "MIT"),
            repository_url=emblem_response.data.get("repository_url"),
            homepage_url=emblem_response.data.get("homepage_url"),
            latest_version=request_body.version,
            downloads_count=emblem_response.data.get("downloads_count", 0),
            created_at=emblem_response.data["created_at"],
            updated_at=emblem_response.data["updated_at"],
        )
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.delete("/{name}")
@limiter.limit(AUTH_LIMIT)
async def delete_emblem(request: Request, name: str, user: User = Depends(get_current_user)):
    try:
        supabase = get_supabase()
        emblem_response = (
            supabase.table("emblems").select("*").eq("name", name).single().execute()
        )

        if not emblem_response.data:
            raise HTTPException(status_code=404, detail="Emblem not found")

        if emblem_response.data["author_id"] != user.id:
            raise HTTPException(
                status_code=403, detail="Not authorized to delete this emblem"
            )

        supabase.table("emblems").delete().eq(
            "id", emblem_response.data["id"]
        ).execute()

        return {"message": "Emblem deleted successfully"}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
