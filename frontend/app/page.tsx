"use client";

import { useState } from "react";
import { useRouter } from "next/navigation"; // <--- Import Router
import { SwarmCompany } from "@/types";
import styles from "./page.module.css";

// ... SWARMS array stays the same ...
// ... Types stay the same ...
const SWARMS: SwarmCompany[] = [
  {
    id: "ironclad",
    name: "IronClad Backend",
    description: "High-performance systems. We build APIs that survive 10k RPS.",
    specialty: "Go / Python / SQL",
    priceModel: "$0.05 / step",
    color: "blue",
    agents: [
      { name: "Atlas", role: "Architect", avatar: "🏛️" },
      { name: "Forge", role: "Developer", avatar: "🔨" },
    ],
  },
  {
    id: "pixel",
    name: "Pixel Perfect Studios",
    description: "Fluid UIs. We treat DOM manipulation as a fine art.",
    specialty: "React / Motion",
    priceModel: "$0.08 / step",
    color: "pink",
    agents: [
      { name: "Venus", role: "Designer", avatar: "🎨" },
      { name: "Flash", role: "Developer", avatar: "⚡" },
    ],
  },
  {
    id: "solid",
    name: "SolidBlock Web3",
    description: "Smart contracts with zero vulnerabilities. Gas optimized.",
    specialty: "Solidity / Rust",
    priceModel: "$0.15 / step",
    color: "emerald",
    agents: [
      { name: "Argus", role: "Auditor", avatar: "👁️" },
      { name: "Chain", role: "QA", avatar: "⛓️" },
    ],
  },
  {
    id: "godmode",
    name: "GodMode Inc.",
    description: "The elite full-stack agency. Autonomous recursive improvement.",
    specialty: "Polyglot / AI",
    priceModel: "$0.50 / step",
    color: "purple",
    agents: [
      { name: "Zeus", role: "Architect", avatar: "⚡" },
      { name: "Athena", role: "QA", avatar: "🛡️" },
      { name: "Vulcan", role: "Developer", avatar: "🔥" },
    ],
  },
];

export default function Marketplace() {
  const router = useRouter(); // <--- Initialize Hook
  const [selected, setSelected] = useState<string | null>(null);
  const [mission, setMission] = useState("");
  const [isDeploying, setIsDeploying] = useState(false);

  const handleHire = (id: string) => {
    setSelected(id);
  };

  const closeModal = () => {
    setSelected(null);
    setMission("");
    setIsDeploying(false);
  };

  const handleDeploy = async () => {
    if (!mission.trim()) return;

    setIsDeploying(true);
    
    try {
      // UPDATE: Send swarm_id along with the task
      const response = await fetch("/api/v1/job", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
            task: mission,
            swarm_id: selected // <--- Vital: Passing the Company ID
        }),
      });

      if (response.ok) {
        const data = await response.json();
        router.push(`/mission/${data.job_id}`);
      } else {
        console.error("❌ Deployment Failed");
        alert("System Error: Unable to contact Swarm Core.");
        setIsDeploying(false);
      }
    } catch (error) {
      console.error("Network Error:", error);
      setIsDeploying(false);
    }
  };

  // ... Render ...
  return (
    <main className={styles.main}>
      <div className={styles.container}>
        {/* Header */}
        <header className={styles.header}>
          <div>
            <h1 className={styles.title}>
              Swarm<span className={styles.brandHighlight}>OS</span>
            </h1>
            <p className={styles.subtitle}>Autonomous Enterprise Workforce Platform</p>
          </div>
          <div className={styles.statusPanel}>
            <div className={styles.systemOnline}>
              <div className={styles.ping}></div> SYSTEM ONLINE
            </div>
            <div style={{ color: '#444', fontSize: '0.8rem' }}>v1.0.0-alpha</div>
          </div>
        </header>

        {/* Grid */}
        <div className={styles.grid}>
          {SWARMS.map((swarm) => (
            <div
              key={swarm.id}
              className={`${styles.card} ${selected === swarm.id ? styles.selected : ''}`}
              data-color={swarm.color}
              onClick={() => handleHire(swarm.id)}
            >
              <div className={styles.cardHeader}>
                <div>
                  <h3 className={styles.companyName}>{swarm.name}</h3>
                  <span className={styles.tag} style={{ color: `var(--accent-${swarm.color})` }}>
                    {swarm.specialty}
                  </span>
                </div>
                <div className={styles.price}>{swarm.priceModel}</div>
              </div>

              <p className={styles.description}>{swarm.description}</p>

              <div className={styles.footer}>
                <span className={styles.label}>Team</span>
                <div className={styles.agentStack}>
                  {swarm.agents.map((agent, i) => (
                    <div key={i} className={styles.agent} title={agent.role}>
                      {agent.avatar}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Briefing Modal */}
      {selected && (
        <div className={styles.modalOverlay}>
          <div className={styles.modalContent}>
            <div className={styles.modalHeader}>
              <div className={styles.modalTitle}>Briefing: {SWARMS.find(s => s.id === selected)?.name}</div>
              <button className={styles.modalClose} onClick={closeModal}>&times;</button>
            </div>
            
            <label className={styles.inputLabel}>Mission Objective</label>
            <textarea
              className={styles.missionInput}
              placeholder="Describe the software you need built... (e.g. 'Create a Python script to scrape stock prices')"
              value={mission}
              onChange={(e) => setMission(e.target.value)}
              autoFocus
            />
            
            <button 
              className={styles.deployButton} 
              onClick={handleDeploy}
              disabled={isDeploying || !mission.trim()}
            >
              {isDeploying ? "Initializing Swarm..." : "Deploy Swarm"}
            </button>
          </div>
        </div>
      )}
    </main>
  );
}