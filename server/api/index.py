"""
Vercel serverless function entry point.
Uses Mangum adapter to wrap FastAPI for AWS Lambda/Vercel runtime.
"""

import sys
from pathlib import Path

# Ensure server directory is in path for imports
server_dir = Path(__file__).parent.parent
if str(server_dir) not in sys.path:
    sys.path.insert(0, str(server_dir))

from mangum import Mangum
from app.main import app

handler = Mangum(app, lifespan="off")
