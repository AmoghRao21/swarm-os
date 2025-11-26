export type AgentRole = "Architect" | "Developer" | "Designer" | "Auditor" | "QA";

export interface SwarmAgent {
  name: string;
  role: AgentRole;
  avatar: string;
}

export interface SwarmCompany {
  id: string;
  name: string;
  description: string;
  priceModel: string;
  specialty: string;
  agents: SwarmAgent[];
  color: "blue" | "pink" | "emerald" | "purple";
}