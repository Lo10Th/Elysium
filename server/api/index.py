"""
Vercel serverless function entry point.
This file is required for Vercel to serve the FastAPI application.
"""

import os

# Set default port for Vercel
os.environ.setdefault("PORT", "8000")

# Import the FastAPI app
from app.main import app

# For Vercel Python runtime, we need to export the app directly
# The runtime will detect it's a FastAPI/Starlette app and handle it as ASGI
# Do not rename this variable - Vercel expects 'app' or 'handler'
handler = app
