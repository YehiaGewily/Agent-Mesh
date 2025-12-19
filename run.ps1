Write-Host "Starting AgentMesh..."
docker-compose up --build -d
if ($LASTEXITCODE -eq 0) {
    Write-Host "AgentMesh is live at http://localhost:5173"
    Start-Process "http://localhost:5173"
} else {
    Write-Host "Failed to start services." -ForegroundColor Red
}
