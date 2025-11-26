from typing import TypedDict, List, Annotated
import operator

class AgentState(TypedDict):
    swarm_id: str 
    task: str
    messages: Annotated[List[str], operator.add]
    plan: List[str]
    current_code: str
    errors: List[str]
    status: str