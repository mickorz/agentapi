@echo off
REM AgentAPI Restart Script (Windows)
REM Usage: restart.bat [agent]

set AGENT=%1
if "%AGENT%"=="" set AGENT=claude

echo Stopping existing agentapi processes...
taskkill /F /IM agentapi.exe 2>nul

echo Starting agentapi server with agent: %AGENT%...
REM Unset CLAUDECODE to allow nested session
set CLAUDECODE=
start "" out\agentapi.exe server %AGENT%

echo Server started on http://localhost:3284
echo Chat UI: http://localhost:3284/chat
