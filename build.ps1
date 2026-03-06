# AgentAPI Build Script (Windows PowerShell)
# Usage: .\build.ps1

Write-Host "=== Step 1: Install frontend dependencies ===" -ForegroundColor Green
Set-Location chat
bun install

Write-Host "=== Step 2: Build frontend Chat UI ===" -ForegroundColor Green
$env:NEXT_PUBLIC_BASE_PATH = "/magic-base-path-placeholder"
bun run build

Write-Host "=== Step 3: Copy frontend build output ===" -ForegroundColor Green
Set-Location ..

# Next.js 15 outputs to .next/server/app/ instead of out/
$nextServerApp = "chat/.next/server/app"
if (Test-Path $nextServerApp) {
    Remove-Item -Recurse -Force lib\httpapi\chat -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Path lib\httpapi\chat -Force | Out-Null
    # Remove 404 directory to avoid path issues on Windows
    Remove-Item -Recurse -Force "chat\.next\server\app\404" -ErrorAction SilentlyContinue
    Copy-Item -Recurse -Force "chat\.next\server\app\*" lib\httpapi\chat\
} else {
    Write-Host "Next.js output not found at $nextServerApp, trying chat/out/" -ForegroundColor Yellow
    Remove-Item -Recurse -Force lib\httpapi\chat -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Path lib\httpapi\chat -Force | Out-Null
    # Remove 404 directory to avoid path issues on Windows
    Remove-Item -Recurse -Force "chat\out\404" -ErrorAction SilentlyContinue
    Copy-Item -Recurse -Force "chat\out\*" lib\httpapi\chat\
}

Write-Host "=== Step 4: Build Go project ===" -ForegroundColor Green
$env:GOPROXY = "https://goproxy.cn,direct"
$env:CGO_ENABLED = "0"
go build -o out/agentapi.exe main.go

Write-Host ""
Write-Host "[Success] Build completed!" -ForegroundColor Cyan
Write-Host "Output: out/agentapi.exe"
Write-Host ""
Write-Host "To start server: .\restart.bat"
Write-Host "Chat UI: http://localhost:3284/chat"
Get-Item out\agentapi.exe | Select-Object Name, Length, LastWriteTime
