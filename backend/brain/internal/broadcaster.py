import json
import logging
import redis.asyncio as redis
from internal.config import settings

logger = logging.getLogger("swarm.broadcaster")

class EventBroadcaster:
    """
    Handles the transmission of signals from the Python Brain back to the Go Core.
    """
    def __init__(self):
        self._redis = None

    @property
    def redis(self):
        # Lazy connection initialization
        if not self._redis:
             self._redis = redis.from_url(settings.REDIS_URL, encoding="utf-8", decode_responses=True)
        return self._redis

    async def broadcast_job_update(self, job_id: str, status: str, result: dict = None):
        """
        Publishes a 'JOB_UPDATE' event to the Redis Bus.
        """
        channel = "job_updates"
        payload = {
            "job_id": job_id,
            "status": status,
            "result": result or {},
            "timestamp": "now" # In real app, use ISO format
        }
        
        try:
            # Fire and forget
            await self.redis.publish(channel, json.dumps(payload))
            logger.info(f"📡 Broadcasted update for Job [{job_id}] -> Status: {status}")
        except Exception as e:
            logger.error(f"❌ Failed to broadcast event: {e}")

broadcaster = EventBroadcaster()