"use client";

import { useEffect, useState, useRef } from "react";
import { useParams } from "next/navigation";
import Editor from "@monaco-editor/react";
import styles from "./page.module.css";

interface MissionData {
  job_id: string;
  status: "queued" | "processing" | "completed" | "failed";
  data: {
    task?: string;
    plan?: string[];
    messages?: string[];
    current_code?: string;
    errors?: string[];
  };
}

// --- Helper: Strip Markdown Fences ---
const cleanCode = (code: string) => {
  if (!code) return "";
  // Remove ```python at start and ``` at end
  return code.replace(/^```\w*\n/, "").replace(/```$/, "");
};

// --- Typewriter Hook ---
const useTypewriter = (targetText: string, speed = 5) => {
  const [displayedText, setDisplayedText] = useState("");
  const indexRef = useRef(0);

  useEffect(() => {
    if (!targetText) return;
    
    // Reset if target changes significantly (new file)
    if (targetText.length < displayedText.length) {
        setDisplayedText("");
        indexRef.current = 0;
    }

    const interval = setInterval(() => {
      if (indexRef.current < targetText.length) {
        // Add a chunk of characters at a time for speed
        const chunkSize = 5; 
        const nextChunk = targetText.slice(0, indexRef.current + chunkSize);
        setDisplayedText(nextChunk);
        indexRef.current += chunkSize;
      } else {
        clearInterval(interval);
      }
    }, speed);

    return () => clearInterval(interval);
  }, [targetText, speed]);

  return displayedText;
};

export default function MissionControl() {
  const params = useParams();
  const id = params?.id as string;

  const [mission, setMission] = useState<MissionData>({
    job_id: id,
    status: "queued",
    data: { messages: [], plan: [], current_code: "" }
  });
  
  const [isConnected, setIsConnected] = useState(false);
  const terminalRef = useRef<HTMLDivElement>(null);
  
  // We clean the code BEFORE passing it to the typewriter
  const rawCode = mission.data.current_code || "";
  const cleanedCode = cleanCode(rawCode);
  
  // This hook generates the "typing" animation
  const animatedCode = useTypewriter(cleanedCode, 1);

  // 1. Initial Hydration
  useEffect(() => {
    if (!id) return;

    const fetchState = async () => {
        try {
            const res = await fetch(`/api/v1/job/${id}`);
            if (res.ok) {
                const data = await res.json();
                console.log("📥 State Hydrated:", data);
                setMission(prev => ({
                    ...prev,
                    status: data.status,
                    data: { ...prev.data, ...data.data }
                }));
            }
        } catch (e) {
            console.error("Hydration Failed:", e);
        }
    };
    fetchState();
  }, [id]);

  // 2. Real-time WebSocket
  useEffect(() => {
    if (!id) return;

    const ws = new WebSocket("ws://localhost:8080/api/v1/ws");

    ws.onopen = () => {
      console.log("🔌 Uplink Established");
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.type === "JOB_UPDATE" && payload.job_id === id) {
            setMission(prev => ({
                ...prev,
                status: payload.status,
                data: { ...prev.data, ...payload.data }
            }));
        }
      } catch (e) {
        console.error("Telemetry Error:", e);
      }
    };

    ws.onclose = () => setIsConnected(false);
    return () => ws.close();
  }, [id]);

  // Auto-scroll logs
  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [mission.data.messages]);

  return (
    <div className={styles.container}>
      {/* Header */}
      <header className={styles.header}>
        <div>
          <h1 className={styles.missionTitle}>Mission Control</h1>
          <div className={styles.missionId}>ID: {id}</div>
        </div>
        <div className={styles.statusPanel}>
          <span className={`${styles.statusBadge} ${styles[`status_${mission.status}`]}`}>
            {mission.status}
          </span>
          <span style={{ marginLeft: 10, fontSize: 12, color: isConnected ? '#10b981' : '#666' }}>
            {isConnected ? "● LIVE FEED" : "○ DISCONNECTED"}
          </span>
        </div>
      </header>

      <div className={styles.dashboard}>
        
        {/* Panel 1: The Plan */}
        <div className={`${styles.panel} ${styles.planPanel}`}>
          <div className={styles.panelHeader}>
            <span>🗺️ Strategic Plan</span>
          </div>
          <div className={styles.stepList}>
            {mission.data.plan && mission.data.plan.length > 0 ? (
                mission.data.plan.map((step, i) => (
                <div key={i} className={styles.step}>
                    <div className={styles.checkbox}>✓</div>
                    <span>{step}</span>
                </div>
                ))
            ) : (
                <div style={{ color: '#444', padding: 10, fontSize: '0.8rem' }}>
                    Waiting for Architect...
                </div>
            )}
          </div>
        </div>

        {/* Panel 2: Logs */}
        <div className={`${styles.panel} ${styles.logsPanel}`}>
          <div className={styles.panelHeader}>
            <span>📟 Swarm Telemetry</span>
          </div>
          <div className={styles.terminal} ref={terminalRef}>
            {mission.data.messages?.map((msg, i) => (
              <div key={i} className={styles.logEntry}>
                <span className={styles.timestamp}>[{new Date().toLocaleTimeString()}]</span>
                <span className={styles.logMessage}>{msg}</span>
              </div>
            ))}
            {mission.status === "queued" && (
                <div className={styles.logEntry}>
                    <span className={styles.timestamp}>[SYSTEM]</span>
                    <span className={styles.logMessage}>Connection established. Waiting for swarm...</span>
                </div>
            )}
          </div>
        </div>

        {/* Panel 3: Code Editor */}
        <div className={`${styles.panel} ${styles.codePanel}`}>
          <div className={styles.panelHeader}>
            <span>📝 Artifact Output</span>
            <span style={{ fontSize: '0.7rem', color: '#666' }}>READ-WRITE ACCESS</span>
          </div>
          {/* FIX: Use the dedicated CSS class instead of Tailwind classes */}
          <div className={styles.editorWrapper}>
             <Editor
                height="100%"
                defaultLanguage="python"
                theme="vs-dark"
                // Use animatedCode to simulate typing even if job is done
                value={animatedCode} 
                options={{
                    minimap: { enabled: false },
                    fontSize: 14,
                    fontFamily: "JetBrains Mono",
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                    padding: { top: 16 },
                    wordWrap: "on"
                }}
             />
          </div>
        </div>

      </div>
    </div>
  );
}