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

# Execute the Agent via systemd-run with MAXIMUM SECURITY CONSTRAINTS
sudo systemd-run \
    --description="MCP-DLP-Sandbox" \
    --pipe --wait --quiet \
    -E ENABLE_LOCAL_LLM="1" \
    -E LOCAL_LLM_ENDPOINT="http://${TAILSCALE_AI_IP}:11434" \
    -p DynamicUser=yes \
    -p ProtectSystem=strict \
    -p ProtectHome=read-only \
    -p PrivateTmp=yes \
    -p NoNewPrivileges=yes \
    -p RestrictSUIDSGID=true \
    -p ProtectKernelTunables=yes \
    -p ProtectKernelModules=yes \
    -p ProtectKernelLogs=yes \
    -p ProtectControlGroups=yes \
    -p PrivateDevices=yes \
    -p MemoryMax=512M \
    -p IPAddressDeny=any \
    -p IPAddressAllow=127.0.0.0/8 \
    -p IPAddressAllow="${TAILSCALE_AI_IP}" \
    -p CPUQuota=50% \
    "$AGENT_PATH" "$@"