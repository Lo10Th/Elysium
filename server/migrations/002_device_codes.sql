-- Device Code Flow for CLI Authentication
-- Stores device codes for browser-based CLI login

CREATE TABLE IF NOT EXISTS public.device_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  device_code TEXT UNIQUE NOT NULL,
  user_code TEXT UNIQUE NOT NULL,
  user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
  access_token TEXT,
  refresh_token TEXT,
  client_name TEXT DEFAULT 'Elysium CLI',
  verified_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_device_codes_device_code ON public.device_codes(device_code);
CREATE INDEX IF NOT EXISTS idx_device_codes_user_code ON public.device_codes(user_code);
CREATE INDEX IF NOT EXISTS idx_device_codes_user_id ON public.device_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_device_codes_expires_at ON public.device_codes(expires_at);

-- Enable RLS
ALTER TABLE public.device_codes ENABLE ROW LEVEL SECURITY;

-- Policies: Anyone can insert (for device code creation), only verified user can read their codes
CREATE POLICY "Anyone can insert device codes"
  ON public.device_codes FOR INSERT
  WITH CHECK (true);

CREATE POLICY "Users can view their own device codes"
  ON public.device_codes FOR SELECT
  USING (auth.uid() = user_id);

CREATE POLICY "Anyone can verify a device code"
  ON public.device_codes FOR UPDATE
  USING (true);

-- Function to clean up expired device codes (run periodically)
CREATE OR REPLACE FUNCTION public.cleanup_expired_device_codes()
RETURNS void AS $$
BEGIN
  DELETE FROM public.device_codes WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;