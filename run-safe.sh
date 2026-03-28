#!/bin/bash
# Tên file: run-safe.sh
# Tác dụng: Bọc file mcp-agent vào lồng kính OS-Level (Full Filesystem, Kernel & Tmp Protection)

AGENT_PATH=$(realpath "$1")

if [ ! -f "$AGENT_PATH" ]; then
    echo "Lỗi: Không tìm thấy file $AGENT_PATH" >&2
    exit 1
fi

# Bỏ qua tham số đầu tiên (đã nạp vào AGENT_PATH)
shift 

# Chạy Agent qua systemd-run với FULL CẤP BẢO MẬT
sudo systemd-run \
    --description="MCP-DLP-Sandbox" \
    --pipe --wait --quiet \
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
    -p IPAddressAllow=localhost \
    -p CPUQuota=50% \
    "$AGENT_PATH" "$@"