# AgentAPI 启动 Claude Code 的内部逻辑分析

## 如述

当你执行 `agentapi server claude` 时，AgentAPI 内部会启动一个 PTY（伪终端）进程来运行 Claude Code CLI。本文档详细分析这个启动流程。

## 命令解析流程

```mermaid
flowchart TD
    A[用户输入: agentapi server claude] --> B[Cobra 解析参数]
    B --> C{args = cobra.MinimumNArgs 1}
    C --> D[args0 = claude 即 agent 名称]
    D --> E[parseAgentType 解析 agent 类型]
    E --> F{返回 AgentTypeClaude}
    F --> G[切换工作目录]
    G --> H{选择传输方式}
    H --> I{PTY 模式}
    H --> J{ACP 模式}
    I --> K[SetupProcess 启动进程]
    J --> L[SetupACP 启动进程]
    K --> M[termexec.StartProcess]
    L --> N[exec.CommandContext]
    M --> O[运行 claude 命令]
    N --> O
```

## 核心代码流程

### 1. 参数解析

**文件**: `cmd/server/server.go`

```go
// 命令定义
serverCmd := &cobra.Command{
    Use:  "server [agent]",
    Args: cobra.MinimumNArgs(1),  // 至少需要 1 个参数
    Run: func(cmd *cobra.Command, args []string) {
        // args[0] = "claude" (agent 名称)
        // args[1:] = 额外参数 (在 -- 之后)
        runServer(ctx, logger, cmd.Flags().Args())
    },
}
```

### 2. Agent 类型识别

**文件**: `cmd/server/server.go`

```go
func runServer(ctx context.Context, logger *slog.Logger, argsToPass []string) error {
    agent := argsToPass[0]  // "claude"
    agentTypeValue := viper.GetString(FlagType)  // --type 参数
    agentType, err := parseAgentType(agent, agentTypeValue)
    // agentType = AgentTypeClaude
}

// agent 类型别名映射
var agentTypeAliases = map[string]AgentType{
    "claude":       AgentTypeClaude,
    "goose":        AgentTypeGoose,
    // ...
}
```

### 3. 工作目录切换

**文件**: `cmd/server/server.go`

```go
projectDir := viper.GetString(FlagProjectDir)  // --project-dir 或 -d 参数
if projectDir != "" {
    absPath, err := filepath.Abs(projectDir)
    // 如果目录不存在, 自动创建
    if _, err := os.Stat(absPath); os.IsNotExist(err) {
        os.MkdirAll(absPath, 0755)
    }
    os.Chdir(absPath)  // 切换工作目录
}
```

### 4. 传输方式选择

**文件**: `cmd/server/server.go`

```go
experimentalACP := viper.GetBool(FlagExperimentalACP)

if experimentalACP {
    // ACP 模式 - 使用 Agent Communication Protocol
    acpResult, err = httpapi.SetupACP(ctx, httpapi.SetupACPConfig{
        Program:     agent,        // "claude"
        ProgramArgs: argsToPass[1:],  // 额外参数
    })
} else {
    // PTY 模式 - 使用伪终端 (默认)
    proc, err := httpapi.SetupProcess(ctx, httpapi.SetupProcessConfig{
        Program:        agent,           // "claude"
        ProgramArgs:    argsToPass[1:],  // 额外参数
        TerminalWidth:  termWidth,       // 80 (默认)
        TerminalHeight: termHeight,      // 1000 (默认)
        AgentType:      agentType,       // AgentTypeClaude
    })
}
```

## 实际执行的命令

### PTY 模式 (默认)

**文件**: `lib/httpapi/setup.go`

```go
func SetupProcess(ctx context.Context, config SetupProcessConfig) (*termexec.Process, error) {
    // 日志输出
    logger.Info(fmt.Sprintf("Running: %s %s", config.Program, strings.Join(config.ProgramArgs, " ")))

    // 启动 PTY 进程
    process, err := termexec.StartProcess(ctx, termexec.StartProcessConfig{
        Program:        config.Program,        // "claude"
        Args:           config.ProgramArgs,   // []
        TerminalWidth:  config.TerminalWidth, // 80
        TerminalHeight: config.TerminalHeight,// 1000
    })
}
```

