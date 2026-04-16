#!/bin/bash
# Filename: run-sandbox.sh
# Purpose: Encapsulate the mcp-agent within an OS-Level Sandbox (Full Filesystem, Kernel & Tmp Protection)

AGENT_PATH=$(realpath "$1")

if [ ! -f "$AGENT_PATH" ]; then
    echo "Error: File $AGENT_PATH not found" >&2
    exit 1
fi

shift 

# Virtual IP of LLM Server (Tailscale)
TAILSCALE_AI_IP="100.108.61.12"

# Array of additional allowed IPs (e.g. Target VMs for debugging)
EXTRA_ALLOW_IPS=(
    "10.0.0.11"
    "10.0.0.5"
)

# Base arguments for systemd-run MAXIMUM SECURITY CONSTRAINTS
SYSTEMD_ARGS=(
    --description="MCP-DLP-Sandbox"
    --pipe --wait --quiet
    -E ENABLE_LOCAL_LLM="1"
    -E LOCAL_LLM_ENDPOINT="http://${TAILSCALE_AI_IP}:11434"
    -p DynamicUser=yes
    -p ProtectSystem=strict
    -p ProtectHome=read-only
    -p PrivateTmp=yes
    -p NoNewPrivileges=yes
    -p RestrictSUIDSGID=true
    -p ProtectKernelTunables=yes
    -p ProtectKernelModules=yes
    -p ProtectKernelLogs=yes
    -p ProtectControlGroups=yes
    -p PrivateDevices=yes
    -p MemoryMax=512M
    -p IPAddressDeny=any
    -p IPAddressAllow=127.0.0.0/8
    -p "IPAddressAllow=${TAILSCALE_AI_IP}"
    -p CPUQuota=50%
)

# Dynamically append extra allowed IPs
for ip in "${EXTRA_ALLOW_IPS[@]}"; do
    SYSTEMD_ARGS+=( -p "IPAddressAllow=${ip}" )
done

# Execute the Agent
sudo systemd-run "${SYSTEMD_ARGS[@]}" "$AGENT_PATH" "$@"