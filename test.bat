@echo off
echo Testing AgentAPI Chat UI...
echo.

echo 1. Testing root redirect...
curl -s -o nul -w "   Status: %%{http_code}\n" http://localhost:3284/

echo 2. Testing chat index...
curl -s -o nul -w "   Status: %%{http_code}\n" http://localhost:3284/chat/index.html

echo 3. Testing chat page...
curl -s -o nul -w "   Status: %%{http_code}\n" http://localhost:3284/chat/

echo.
echo Done. Open http://localhost:3284/chat in browser to verify.
pause
