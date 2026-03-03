-- Elysium Database Migration
-- Run this in Supabase SQL Editor

-- ============================================================================
-- PROFILES TABLE
-- ============================================================================

-- Create profiles table (if not exists)
CREATE TABLE IF NOT EXISTS public.profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  username TEXT UNIQUE NOT NULL,
  email TEXT,
  avatar_url TEXT,
  bio TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;

-- Create index on username for faster lookups
CREATE INDEX IF NOT EXISTS idx_profiles_username ON public.profiles(username);

-- Create index on id for joins
CREATE INDEX IF NOT EXISTS idx_profiles_id ON public.profiles(id);

-- Policies for profiles
CREATE POLICY "Public profiles are viewable by everyone"
  ON public.profiles FOR SELECT
  USING (true);

CREATE POLICY "Users can update own profile"
  ON public.profiles FOR UPDATE
  USING (auth.uid() = id);

CREATE POLICY "Users can insert own profile"
  ON public.profiles FOR INSERT
  WITH CHECK (auth.uid() = id);

-- ============================================================================
-- TRIGGER: Auto-create profile on signup
-- ============================================================================

CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.profiles (id, username, email)
  VALUES (
    NEW.id, 
    COALESCE(NEW.raw_user_meta_data->>'username', split_part(NEW.email, '@', 1)),
    NEW.email
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Drop existing trigger if exists, then create new one
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- ============================================================================
-- UPDATED_AT TRIGGER
-- ============================================================================

CREATE OR REPLACE FUNCTION public.update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS on_profiles_updated ON public.profiles;
CREATE TRIGGER on_profiles_updated
  BEFORE UPDATE ON public.profiles
  FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

-- ============================================================================
-- FULL-TEXT SEARCH FOR EMBLEMS
-- ============================================================================

-- Add search_vector column to emblems (if not exists)
DO $$ 
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_schema = 'public' 
    AND table_name = 'emblems' 
    AND column_name = 'search_vector'
  ) THEN
    ALTER TABLE public.emblems ADD COLUMN search_vector tsvector;
  END IF;
END $$;

-- Create GIN index for full-text search
DROP INDEX IF EXISTS idx_emblems_search;
CREATE INDEX idx_emblems_search ON public.emblems USING GIN(search_vector);

-- Function to update search vector
CREATE OR REPLACE FUNCTION public.emblems_search_trigger() 
RETURNS TRIGGER AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(ARRAY_TO_STRING(NEW.tags, ' '), '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger, create new one
DROP TRIGGER IF EXISTS emblems_search_update ON public.emblems;
CREATE TRIGGER emblems_search_update
  BEFORE INSERT OR UPDATE ON public.emblems
  FOR EACH ROW EXECUTE FUNCTION public.emblems_search_trigger();

-- Update existing emblems with search vectors
UPDATE public.emblems SET search_vector =
  setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
  setweight(to_tsvector('english', COALESCE(description, '')), 'B') ||
  setweight(to_tsvector('english', COALESCE(ARRAY_TO_STRING(tags, ' '), '')), 'C')
WHERE search_vector IS NULL;

-- ============================================================================
-- ENSURE EMBLEMS TABLE HAS PROPER STRUCTURE
-- ============================================================================

-- Ensure author_id references profiles
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints 
    WHERE constraint_name = 'emblems_author_id_fkey' 
    AND table_name = 'emblems'
  ) THEN
    ALTER TABLE public.emblems 
    ADD CONSTRAINT emblems_author_id_fkey 
    FOREIGN KEY (author_id) REFERENCES public.profiles(id) ON DELETE SET NULL;
  END IF;
END $$;

-- Create index on author_id for faster queries
CREATE INDEX IF NOT EXISTS idx_emblems_author_id ON public.emblems(author_id);

-- ============================================================================
-- STORAGE BUCKET FOR AVATARS (Optional)
-- ============================================================================

