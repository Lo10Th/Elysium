import asyncio
from typing import Any, Callable, Optional
from supabase import create_client, Client
from app.config import get_settings

# Initialize Supabase client lazily
_supabase: Optional[Client] = None
_supabase_service: Optional[Client] = None


def get_supabase() -> Client:
    """Get Supabase client instance (lazy initialization)."""
    global _supabase
    if _supabase is None:
        settings = get_settings()
        _supabase = create_client(
            settings.SUPABASE_URL, settings.effective_supabase_key
        )
    return _supabase


def get_supabase_service_client() -> Client:
    """Get Supabase client with service role key (bypasses RLS).

    Use this for operations that need to bypass Row-Level Security,
    such as device code creation which happens before authentication.
    """
    global _supabase_service
    if _supabase_service is None:
        settings = get_settings()
        _supabase_service = create_client(
            settings.SUPABASE_URL, settings.effective_supabase_service_key
        )
    return _supabase_service


async def run_sync(func: Callable, *args: Any, **kwargs: Any) -> Any:
    """Run a synchronous (blocking) callable in the default thread pool.

    Use this to offload Supabase (and other blocking) calls from async route
    handlers without blocking the event loop.

    Example::

        response = await run_sync(
            supabase.table("emblems").select("*").execute
        )
    """
    return await asyncio.to_thread(func, *args, **kwargs)


# DO NOT access at module level - use get_supabase() in routes
