from supabase import create_client, Client
from app.config import settings

# Initialize Supabase client
# Uses SUPABASE_KEY (supports both old and new naming via Settings class)
supabase: Client = create_client(
    settings.SUPABASE_URL,
    settings.SUPABASE_KEY,  # This is set from either SUPABASE_KEY or SUPABASE_ANON_KEY
)
