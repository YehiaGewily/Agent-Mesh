#!/bin/bash
set -e

echo "Starting AgentMesh..."
sudo docker compose up --build -d

if [ $? -eq 0 ]; then
    echo "AgentMesh is live at http://localhost:5173"
    
    # Try to open the browser
    if command -v xdg-open > /dev/null; then
        xdg-open "http://localhost:5173"
    elif command -v open > /dev/null; then
        open "http://localhost:5173"
    else
        echo "Please open http://localhost:5173 in your browser."
    fi
else
    echo -e "\033[0;31mFailed to start services.\033[0m"
    exit 1
fi
