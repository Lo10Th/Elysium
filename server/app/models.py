import re
from datetime import datetime
from typing import List, Optional, Any
from pydantic import BaseModel, EmailStr, Field, constr, field_validator
from enum import Enum

# ---------------------------------------------------------------------------
# Shared validators
# ---------------------------------------------------------------------------

_USERNAME_RE = re.compile(r"^[a-zA-Z0-9_-]{3,30}$")


class AuthType(str, Enum):
    none = "none"
    api_key = "api_key"
    bearer = "bearer"
    basic = "basic"
    oauth2 = "oauth2"


class AuthConfig(BaseModel):
    type: AuthType
    keyEnv: Optional[str] = None
    header: Optional[str] = None
    prefix: Optional[str] = None


class ParameterLocation(str, Enum):
    query = "query"
    path = "path"
    header = "header"
    body = "body"


class Parameter(BaseModel):
    name: str
    type: str
    in_: ParameterLocation = Field(..., alias="in")
    required: bool = False
    description: Optional[str] = None
    default: Optional[Any] = None
    enum: Optional[List[Any]] = None

    class Config:
        populate_by_name = True


class Response(BaseModel):
    description: str
    schema_: Optional[dict] = Field(None, alias="schema")

    class Config:
        populate_by_name = True


class Action(BaseModel):
    description: str
    method: str
    path: str
    parameters: Optional[List[Parameter]] = None
    requestBody: Optional[dict] = None
    responses: Optional[dict[str, Response]] = None
    errors: Optional[List[dict]] = None


class TypeDefinition(BaseModel):
    description: Optional[str] = None
    properties: dict


class EmblemYAML(BaseModel):
    apiVersion: str
    name: constr(pattern=r"^[a-z0-9][a-z0-9-]*[a-z0-9]$", min_length=1, max_length=64)
    version: constr(
        pattern=r"^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$"
    )
    description: constr(min_length=10, max_length=500)
    author: Optional[str] = None
    license: str = "MIT"
    repository: Optional[str] = None
    homepage: Optional[str] = None
    baseUrl: str
    auth: Optional[AuthConfig] = None
    tags: Optional[List[constr(max_length=50)]] = None
    category: Optional[str] = None
    types: Optional[dict[str, TypeDefinition]] = None
    actions: dict[str, Action]


class EmblemCreate(BaseModel):
    name: constr(pattern=r"^[a-z0-9][a-z0-9-]*[a-z0-9]$", min_length=1, max_length=64)
    description: constr(min_length=10, max_length=500)
    yaml_content: str
    category: Optional[str] = None
    tags: Optional[List[str]] = None
    license: str = "MIT"
    repository_url: Optional[str] = None
    homepage_url: Optional[str] = None
    version: str = "1.0.0"


class EmblemUpdate(BaseModel):
    description: Optional[constr(min_length=10, max_length=500)] = None
    yaml_content: str
    version: str


class Emblem(BaseModel):
    id: str
    name: str
    description: str
    author_id: Optional[str] = None
    author_name: Optional[str] = None
    author_verified: Optional[bool] = None
    category: Optional[str] = None
    tags: Optional[List[str]] = None
    license: str = "MIT"
    repository_url: Optional[str] = None
    homepage_url: Optional[str] = None
    latest_version: Optional[str] = None
    downloads_count: int = 0
    created_at: datetime
    updated_at: datetime
    security_advisory: Optional[str] = None
    security_severity: Optional[str] = None


class EmblemVersion(BaseModel):
    id: str
    emblem_id: str
    version: str
    yaml_content: str
    changelog: Optional[str] = None
    published_by: Optional[str] = None
    published_at: datetime


class EmblemWithVersion(Emblem):
    yaml_content: str


class User(BaseModel):
    id: str
    email: str
    username: Optional[str] = None


class SearchQuery(BaseModel):
    q: str
    category: Optional[str] = None
    sort: Optional[str] = "downloads"
    limit: int = 20
    offset: int = 0


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


# ---------------------------------------------------------------------------
# Auth request / response models
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
    is_verified: bool = False
    created_at: str
    updated_at: str


class AuthResponse(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str = "bearer"
    user: User


class OAuthStartRequest(BaseModel):
    redirect_uri: str


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
