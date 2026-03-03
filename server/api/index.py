"""
Vercel serverless function entry point.
This file is required for Vercel to serve the FastAPI application.
"""

# Set environment variables for Vercel serverless environment
import os
os.environ.setdefault('PORT', '8000')

# Import the FastAPI app
from app.main import app

# Vercel expects the app instance to be named 'handler' or 'app'
# The @vercel/python builder will use this as the WSGI/ASGI app
handler = app