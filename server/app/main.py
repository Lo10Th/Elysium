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

limiter = Limiter(key_func=get_remote_address)

app = FastAPI(
    title="Elysium Registry",
    version="0.1.0",
    description="Registry for API emblems - discover and use APIs programmatically",
)

app.state.limiter = limiter

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
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
    return {"status": "healthy", "version": "0.1.0"}


@app.get("/")
async def root():
    return {
        "name": "Elysium Registry",
        "version": "1.0.0",
        "docs": "/docs",
        "health": "/health",
    }


app.include_router(auth.router, prefix="/api/auth", tags=["Authentication"])
app.include_router(emblems.router, prefix="/api/emblems", tags=["Emblems"])
app.include_router(keys.router, prefix="/api/keys", tags=["API Keys"])


if __name__ == "__main__":
    import uvicorn

    settings = get_settings()
    uvicorn.run(app, host=settings.HOST, port=settings.PORT)
