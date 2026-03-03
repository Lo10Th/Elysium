from slowapi import Limiter
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)

# Rate limit constants
PUBLIC_LIMIT = "60/minute"
AUTH_LIMIT = "30/minute"
STRICT_LIMIT = "10/minute"
REGISTER_LIMIT = "5/minute"
REFRESH_LIMIT = "20/minute"