-- Create avatars bucket if it doesn't exist
INSERT INTO storage.buckets (id, name, public)
VALUES ('avatars', 'avatars', true)
ON CONFLICT (id) DO NOTHING;

-- Policy for avatars bucket
CREATE POLICY "Avatar images are publicly accessible"
  ON storage.objects FOR SELECT
  USING (bucket_id = 'avatars');

CREATE POLICY "Anyone can upload an avatar"
  ON storage.objects FOR INSERT
  WITH CHECK (bucket_id = 'avatars');

CREATE POLICY "Anyone can update their own avatar"
  ON storage.objects FOR UPDATE
  USING (bucket_id = 'avatars');

-- ============================================================================
-- GRANT PERMISSIONS
-- ============================================================================

-- Grant necessary permissions to authenticated and anon users
GRANT USAGE ON SCHEMA public TO authenticated, anon;
GRANT ALL ON public.profiles TO authenticated, anon;
GRANT ALL ON public.emblems TO authenticated, anon;
GRANT ALL ON public.emblem_versions TO authenticated, anon;

-- ============================================================================
-- FULL-TEXT SEARCH RPC FUNCTION
-- ============================================================================

CREATE OR REPLACE FUNCTION public.search_emblems_fts(
  query TEXT,
  category_filter TEXT DEFAULT NULL,
  sort_by TEXT DEFAULT 'downloads',
  limit_count INT DEFAULT 20,
  offset_count INT DEFAULT 0
)
RETURNS TABLE (
  id UUID,
  name TEXT,
  description TEXT,
  author_id UUID,
  author_name TEXT,
  category TEXT,
  tags TEXT[],
  license TEXT,
  repository_url TEXT,
  homepage_url TEXT,
  latest_version TEXT,
  downloads_count INT,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
)
LANGUAGE plpgsql
AS $$
BEGIN
  RETURN QUERY
  SELECT 
    e.id,
    e.name,
    e.description,
    e.author_id,
    p.username AS author_name,
    e.category,
    e.tags,
    e.license,
    e.repository_url,
    e.homepage_url,
    e.latest_version,
    e.downloads_count,
    e.created_at,
    e.updated_at
  FROM public.emblems e
  LEFT JOIN public.profiles p ON e.author_id = p.id
  WHERE 
    e.search_vector @@ to_tsquery('english', query)
    AND (category_filter IS NULL OR e.category = category_filter)
  ORDER BY
    CASE WHEN sort_by = 'downloads' THEN e.downloads_count END DESC,
    CASE WHEN sort_by = 'recent' THEN e.created_at END DESC,
    CASE WHEN sort_by = 'name' THEN 0 END,
    CASE WHEN sort_by = 'name' THEN e.name END ASC,
    CASE WHEN sort_by NOT IN ('downloads', 'recent', 'name') THEN e.downloads_count END DESC
  LIMIT limit_count
  OFFSET offset_count;
END;
$$;

-- Grant execute permission on the function
GRANT EXECUTE ON FUNCTION public.search_emblems_fts(TEXT, TEXT, TEXT, INT, INT) TO authenticated, anon;

-- ============================================================================
-- SAMPLE DATA (Optional - for testing)
-- ============================================================================

-- Uncomment to insert sample data
-- INSERT INTO public.emblems (id, name, description, author_id, category, tags, license, latest_version, downloads_count, created_at, updated_at)
-- SELECT 
--   gen_random_uuid(),
--   'sample-api',
--   'A sample API emblem for demonstration purposes. This showcases the structure of an emblem.',
--   (SELECT id FROM profiles LIMIT 1),
--   'Utilities',
--   ARRAY['sample', 'demo', 'test'],
--   'MIT',
--   '1.0.0',
--   100,
--   NOW(),
--   NOW()
-- WHERE NOT EXISTS (SELECT 1 FROM emblems WHERE name = 'sample-api');