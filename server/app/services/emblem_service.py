"""Service layer for emblem business logic.

All DB operations and business rules related to emblems live here.
Routes call these methods and handle HTTP concerns (status codes,
request parsing, response serialization).

Services accept the Supabase client as a parameter so that route-level
mocks continue to work in tests without any conftest changes.
"""

import json
import logging
import yaml
from pathlib import Path
from typing import List, Optional

from fastapi import HTTPException
from jsonschema import validate, ValidationError as JsonSchemaError
from supabase import Client

from app.models import Emblem, EmblemCreate, EmblemUpdate, User

logger = logging.getLogger(__name__)

SCHEMA_PATH = (
    Path(__file__).parent.parent.parent.parent / "schemas" / "emblem.schema.json"
)


def _load_schema() -> dict:
    """Load the JSON schema for emblem YAML validation."""
    with open(SCHEMA_PATH) as f:
        return json.load(f)


def _row_to_emblem(row: dict) -> Emblem:
    """Convert a Supabase DB row (with optional joined profiles) to an Emblem model.

    Centralises the repeated construction that previously appeared in five
    different route handlers.
    """
    profiles = row.get("profiles")
    author_name: Optional[str] = None
    author_verified: Optional[bool] = None

    if isinstance(profiles, dict):
        author_name = profiles.get("username")
        author_verified = profiles.get("is_verified")
    else:
        # RPC responses return author_name and author_verified directly as flat fields
        author_name = row.get("author_name")
        author_verified = row.get("author_verified")

    return Emblem(
        id=row["id"],
        name=row["name"],
        description=row["description"],
        author_id=row.get("author_id"),
        author_name=author_name,
        author_verified=author_verified,
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


class EmblemService:
    """Business logic for emblem management."""

    @staticmethod
    def validate_yaml(yaml_content: str) -> dict:
        """Parse and schema-validate emblem YAML content.

        Raises HTTPException 400 on invalid YAML or schema failure.
        """
        try:
            data = yaml.safe_load(yaml_content)
            schema = _load_schema()
            validate(instance=data, schema=schema)
            return data
        except yaml.YAMLError as exc:
            raise HTTPException(status_code=400, detail=f"Invalid YAML: {exc}")
        except JsonSchemaError as exc:
            raise HTTPException(
                status_code=400, detail=f"Schema validation failed: {exc.message}"
            )

    @staticmethod
    def list_emblems(
        supabase: Client,
        category: Optional[str],
        limit: int,
        offset: int,
    ) -> List[Emblem]:
        """Return a paginated, optionally-filtered list of emblems."""
        try:
            query = supabase.table("emblems").select(
                "*, profiles(username, is_verified)"
            )
            if category:
                query = query.eq("category", category)
            response = (
                query.order("downloads_count", desc=True)
                .range(offset, offset + limit - 1)
                .execute()
            )
            return [_row_to_emblem(row) for row in response.data]
        except Exception as exc:
            logger.error("Failed to list emblems: %s", exc)
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def search_emblems(
        supabase: Client,
        q: str,
        category: Optional[str],
        sort: str,
        limit: int,
        offset: int,
    ) -> List[Emblem]:
        """Full-text search for emblems, falling back to ILIKE on RPC errors."""
        try:
            # Try full-text search via RPC first.
            search_query = " & ".join(q.split())
            response = supabase.rpc(
                "search_emblems_fts",
                {
                    "query": search_query,
                    "category_filter": category,
                    "sort_by": sort,
                    "limit_count": limit,
                    "offset_count": offset,
                },
            ).execute()

            return [_row_to_emblem(row) for row in response.data]

        except Exception:
            # Fallback: ILIKE search when the RPC function is not available.
            try:
                query = supabase.table("emblems").select(
                    "*, profiles(username, is_verified)"
                )
                # Escape LIKE metacharacters to prevent pattern injection.
                # PostgREST's `.ilike.` filter passes these escaped characters
                # directly to PostgreSQL's ILIKE operator, which respects the
                # standard SQL escape sequence (ESCAPE '\').
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
                return [_row_to_emblem(row) for row in response.data]
            except Exception as exc:
                logger.error("Failed to search emblems: %s", exc)
                raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def get_emblem(supabase: Client, name: str) -> Emblem:
        """Return a single emblem by name, or raise 404."""
        try:
            response = (
                supabase.table("emblems")
                .select("*, profiles(username, is_verified)")
                .eq("name", name)
                .single()
                .execute()
            )
            if not response.data:
                raise HTTPException(status_code=404, detail="Emblem not found")
            return _row_to_emblem(response.data)
        except HTTPException:
            raise
        except Exception as exc:
            if "not found" in str(exc).lower():
                raise HTTPException(status_code=404, detail="Emblem not found")
            logger.error("Failed to get emblem '%s': %s", name, exc)
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def get_emblem_version(supabase: Client, name: str, version: str) -> dict:
        """Return YAML content for a specific emblem version."""
        try:
            emblem_resp = (
                supabase.table("emblems")
                .select("id")
                .eq("name", name)
                .single()
                .execute()
            )
            if not emblem_resp.data:
                raise HTTPException(status_code=404, detail="Emblem not found")

            emblem_id = emblem_resp.data["id"]
            version_resp = (
                supabase.table("emblem_versions")
                .select("*")
                .eq("emblem_id", emblem_id)
                .eq("version", version)
                .single()
                .execute()
            )
            if not version_resp.data:
                raise HTTPException(status_code=404, detail="Version not found")

            return {
                "name": name,
                "version": version,
                "yaml_content": version_resp.data["yaml_content"],
                "changelog": version_resp.data.get("changelog"),
                "published_at": version_resp.data["published_at"],
            }
        except HTTPException:
            raise
        except Exception as exc:
            logger.error("Failed to get emblem version '%s@%s': %s", name, version, exc)
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def create_emblem(
        supabase: Client,
        request_body: EmblemCreate,
        user: User,
    ) -> Emblem:
        """Create a new emblem and its initial version record."""
        try:
            existing = (
                supabase.table("emblems")
                .select("id")
                .eq("name", request_body.name)
                .execute()
            )
            if existing.data:
                raise HTTPException(
                    status_code=400, detail="Emblem with this name already exists"
                )

            emblem_resp = (
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
            if not emblem_resp.data:
                raise HTTPException(status_code=500, detail="Failed to create emblem")

            emblem_id = emblem_resp.data[0]["id"]

            supabase.table("emblem_versions").insert(
                {
                    "emblem_id": emblem_id,
                    "version": request_body.version,
                    "yaml_content": request_body.yaml_content,
                    "published_by": user.id,
                }
            ).execute()

            supabase.table("emblem_pulls").insert(
                {
                    "emblem_id": emblem_id,
                    "version": request_body.version,
                    "pulled_by": user.id,
                }
            ).execute()

            row = dict(emblem_resp.data[0])
            row["author_name"] = (
                user.username
            )  # inject author_name since there's no profiles join
            row["author_verified"] = None  # will be populated when emblem is fetched
            return _row_to_emblem(row)
        except HTTPException:
            raise
        except Exception as exc:
            logger.error("Failed to create emblem '%s': %s", request_body.name, exc)
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def update_emblem(
        supabase: Client,
        name: str,
        request_body: EmblemUpdate,
        user: User,
    ) -> Emblem:
        """Publish a new version and update the emblem metadata."""
        try:
            emblem_resp = (
                supabase.table("emblems")
                .select("*")
                .eq("name", name)
                .single()
                .execute()
            )
            if not emblem_resp.data:
                raise HTTPException(status_code=404, detail="Emblem not found")

            if emblem_resp.data["author_id"] != user.id:
                raise HTTPException(
                    status_code=403, detail="Not authorized to update this emblem"
                )

            emblem_id = emblem_resp.data["id"]
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
                    or emblem_resp.data["description"],
                    "latest_version": request_body.version,
                }
            ).eq("id", emblem_id).execute()

            row = dict(emblem_resp.data)
            row["author_name"] = user.username  # inject since no profiles join
            row["author_verified"] = None  # will be populated when emblem is fetched
            row["description"] = request_body.description or row["description"]
            row["latest_version"] = request_body.version
            return _row_to_emblem(row)
        except HTTPException:
            raise
        except Exception as exc:
            logger.error("Failed to update emblem '%s': %s", name, exc)
            raise HTTPException(status_code=500, detail="Internal server error")

    @staticmethod
    def delete_emblem(supabase: Client, name: str, user: User) -> dict:
        """Delete an emblem owned by the given user."""
        try:
            emblem_resp = (
                supabase.table("emblems")
                .select("*")
                .eq("name", name)
                .single()
                .execute()
            )
            if not emblem_resp.data:
                raise HTTPException(status_code=404, detail="Emblem not found")

            if emblem_resp.data["author_id"] != user.id:
                raise HTTPException(
                    status_code=403, detail="Not authorized to delete this emblem"
                )

            supabase.table("emblems").delete().eq(
                "id", emblem_resp.data["id"]
            ).execute()

            return {"message": "Emblem deleted successfully"}
        except HTTPException:
            raise
        except Exception as exc:
            logger.error("Failed to delete emblem '%s': %s", name, exc)
            raise HTTPException(status_code=500, detail="Internal server error")
