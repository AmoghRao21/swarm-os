import logging
from internal.workflow import runner
from internal.broadcaster import broadcaster

logger = logging.getLogger("swarm.brain")

class AgentOrchestrator:
    """
    Manages the lifecycle of AI agents using LangGraph.
    """
    async def process_job(self, job_data: dict):
        job_id = job_data.get("job_id", "unknown")
        task = job_data.get("task", "")
        
        logger.info(f"🧠 Brain received Job [{job_id}]. Invoking Swarm...")

        # 1. Notify Core that we have started
        await broadcaster.broadcast_job_update(job_id, "processing")

        # Initialize the State
        initial_state = {
            "task": task,
            "messages": [],
            "plan": [],
            "current_code": "",
            "errors": [],
            "status": "started"
        }

        # Run the Graph (Invoke is synchronous in this version, wrapping in future if needed)
        try:
            final_state = await runner.ainvoke(initial_state)
            
            status = final_state.get("status", "completed")
            logger.info(f"✅ Job [{job_id}] finished. Final Status: {status}")
            
            # 2. Notify Core of the result
            await broadcaster.broadcast_job_update(job_id, "completed", final_state)
            return final_state
            
        except Exception as e:
            logger.error(f"❌ Graph execution failed: {e}")
            await broadcaster.broadcast_job_update(job_id, "failed", {"error": str(e)})
            return None

orchestrator = AgentOrchestrator()