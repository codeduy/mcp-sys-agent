# MCP System Agent

A simple and secure MCP server that allows AI assistants (like Gemini) to read logs and troubleshoot my Linux VMs.

To keep things safe, the agent only runs read-only commands inside a systemd sandbox. Before sending any server logs back to the cloud AI, it uses a local AI model (Ollama) over Tailscale to automatically redact passwords, API keys, and other sensitive data.

## 🏛️ Architecture
```mermaid
graph TD
    subgraph Outside ["Outside VPN"]
        CloudLLM["Cloud LLM<br>(Gemini, OpenAI, etc.)"]
    end

    Client["MCP Client<br>(Claude Desktop, Antigravity, etc.)"]

    subgraph VPN ["Inside Tailscale VPN"]
        Agent["run-sandbox.sh + MCP Server<br>(agent main.go)"]
        OS["OS / Bash"]
        LocalLLM["Local LLM"]

        Agent -->|"(4) Execute read command"| OS
        Agent -->|"(5) Requesting hide sensitive info<br>before send to Cloud LLM"| LocalLLM
        LocalLLM -->|"(6) Send data context<br>that hide sensitive info"| Agent
    end

    %% External and cross-VPN communication flow
    Client -->|"(1) Sending user request"| CloudLLM
    CloudLLM -->|"(2) Request context on VM's MCP server<br>based on user request"| Client
    Client -->|"(3) Request agent execute read command<br>for context collecting"| Agent
    
    Agent -.->|"(6)"| Client
    Client -.->|"(6)"| CloudLLM
    
    CloudLLM -->|"(7) Response recommend actions<br>based on user request"| Client

    %% Background color
    style VPN fill:#f4f8ff,stroke:#333,stroke-width:2px
    style Outside fill:#fff5f5,stroke:#333,stroke-width:2px,stroke-dasharray: 5 5
```



## 🔄 Data Workflow
```mermaid
sequenceDiagram
    participant Client as MCP Client<br>(Claude Desktop, Antigravity, etc.)
    
    box Outside VPN
        participant CloudLLM as Cloud LLM<br>(Gemini, OpenAI, etc.)
    end
    
    box rgb(244, 248, 255) Inside Tailscale VPN
        participant Agent as MCP Server<br>(main.go)
        participant OS as OS / Bash
        participant LocalLLM as Local LLM
    end

    Client->>CloudLLM: (1) Sending user request
    CloudLLM-->>Client: (2) Request context on VM's MCP server
    Client->>Agent: (3) Request agent execute read command
    Agent->>OS: (4) Execute read command
    Note right of Agent: Gather raw logs/configs
    Agent->>LocalLLM: (5) Request hide sensitive info (DLP)
    LocalLLM-->>Agent: (6) Send safe data context
    Agent-->>Client: (6) Return safe context over Tailscale
    Client->>CloudLLM: (6) Forward safe context
    CloudLLM-->>Client: (7) Response recommend actions
```

## 🗺️ Roadmap

```mermaid
%%{init: { 'theme': 'base', 'themeVariables': { 'cScale0': '#3b82f6', 'cScale1': '#8b5cf6', 'cScale2': '#ec4899', 'cScale3': '#f97316' } } }%%
timeline
    Phase 1 (Now)
        : Foundation
        : Core Linux Read-Only Agent
        : Tailscale VPN Integration
        : Basic DLP via Local LLM
        
    Phase 2 (Next)
        : Cloud Native & Cross-Platform
        : Windows Agent Support
        : Kubernetes DaemonSet/Sidecar
        : Cloud Provider APIs (AWS/GCP/Azure)
        : Hybrid DLP Engine
        
    Phase 3 (Later)
        : Automation & Remediation
        : Ansible / Terraform Integration
        : Human-in-the-loop Execution
        : Pre-approved Safe Actions
        
    Phase 4 (Future)
        : Enterprise Fleet Management
        : Fleet Server Control Plane
        : Fleet Proxy for VPCs
        : OpenTelemetry & Proactive Alerts
```
