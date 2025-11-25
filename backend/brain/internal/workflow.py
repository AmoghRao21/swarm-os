from langgraph.graph import StateGraph, END
from internal.state import AgentState
import logging

logger = logging.getLogger("swarm.workflow")

# --- Nodes ---

def node_planner(state: AgentState):
    """The Architect: Breaks the task into a plan."""
    logger.info("🤔 Architect is planning...")
    return {
        "plan": ["Analyze requirements", "Write Code", "Review"],
        "messages": ["Plan created."]
    }

def node_coder(state: AgentState):
    """The Worker: Writes the code."""
    logger.info("👨‍💻 Coder is writing...")
    return {
        "current_code": "print('Hello from SwarmOS')",
        "messages": ["Code generation complete."]
    }

def node_reviewer(state: AgentState):
    """The QA: Checks the code."""
    logger.info("🔍 Reviewer is checking...")
    # Simple logic: If code exists, approve it.
    if state.get("current_code"):
        return {"status": "approved", "messages": ["Code looks good."]}
    return {"status": "rejected", "messages": ["No code found."]}

# --- Graph Definition ---

def build_graph():
    workflow = StateGraph(AgentState)

    # 1. Add Nodes
    workflow.add_node("planner", node_planner)
    workflow.add_node("coder", node_coder)
    workflow.add_node("reviewer", node_reviewer)

    # 2. Define Edges (The Logic Flow)
    # Start -> Planner -> Coder -> Reviewer -> End
    workflow.set_entry_point("planner")
    workflow.add_edge("planner", "coder")
    workflow.add_edge("coder", "reviewer")
    workflow.add_edge("reviewer", END)

    return workflow.compile()

runner = build_graph()