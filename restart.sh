#!/bin/bash
# AgentAPI Restart Script (Unix/Linux/macOS)
# Usage: ./restart.sh [agent]

AGENT=${1:-claude}

echo "Stopping existing agentapi processes..."
pkill -f agentapi 2>/dev/null || true

echo "Starting agentapi server with agent: $AGENT..."
./out/agentapi server $AGENT &

echo "Server started on http://localhost:3284"
echo "Chat UI: http://localhost:3284/chat"
