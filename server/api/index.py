"""
Vercel serverless function entry point.
"""

import os
import sys

# Add the parent directory to Python path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Import the FastAPI app
from app.main import app

# Export for Vercel
handler = app
