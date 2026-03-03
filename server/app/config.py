from pydantic_settings import BaseSettings
from typing import List, Optional
import os


class Settings(BaseSettings):
    # App settings
    APP_NAME: str = "Elysium Registry"
    APP_VERSION: str = "1.0.0"
    DEBUG: bool = False

    # Server settings
    HOST: str = "0.0.0.0"
    PORT: int = 8000

    # Supabase URL (required)
    SUPABASE_URL: str = ""

    # Support BOTH naming conventions
    # These will be populated from env vars automatically
    SUPABASE_KEY: str = ""
    SUPABASE_SERVICE_ROLE_KEY: str = ""
    SUPABASE_ANON_KEY: str = ""
    SUPABASE_SERVICE_KEY: str = ""

    # CORS
    CORS_ORIGINS: List[str] = ["*"]

    # Rate limiting
    RATE_LIMIT_REQUESTS: int = 100
    RATE_LIMIT_WINDOW: int = 60

    # Custom domain
    DOMAIN: Optional[str] = None

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = True

    @property
    def effective_supabase_key(self) -> str:
        """Get the Supabase key from either naming convention."""
        key = self.SUPABASE_KEY or self.SUPABASE_ANON_KEY
        if not key:
            raise ValueError(
                "Set SUPABASE_KEY or SUPABASE_ANON_KEY environment variable"
            )
        return key

    @property
    def effective_supabase_service_key(self) -> str:
        """Get the service key from either naming convention."""
        key = self.SUPABASE_SERVICE_ROLE_KEY or self.SUPABASE_SERVICE_KEY
        if not key:
            raise ValueError(
                "Set SUPABASE_SERVICE_ROLE_KEY or SUPABASE_SERVICE_KEY environment variable"
            )
        return key


# Don't instantiate at module level - let it be lazy
# This prevents import errors if env vars aren't set yet
_settings: Optional[Settings] = None


def get_settings() -> Settings:
    """Get settings instance (lazy initialization)."""
    global _settings
    if _settings is None:
        _settings = Settings()
    return _settings


# For backward compatibility with direct access
settings = get_settings()

# If custom domain is set, use it for CORS
if settings.DOMAIN:
    settings.CORS_ORIGINS = [f"https://{settings.DOMAIN}"]
