import asyncio
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI
from internal.consumer import RedisConsumer

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("swarm.main")

consumer = RedisConsumer()

@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("🚀 Swarm Brain booting up...")
    await consumer.connect()
    
    task = asyncio.create_task(consumer.listen())
    
    yield

    logger.info("🛑 Swarm Brain shutting down...")
    await consumer.close()
    task.cancel()

app = FastAPI(title="SwarmOS Brain", version="1.0.0", lifespan=lifespan)

@app.get("/health")
async def health_check():
    return {"status": "operational", "service": "swarm-brain"}