**最终执行**:
```bash
claude
```

### ACP 模式 (实验性)

**文件**: `lib/httpapi/setup.go`

```go
func SetupACP(ctx context.Context, config SetupACPConfig) (*SetupACPResult, error) {
    args := config.ProgramArgs
    logger.Info(fmt.Sprintf("Running (ACP): %s %s", config.Program, strings.Join(args, " ")))

    cmd := exec.CommandContext(ctx, config.Program, args...)
    // 启动子进程
    cmd.Start()
}
```

## Claude Code 接收的参数

### 直接传递的参数

当使用 `--` 分隔符时, 后面的参数会直接传给 Claude Code:

```powershell
agentapi server claude -- --dangerously-skip-permissions --verbose
```

这会执行:
```bash
claude --dangerously-skip-permissions --verbose
```

### AgentAPI 处理的参数

| 参数 | 说明 | 对 Claude 的影响 |
|------|------|-----------------|
| `-p, --port` | API 服务端口 | 无 (AgentAPI 自己处理) |
| `-d, --project-dir` | 工作目录 | 通过 `os.Chdir()` 影响 Claude 的运行目录 |
| `-W, --term-width` | 终端宽度 | 传递给 PTY, 影响 Claude 的显示宽度 |
| `-H, --term-height` | 终端高度 | 传递给 PTY, 影响 Claude 的显示高度 |
| `-I, --initial-prompt` | 初始提示词 | 启动后发送给 Claude |
| `-t, --type` | Agent 类型覆盖 | 无 (已经通过位置参数指定) |

## 数据流

```mermaid
sequenceDiagram
    participant User as 用户
    participant API as AgentAPI Server
    participant PTY as 伪终端
    participant Claude as Claude Code CLI

    User->>API: POST /message {content: "帮我分析项目"}
    API->>PTY: 写入消息到 PTY
    PTY->>Claude: 发送键盘输入
    Claude->>PTY: 终端输出
    PTY->>API: 读取屏幕内容
    API->>User: SSE 事件推送 (屏幕更新)
```

## 终端模拟

AgentAPI 使用 PTY (伪终端) 来模拟真实的终端环境:

**文件**: `lib/termexec/process.go`

```go
type StartProcessConfig struct {
    Program        string   // "claude"
    Args           []string // 额外参数
    TerminalWidth  uint16   // 80 (默认)
    TerminalHeight uint16   // 1000 (默认)
}
```

这样 Claude Code 会认为自己运行在一个真实的终端中, 支持:
- 颜色输出
- 光标移动
- 屏幕刷新
- 交互式输入

## 完整启动示例

```powershell
# 基本启动
agentapi server claude
# 内部执行: claude (在当前目录的 PTY 中)

# 指定工作目录
agentapi server claude -d D:\MyProject
# 1. os.Chdir("D:\MyProject")
# 2. 内部执行: claude (在新目录的 PTY 中)

# 传递参数给 Claude
agentapi server claude -- --dangerously-skip-permissions
# 内部执行: claude --dangerously-skip-permissions

# 组合使用
agentapi server claude -d D:\MyProject -p 3285 -- --verbose
# 1. os.Chdir("D:\MyProject")
# 2. 在端口 3285 启动 API 服务
# 3. 内部执行: claude --verbose
```

## 关键文件

| 文件 | 职责 |
|------|------|
| `cmd/server/server.go` | 命令行解析、参数处理、主流程控制 |
| `lib/httpapi/setup.go` | 进程启动逻辑 (PTY/ACP) |
| `lib/termexec/process.go` | PTY 进程管理 |
| `lib/httpapi/server.go` | HTTP API 服务器 |
| `lib/screentracker/conversation.go` | 对话管理和屏幕跟踪 |

## 相关链接

- [AgentAPI GitHub](https://github.com/coder/agentapi)
- [PTY (伪终端) 详解](https://en.wikipedia.org/wiki/Pseudoterminal)
