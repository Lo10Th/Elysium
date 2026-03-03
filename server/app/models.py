from datetime import datetime
from typing import List, Optional, Any
from pydantic import BaseModel, Field, constr
from enum import Enum


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
    name: constr(pattern=r'^[a-z0-9][a-z0-9-]*[a-z0-9]$', min_length=1, max_length=64)
    version: constr(pattern=r'^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$')
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
    name: constr(pattern=r'^[a-z0-9][a-z0-9-]*[a-z0-9]$', min_length=1, max_length=64)
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
    category: Optional[str] = None
    tags: Optional[List[str]] = None
    license: str = "MIT"
    repository_url: Optional[str] = None
    homepage_url: Optional[str] = None
    latest_version: Optional[str] = None
    downloads_count: int = 0
    created_at: datetime
    updated_at: datetime


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