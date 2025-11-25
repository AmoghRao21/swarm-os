from typing import TypedDict, List, Annotated
import operator

class AgentState(TypedDict):
    # The original user request
    task: str
    
    # The conversation history (append-only)
    messages: Annotated[List[str], operator.add]
    
    # The current plan/steps
    plan: List[str]
    
    # The code generated so far
    current_code: str
    
    # Errors encountered during execution
    errors: List[str]
    
    # Status of the entire workflow
    status: str