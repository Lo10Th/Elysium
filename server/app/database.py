from supabase import create_client, Client
from supabase import create_client, Client
from app.config import settings

# Initialize Supabase client with new API key naming
supabase: Client = create_client(
    settings.SUPABASE_URL,
    settings.SUPABASE_KEY  # This is the anon/public key
)