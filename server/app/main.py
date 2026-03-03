from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from slowapi import Limiter
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded
from app.config import get_settings
from app.routes import auth, emblems, keys
import logging

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)

_settings = get_settings()

limiter = Limiter(key_func=get_remote_address)

app = FastAPI(
    title=_settings.APP_NAME,
    version=_settings.APP_VERSION,
    description="Registry for API emblems - discover and use APIs programmatically",
)

app.state.limiter = limiter

app.add_middleware(
    CORSMiddleware,
    allow_origins=_settings.CORS_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.exception_handler(RateLimitExceeded)
async def rate_limit_handler(request: Request, exc: RateLimitExceeded):
    from fastapi.responses import JSONResponse

    return JSONResponse(
        status_code=429,
        content={"error": "Rate limit exceeded", "detail": str(exc.detail)},
    )


@app.get("/health")
async def health():
    return {"status": "healthy", "version": _settings.APP_VERSION}


@app.get("/")
async def root():
    return {
        "name": _settings.APP_NAME,
        "version": _settings.APP_VERSION,
        "docs": "/docs",
        "health": "/health",
    }


app.include_router(auth.router, prefix="/api/auth", tags=["Authentication"])
app.include_router(emblems.router, prefix="/api/emblems", tags=["Emblems"])
app.include_router(keys.router, prefix="/api/keys", tags=["API Keys"])


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host=_settings.HOST, port=_settings.PORT)
