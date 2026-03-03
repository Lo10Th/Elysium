"""
Vercel serverless function entry point - DEBUG VERSION
"""

import json


def handler(event, context):
    """Simple test handler for debugging Vercel deployment"""
    return {
        "statusCode": 200,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps(
            {
                "status": "healthy",
                "message": "Debug handler works!",
                "event_type": type(event).__name__,
            }
        ),
    }
