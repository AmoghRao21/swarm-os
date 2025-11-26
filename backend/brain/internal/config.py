import os
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    REDIS_URL: str = "redis://localhost:6379/0"
    DATABASE_URL: str = "postgresql://swarm_admin:secure_dev_password@localhost:5432/swarm_os"
    QUEUE_NAME: str = "job_queue"
    
    GROQ_API_KEY: str

    MODEL_ARCHITECT: str = "llama-3.3-70b-versatile"
    MODEL_CODER_BACKEND: str = "llama-3.3-70b-versatile" 
    MODEL_CODER_FRONTEND: str = "llama-3.3-70b-versatile" 
    
    class Config:
        case_sensitive = True

try:
    settings = Settings()
    key_sample = settings.GROQ_API_KEY[:4] + "..." + settings.GROQ_API_KEY[-4:]
    print(f"✅ CONFIG LOADED. AI Engine Active. Key: {key_sample}")
except Exception as e:
    print("❌ CRITICAL: Missing Environment Variables.")
    print(e)
    raise e