# AgentAPI HTTP API 使用文档

## 概述

AgentAPI 提供了一个 RESTful HTTP API，允许你与 AI 编程助手进行交互。所有 API 端点都在 `http://localhost:3284` 下可用。

## 如览

| 端点 | 方法 | 说明 |
|------|------|------|
| `/status` | GET | 获取服务器状态 |
| `/messages` | GET | 获取对话历史 |
| `/message` | POST | 发送消息给 Agent |
| `/upload` | POST | 上传文件 |
| `/events` | GET (SSE) | 订阅事件流 |
| `/internal/screen` | GET (SSE) | 订阅屏幕更新 |

---

## 详细端点说明

### 1. 获取状态 - GET /status

获取当前服务器的运行状态。

**请求**
```
GET /status
```

**响应**
```json
{
  "status": "running | stable",
  "agent_type": "claude | goose | aider | ...",
  "transport": "pty | acp"
}
```

**字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| `status` | string | Agent 当前状态。`running` 表示正在处理消息， `stable` 表示空闲等待输入 |
| `agent_type` | string | 正在使用的 Agent 类型 |
| `transport` | string | 后端传输方式 (`acp` 或 `pty`) |

**示例**
```powershell
curl http://localhost:3284/status
```

---

### 2. 获取消息历史 - GET /messages

获取完整的对话历史记录。

**请求**
```
GET /messages
```

**响应**
```json
{
  "messages": [
    {
      "id": 0,
      "content": "Hello, how can I help?",
      "role": "assistant",
      "time": "2026-03-06T10:00:00Z"
    },
    {
      "id": 1,
      "content": "帮我分析这个项目",
      "role": "user",
      "time": "2026-03-06T10:00:05Z"
    }
  ]
}
```

**消息字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | 消息唯一标识符，也表示消息在历史中的顺序 |
| `content` | string | 消息内容（格式化为终端显示格式，每行80字符） |
| `role` | string | 消息作者角色 (`user` / `assistant` / `system`) |
| `time` | datetime | 消息时间戳 |

**示例**
```powershell
curl http://localhost:3284/messages
```

---

### 3. 发送消息 - POST /message

向 Agent 发送消息。

**请求**
```
POST /message
Content-Type: application/json

{
  "content": "帮我分析这个项目",
  "type": "user"
}
```

**请求体字段**

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `content` | string | 是 | 消息内容 |
| `type` | string | 是 | 消息类型 |

**消息类型**

| 类型 | 说明 |
|------|------|
| `user` | 用户消息，会被记录到对话历史并提交给 Agent。 AgentAPI 会等待 Agent 开始执行任务后才响应 |
| `raw` | 原始消息，直接写入 Agent 的终端会话作为按键，不会保存到对话历史。用于发送转义序列等 |

**响应**
```json
{
  "ok": true
}
```

**示例**
```powershell
# 发送用户消息
curl -X POST http://localhost:3284/message \
  -H "Content-Type: application/json" \
  -d '{"content": "帮我分析这个项目", "type": "user"}'

# 发送原始按键 (如 Ctrl+C)
curl -X POST http://localhost:3284/message \
  -H "Content-Type: application/json" \
  -d '{"content": "\u0003", "type": "raw"}'
```

---

### 4. 上传文件 - POST /upload

上传文件到服务器，**注意**: 使用 `multipart/form-data` 格式。

**请求**
```
POST /upload
Content-Type: multipart/form-data

file: <文件内容>
```

**限制**
- 最大文件大小: 10MB
- 文件保存在临时目录中

**响应**
```json
{
  "ok": true,
  "filePath": "/tmp/agentapi-uploads-xxx/abc123/filename.txt"
}
```

**示例**
```powershell
curl -X POST http://localhost:3284/upload \
  -F "file=@test.txt"
```

---

### 5. 订阅事件流 - GET /events (SSE)

通过 Server-Sent Events (SSE) 订阅实时事件流。

**请求**
```
GET /events
Accept: text/event-stream
```

**事件类型**

| 事件类型 | 说明 |
|---------|------|
| `message` | 新消息事件 |
| `status` | 状态变更事件 |

**示例**
```powershell
curl -N -H "Accept: text/event-stream" http://localhost:3284/events
```

**JavaScript 示例**
```javascript
const eventSource = new EventSource('http://localhost:3284/events');
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data);
};
```

---

### 6. 订阅屏幕更新 - GET /internal/screen (SSE)

订阅 Agent 终端屏幕的实时更新。这是一个内部 API。

**请求**
```
GET /internal/screen
Accept: text/event-stream
```

**用途**
- 实时监控 Agent 终端输出
- 调试和开发用途

---

## 数据类型

### AgentStatus

```go
type AgentStatus string

const (
    AgentStatusRunning AgentStatus = "running"
    AgentStatusStable  AgentStatus = "stable"
)
```

### Transport

```go
type Transport string

const (
    TransportPTY Transport = "pty"  // 伪终端
    TransportACP Transport = "acp"  // Agent Communication Protocol
)
```

### MessageType

```go
type MessageType string

const (
    MessageTypeUser MessageType = "user"  // 用户消息
    MessageTypeRaw  MessageType = "raw"  // 原始按键
)
```

---

## CORS 配置

默认允许的来源:
- `http://localhost:3284`
- `http://localhost:3000`
- `http://localhost:3001`

可通过 `--allowed-origins` 参数或 `AGENTAPI_ALLOWED_ORIGINS` 环境变量修改。

---

## 错误处理

API 使用标准 HTTP 状态码:

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求无效 |
| 500 | 服务器内部错误 |

---

## 完整使用示例

### PowerShell 示例

```powershell
# 获取状态
$status = Invoke-RestMethod -Uri "http://localhost:3284/status"
Write-Host "Agent Status: $($status.status)"

# 发送消息
$body = @{
    content = "帮我分析这个项目"
    type = "user"
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:3284/message" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body

# 获取消息历史
$messages = Invoke-RestMethod -Uri "http://localhost:3284/messages"
$messages.messages | ForEach-Object { Write-Host "$($_.role): $($_.content)" }
```

### Python 示例
```python
import requests
import json

BASE_URL = "http://localhost:3284"

# 获取状态
response = requests.get(f"{BASE_URL}/status")
print(f"Status: {response.json()}")

# 发送消息
response = requests.post(
    f"{BASE_URL}/message",
    json={"content": "帮我分析这个项目", "type": "user"}
)
print(f"Sent: {response.json()}")

# 获取消息历史
response = requests.get(f"{BASE_URL}/messages")
for msg in response.json()["messages"]:
    print(f"{msg['role']}: {msg['content']}")
```

### JavaScript 示例
```javascript
const BASE_URL = 'http://localhost:3284';

// 获取状态
async function getStatus() {
  const response = await fetch(`${BASE_URL}/status`);
  return response.json();
}

// 发送消息
async function sendMessage(content, type = 'user') {
  const response = await fetch(`${BASE_URL}/message`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content, type })
  });
  return response.json();
}

// 订阅事件流
function subscribeToEvents(onEvent) {
  const eventSource = new EventSource(`${BASE_URL}/events`);
  eventSource.onmessage = (event) => {
    onEvent(JSON.parse(event.data));
  };
  return eventSource;
}
```

---

## 相关链接

- [AgentAPI GitHub](https://github.com/coder/agentapi)
- [OpenAPI Schema](http://localhost:3284/openapi.json) (服务器运行时可用)
