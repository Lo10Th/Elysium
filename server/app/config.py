from pydantic_settings import BaseSettings
from typing import List, Optional


class Settings(BaseSettings):
    # App settings
    APP_NAME: str = "Elysium Registry"
    APP_VERSION: str = "1.0.0"
    DEBUG: bool = False

    # Server settings (for local dev only)
    HOST: str = "0.0.0.0"
    PORT: int = 8000

    # Supabase credentials (support both old and new naming)
    SUPABASE_URL: str

    # New naming (preferred)
    SUPABASE_KEY: Optional[str] = None
    SUPABASE_SERVICE_ROLE_KEY: Optional[str] = None

    # Old naming (for backward compatibility)
    SUPABASE_ANON_KEY: Optional[str] = None
    SUPABASE_SERVICE_KEY: Optional[str] = None

    # CORS
    CORS_ORIGINS: List[str] = ["*"]

    # Rate limiting
    RATE_LIMIT_REQUESTS: int = 100
    RATE_LIMIT_WINDOW: int = 60  # seconds

    # Custom domain (optional, for CORS)
    DOMAIN: Optional[str] = None

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = True

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

        # Support both old and new naming conventions
        # New naming takes precedence
        if not self.SUPABASE_KEY and self.SUPABASE_ANON_KEY:
            self.SUPABASE_KEY = self.SUPABASE_ANON_KEY

        if not self.SUPABASE_SERVICE_ROLE_KEY and self.SUPABASE_SERVICE_KEY:
            self.SUPABASE_SERVICE_ROLE_KEY = self.SUPABASE_SERVICE_KEY

        # Validate that we have the required keys
        if not self.SUPABASE_KEY:
            raise ValueError(
                "SUPABASE_KEY (or SUPABASE_ANON_KEY) is required. "
                "Set it in Vercel environment variables."
            )

        if not self.SUPABASE_SERVICE_ROLE_KEY:
            raise ValueError(
                "SUPABASE_SERVICE_ROLE_KEY (or SUPABASE_SERVICE_KEY) is required. "
                "Set it in Vercel environment variables."
            )


settings = Settings()

# If custom domain is set, use it for CORS
if settings.DOMAIN:
    settings.CORS_ORIGINS = [f"https://{settings.DOMAIN}"]
