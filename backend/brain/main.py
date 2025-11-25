from fastapi import FastAPI
from contextlib import asynccontextmanager
import uvicorn
import redis.asyncio as redis
import os

redis_client = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global redis_client
    redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")
    redis_client = redis.from_url(redis_url, encoding="utf-8", decode_responses=True)
    yield
    await redis_client.close()

app = FastAPI(title="SwarmOS Brain", version="1.0.0", lifespan=lifespan)

@app.get("/health")
async def health_check():
    return {"status": "operational", "service": "swarm-brain"}

@app.post("/agent/dispatch")
async def dispatch_agent(payload: dict):
    return {"job_id": "pending_implementation", "status": "queued"}

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)