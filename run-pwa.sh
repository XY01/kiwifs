#!/bin/bash
# KiwiFS PWA Launcher (HTTP + HTTPS)
cd ~/Documents/GitHub/kiwifs

# Kill existing
pkill -f "kiwifs serve" 2>/dev/null
pkill -f "kiwifs mcp" 2>/dev/null
pkill -f "node pwa-server" 2>/dev/null
pkill -f "node https-proxy" 2>/dev/null
sleep 1

# Start KiwiFS backend server (port 3334)
./kiwifs serve --root ./knowledge --port 3334 --host 0.0.0.0 \
  --space family=./knowledge-family \
  --space projects=./knowledge-projects \
  --space research=./knowledge-research \
  --space agent=./agent &

# Start KiwiFS MCP (port 3007, streamable HTTP for AI agents)
./kiwifs mcp --remote http://127.0.0.1:3334 --space agent --http --port 3007 &

# Start HTTPS proxy (port 3007 -> backend 3334)
node https-proxy.js &

sleep 2
echo "Services running:"
echo "  Backend API:  http://0.0.0.0:3334"
echo "  MCP endpoint: http://0.0.0.0:3007/mcp"
echo "  HTTPS proxy:  https://0.0.0.0:3007/"