-- ============================================================================
-- VERIFIED ACCOUNTS
-- ============================================================================

-- Add is_verified column to profiles table
ALTER TABLE public.profiles ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT FALSE;

-- Create index for faster queries on verified users
CREATE INDEX IF NOT EXISTS idx_profiles_is_verified ON public.profiles(is_verified) WHERE is_verified = TRUE;

-- ============================================================================
-- UPDATE SEARCH FUNCTION TO INCLUDE AUTHOR_VERIFIED
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
  author_verified BOOLEAN,
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
    p.is_verified AS author_verified,
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

-- Grant execute permission
GRANT EXECUTE ON FUNCTION public.search_emblems_fts(TEXT, TEXT, TEXT, INT, INT) TO authenticated, anon;