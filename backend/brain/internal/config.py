from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    REDIS_URL: str = "redis://localhost:6379/0"
    DATABASE_URL: str = "postgresql://swarm_admin:secure_dev_password@localhost:5432/swarm_os"
    QUEUE_NAME: str = "job_queue"
    
    class Config:
        env_file = ".env"

settings = Settings()