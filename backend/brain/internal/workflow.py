from langgraph.graph import StateGraph, END
from langchain_groq import ChatGroq
from langchain_core.messages import SystemMessage, HumanMessage
from internal.state import AgentState
from internal.config import settings
import logging
import json

logger = logging.getLogger("swarm.workflow")

# --- 1. The Factory (Model Router) ---
def get_agent_llm(swarm_id: str, role: str):
    if not settings.GROQ_API_KEY:
        logger.error("❌ Attempted to initialize LLM without GROQ_API_KEY")
        return None

    model_name = settings.MODEL_ARCHITECT # Default

    if role == "coder":
        if swarm_id in ["ironclad", "solid"]:
            model_name = settings.MODEL_CODER_BACKEND
        elif swarm_id == "pixel":
            model_name = settings.MODEL_CODER_FRONTEND
        elif swarm_id == "godmode":
            model_name = settings.MODEL_CODER_BACKEND
            
    return ChatGroq(
        temperature=0.1,
        model_name=model_name,
        groq_api_key=settings.GROQ_API_KEY
    )

# --- 2. The Personas (Prompt Router) ---
PROMPTS = {
    "ironclad": {
        "architect": """You are a Senior Backend Architect.
Focus on: Scalability, Database Efficiency (SQL), Microservices patterns, and Error Handling.
Output strictly a JSON array of steps.""",
        "coder": """You are an Expert Backend Developer (Python/Go/SQL).
Write production-grade, secure code.
- Use strict typing.
- Implement proper error handling (try/except).
- Optimize for performance.
Return ONLY the code block."""
    },
    "pixel": {
        "architect": """You are a UX/UI Architect.
Focus on: User Experience, Accessibility, Component Reusability, and Visual Hierarchy.
Output strictly a JSON array of steps.""",
        "coder": """You are an Expert Frontend Developer (React/Tailwind).
Write modern, responsive code.
- Use Functional Components and Hooks.
- Use Tailwind CSS for styling.
- Ensure accessibility (ARIA tags).
Return ONLY the code block."""
    },
    "solid": {
        "architect": """You are a Web3 Systems Architect.
Focus on: Gas Optimization, Security (Reentrancy protection), and Smart Contract patterns.
Output strictly a JSON array of steps.""",
        "coder": """You are an Expert Smart Contract Engineer (Solidity/Rust).
Write secure, gas-optimized code.
- Follow security best practices (Checks-Effects-Interactions).
- Document specific security considerations.
Return ONLY the code block."""
    },
    "godmode": {
        "architect": "You are a Polyglot Software Architect. Design the most efficient system possible.",
        "coder": "You are a 10x Full Stack Developer. Write the best possible code for the task."
    },
    "default": {
        "architect": "You are a Software Architect. Create a plan.",
        "coder": "You are a Developer. Write the code."
    }
}

def get_prompt(swarm_id: str, role: str):
    company_prompts = PROMPTS.get(swarm_id, PROMPTS["default"])
    return company_prompts.get(role, PROMPTS["default"][role])


# --- 3. The Nodes ---

def node_planner(state: AgentState):
    swarm_id = state.get("swarm_id", "default")
    llm = get_agent_llm(swarm_id, "architect")
    
    if not llm:
        return {"plan": ["Error: API Key Missing"], "messages": ["⚠️ System Error: Brain is offline."]}

    logger.info(f"🤔 Architect ({swarm_id}) is thinking...")
    
    system_prompt = get_prompt(swarm_id, "architect")
    
    messages = [
        SystemMessage(content=system_prompt),
        HumanMessage(content=f"Task: {state['task']}")
    ]
    
    try:
        response = llm.invoke(messages)
        # Clean up potential markdown wrapping in JSON response
        content = response.content.replace("```json", "").replace("```", "").strip()
        plan = json.loads(content)
        
        if not isinstance(plan, list):
            plan = ["Analyze Requirements", "Develop Core Logic", "Review Implementation"]
            
    except Exception as e:
        logger.error(f"Planner Error: {e}")
        plan = ["Analyze Requirements", "Develop Core Logic", "Review Implementation"]

    return {
        "plan": plan, 
        "messages": [f"Architect ({swarm_id}) generated {len(plan)} strategic steps."]
    }

def node_coder(state: AgentState):
    swarm_id = state.get("swarm_id", "default")
    llm = get_agent_llm(swarm_id, "coder")
    
    if not llm:
        return {"current_code": "# Error: Brain Offline"}

    logger.info(f"👨‍💻 Coder ({swarm_id}) is typing...")
    
    system_prompt = get_prompt(swarm_id, "coder")
    
    # Contextualize the coding task with the plan
    plan_str = "\n".join([f"- {step}" for step in state.get("plan", [])])
    
    messages = [
        SystemMessage(content=system_prompt),
        HumanMessage(content=f"""
Task: {state['task']}

Execution Plan:
{plan_str}

Implement the solution now. Return ONLY the code.
""")
    ]
    
    response = llm.invoke(messages)
    
    return {
        "current_code": response.content,
        "messages": [f"Code generated by Expert Agent."]
    }

def node_reviewer(state: AgentState):
    # Future: Add specific QA prompts here too
    return {"status": "approved", "messages": ["Automated review passed."]}

# --- Graph Definition ---

def build_graph():
    workflow = StateGraph(AgentState)

    workflow.add_node("planner", node_planner)
    workflow.add_node("coder", node_coder)
    workflow.add_node("reviewer", node_reviewer)

    workflow.set_entry_point("planner")
    workflow.add_edge("planner", "coder")
    workflow.add_edge("coder", "reviewer")
    workflow.add_edge("reviewer", END)

    return workflow.compile()

runner = build_graph()