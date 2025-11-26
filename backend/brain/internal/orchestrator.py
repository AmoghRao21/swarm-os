import logging
from internal.workflow import runner
from internal.broadcaster import broadcaster

logger = logging.getLogger("swarm.brain")

class AgentOrchestrator:
    async def process_job(self, job_data: dict):
        job_id = job_data.get("job_id", "unknown")
        task = job_data.get("task", "")
        swarm_id = job_data.get("swarm_id", "default") 
        
        logger.info(f"🧠 Brain received Job [{job_id}] for Swarm [{swarm_id}]. Invoking...")

        await broadcaster.broadcast_job_update(job_id, "processing")

        initial_state = {
            "swarm_id": swarm_id,
            "task": task,
            "messages": [],
            "plan": [],
            "current_code": "",
            "errors": [],
            "status": "started"
        }

        try:
            final_state = await runner.ainvoke(initial_state)
            status = final_state.get("status", "completed")
            logger.info(f"✅ Job [{job_id}] finished.")
            
            await broadcaster.broadcast_job_update(job_id, "completed", final_state)
            return final_state
            
        except Exception as e:
            logger.error(f"❌ Graph execution failed: {e}")
            await broadcaster.broadcast_job_update(job_id, "failed", {"error": str(e)})
            return None

orchestrator = AgentOrchestrator()