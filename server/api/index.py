"""
Vercel serverless function entry point.
"""

import os
import sys
import json
from pathlib import Path

# Ensure server directory is in path for imports
server_dir = Path(__file__).parent.parent
if str(server_dir) not in sys.path:
    sys.path.insert(0, str(server_dir))

try:
    from mangum import Mangum
    from app.main import app

    handler = Mangum(app, lifespan="off")
except Exception as e:
    # Fallback handler that reports errors
    def handler(event, context):
        return {
            "statusCode": 500,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps(
                {
                    "error": "Failed to initialize app",
                    "detail": str(e),
                    "type": type(e).__name__,
                }
            ),
        }
