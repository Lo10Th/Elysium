from typing import Optional
from supabase import create_client, Client
from app.config import get_settings

# Initialize Supabase client lazily
_supabase: Optional[Client] = None


def get_supabase() -> Client:
    """Get Supabase client instance (lazy initialization)."""
    global _supabase
    if _supabase is None:
        settings = get_settings()
        _supabase = create_client(
            settings.SUPABASE_URL, settings.effective_supabase_key
        )
    return _supabase


# DO NOT access at module level - use get_supabase() in routes
