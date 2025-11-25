import asyncio
import json
import logging
import redis.asyncio as redis
from internal.config import settings
from internal.orchestrator import orchestrator

logger = logging.getLogger("swarm.listener")

class RedisConsumer:
    def __init__(self):
        self.redis: redis.Redis | None = None
        self.pubsub = None

    async def connect(self):
        self.redis = redis.from_url(settings.REDIS_URL, encoding="utf-8", decode_responses=True)
        self.pubsub = self.redis.pubsub()
        await self.pubsub.subscribe(settings.QUEUE_NAME)
        logger.info(f"👂 Connected to Redis. Listening on channel: {settings.QUEUE_NAME}")

    async def listen(self):
        """
        Infinite loop to process incoming messages.
        """
        if not self.pubsub:
            await self.connect()

        try:
            async for message in self.pubsub.listen():
                if message["type"] == "message":
                    raw_data = message["data"]
                    try:
                        payload = json.loads(raw_data)
                        asyncio.create_task(orchestrator.process_job(payload))
                    except json.JSONDecodeError:
                        logger.error(f"❌ Failed to decode message: {raw_data}")
        except Exception as e:
            logger.critical(f"🔥 Redis listener crashed: {e}")
            
    async def close(self):
        if self.redis:
            await self.redis.close()