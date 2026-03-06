-- Fix table-level permissions for anonymous users
-- Required for OAuth device flow and public pull tracking
-- 
-- Issue: RLS policies allow operations but postgres roles need explicit table permissions
-- See: https://github.com/Lo10Th/Elysium/issues/126

-- ============================================================================
-- DEVICE_CODES
-- ============================================================================

-- Anonymous users need to create device codes (OAuth device flow)
-- This is safe because RLS policies already restrict operations
GRANT INSERT ON public.device_codes TO anon;
GRANT SELECT ON public.device_codes TO anon;
GRANT UPDATE ON public.device_codes TO anon;

-- Explicit grants for authenticated role (for clarity)
GRANT INSERT ON public.device_codes TO authenticated;
GRANT SELECT ON public.device_codes TO authenticated;
GRANT UPDATE ON public.device_codes TO authenticated;

-- ============================================================================
-- EMBLEM_PULLS
-- ============================================================================

-- Anonymous users can track downloads (for public API analytics)
GRANT INSERT ON public.emblem_pulls TO anon;
GRANT SELECT ON public.emblem_pulls TO authenticated;

-- ============================================================================
-- SEQUENCES
-- ============================================================================

-- Grant USAGE on sequences for UUID generation
-- Required for gen_random_uuid() in id columns
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO anon;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO authenticated;

-- ============================================================================
-- SECURITY IMPROVEMENTS
-- ============================================================================

-- Drop overly permissive UPDATE policy on device_codes
-- Currently: "Anyone can verify a device code" allows UPDATE on ANY row
-- Risk: Malicious actor could modify already-verified codes or tokens
DROP POLICY IF EXISTS "Anyone can verify a device code" ON public.device_codes;

-- Replace with restrictive policy: only allow UPDATE on unverified codes
-- This prevents tampering with codes that have already been verified
CREATE POLICY "Anyone can update unverified device codes"
  ON public.device_codes FOR UPDATE
  USING (verified_at IS NULL)
  WITH CHECK (verified_at IS NULL);

-- ============================================================================
-- NOTES
-- ============================================================================
-- 
-- Why not DELETE permissions?
-- - Device codes should only be deleted by the cleanup function (SECURITY DEFINER)
-- - Prevents accidental or malicious deletion by users
-- - Cleanup function: cleanup_expired_device_codes() runs as table owner
--
-- Security model:
-- - INSERT: Anyone (anon) - required for OAuth device flow
-- - SELECT: Authenticated users via auth.uid() = user_id RLS policy
-- - UPDATE: Only unverified codes - prevents token theft
-- - DELETE: Only via cleanup function - controlled lifecycle