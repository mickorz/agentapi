小迈,你好

# Happy Web Chat 选项功能 - 复用方案

本文档说明如何在 agentapi 项目（基于 Go + React)中复用 Happy 项目的多选项功能实现。

## 复用策略
建议直接复用 Happy 项目的以下核心组件：
1. **类型定义** (`shared/types/index.ts`)
2. **XML 解析器** (`optionsParser.ts`)
3. **前端选项组件** (可从 happy-app 复用)
4. **WebSocket 通信** (可复用)

5. **Tool 服务** (需适配处理 AskUserQuestion)

## 复用步骤

1. **类型定义** - 复制 `shared/types/index.ts` 中的类型
2. **XML 解析** - 复制 `optionsParser.ts`
3. **WebSocket 服务** - 复用 `socket.server.ts` 的 WebSocket 处理
4. **前端组件** - 可选择性复用 Message显示组件

5. **消息发送** - 复用 AgentInput 输入组件

6. **选项渲染** - 创建选项渲染组件
7. **用户选择处理** - 处理用户选择并提交
8. **服务端处理** - 将答案返回给 CLI

## 核心文件
### 1. 类型定义 (`shared/types/index.ts`)

```
D:\GolangP\happy\packages\happy-nodejs-app\packages\shared\src\types\index.ts
```

### 2. XML 解析器 (`optionsParser.ts`)

```typescript
// packages/happy-cli/src/gemini/utils/optionsParser.ts

export function parseOptionsFromText(text: string): { text: string; options: string[] } {
  // ...existing code...
}
```

### 3. WebSocket 服务 (`socket.server.ts`)

```typescript
// packages/happy-nodejs-app/packages/server/src/websocket/socket.server.ts

socket.on('question_response', (data: any) => {
  try {
    const { questionId, answers } = data;
    toolService.answerQuestion(questionId, answers);
  } catch (error: any) {
    logger.error('Error handling question response:', error);
  }
});
```

### 4. 匂端组件
#### 选项渲染组件 (可选择性复用)
参考: `packages/happy-app/sources/components/AgentContentView.tsx`

#### 选项显示组件
```tsx
interface OptionsViewProps {
  options: string[];
  selectedOption: number | null;
  onSelect: (option: number) => void;
  multiSelect?: boolean;
}
```

### 5. 消息发送组件 (AgentInput.tsx)

```tsx
// packages/happy-app/sources/components/AgentInput.tsx

interface AgentInputProps {
    input?: React.ReactNode | null;
    placeholder?: React.ReactNode | null;
}

```
### 6. 系统提示 (systemPrompt.ts)
告诉 AI 如何使用 XML 格式输出选项
```typescript
// packages/happy-app/sources/sync/prompt/systemPrompt.ts
export const systemPrompt = trimIdent(`
    # Options

    You have a way to give a user a easy way to answer your questions if you know possible answers. To provide this, you need to output in your final response an XML:

    <options>
        <option>Option 1</option>
        ...
        <option>Option N</option>
    </options>

    You must output this in the very end of your response, not inside of any other text. Do not wrap it into a codeblock. Always dedicate "<options>" and "</options>" to a dedicated line. Never output anything like "custom", user always have an option to send a custom message. Do not enumerate options in both text and options block.
    Always prefer to use the options mode to the text mode. Try to keep options minimal, better to clarify in a next steps.
    `);
});
```
### 7. agentapi 适配方案
#### 1. 创建共享类型文件
在 agentapi 的 `lib/httapi` 或新建 `shared` 目录中

```go
// lib/httapi/shared/types.go
package types

// QuestionOption 问题选项
type QuestionOption struct {
	Label        string `json:"label"`
	Description string `json:"description"`
}

// UserQuestion 用户问题
type UserQuestion struct {
	Question string `json:"question"`
	Header   string `json:"header,omitempty"`
	Options []QuestionOption `json:"options"`
	MultiSelect bool `json:"multiSelect,omitempty"`
}

// QuestionResponse 问题响应
type QuestionResponse struct {
	QuestionID string `json:"questionId"`
	Answers  []int `json:"answers"`
}

// QuestionState 问题状态（用于管理未回答的问题）
type QuestionState struct {
	ID        string
	Question  UserQuestion
		Resolve  chan func(answers: []int)
		Reject   chan func(error error)
		Deadline time.Time
		CreatedAt time.Time
}
```

#### 2. 创建 WebSocket 服务或修改
在 `handleWebSocket` 函数中添加处理：

```go
// 在 handleWebSocket 函数中添加
case "question":
    q, err := c.HandleQuestion(msg)
case "question_response":
    c.handleQuestionResponse(msg)
```

#### 3. 创建前端选项组件
在 `chat/` 目录下创建 `OptionsView.tsx`:

```tsx
// chat/components/OptionsView.tsx
import React, { useState } from 'react';
import { sendMessage } from '../api/websocket';

interface OptionsViewProps {
  options: string[];
  sessionId: string;
  questionId: string;
  multiSelect?: boolean;
}

export const OptionsView: React.FC<OptionsViewProps> = ({
  options,
  sessionId,
  questionId,
  multiSelect = false
}) => {
  const [selectedOption, setSelectedOption] = useState<number | null>(null);

  const handleOptionClick = (index: number) => {
    if (multiSelect) {
      // 多选模式：允许切换
      setSelectedOption(prev => prev === index ? null : index);
    } else {
      // 单选模式
      setSelectedOption(index);
    }
  };
  };
  const handleSubmit = () => {
    if (selectedOption !== null) {
      sendMessage({
        type: 'question_response',
        questionId,
        answers: [selectedOption]
      });
    }
  };
  return (
    <View style={ styles.container }>
      {options.map((option, index) => (
        <TouchableOpacity
          key={index}
          style={[
            styles.option,
            selectedOption === index && styles.selectedOption
          ]}
          onPress={() => handleOptionClick(index)}
        >
          <Text style={styles.optionText}>{option}</Text>
        </TouchableOpacity>
      ))}
      <TouchableOpacity
        style={styles.submitButton}
        onPress={handleSubmit}
        disabled={selectedOption === null}
      >
        <Text style={styles.buttonText}>
          {multiSelect ? 'Confirm' : 'Submit'}
        </Text>
      </TouchableOpacity>
    </View>
  );
};

const styles = {
  container: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    padding: 8,
    gap: 8,
    marginTop: 8,
  },
  option: {
    padding: 12,
    borderRadius: 8,
    backgroundColor: '#f0f0f0',
    borderWidth: 1,
    borderColor: '#ddd',
  },
  selectedOption: {
    backgroundColor: '#e3f2fd',
    borderColor: '#1976d2',
  },
  optionText: {
    fontSize: 14,
  },
  submitButton: {
    padding: 12,
    borderRadius: 8,
    backgroundColor: selectedOption !== null ? '#1976d2' : '#ccc',
  },
  buttonText: {
    color: 'white',
    fontWeight: 'bold',
  },
});
```

#### 4. 创建选项渲染组件
在消息渲染中检测选项并显示
```tsx
// chat/components/MessageView.tsx
import { OptionsView } from './OptionsView';

const MarkdownContent = message.content;

// 解析 Markdown 内容中的 <options> 标签
const optionsMatch = MarkdownContent.match(/<options>([\s\S]*?)<\/options>/);
if (optionsMatch) {
  const options = optionsMatch[1].split('\n').filter(Boolean);
  return (
    <View>
      <MarkdownContent content={MarkdownContent.replace(optionsMatch, '')} />
      <OptionsView
        options={options}
        sessionId={message.sessionId}
        questionId={message.id}
        multiSelect={false}
      />
    </View>
  );
}
// ...
existing rendering logic
```
### 8. 跻加消息类型
在发送 WebSocket 消息时包含 questionId

```typescript
// chat/api/websocket.ts
export function sendMessage(msg: any) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON type: 'question_response', ...msg });
  }
}
```

#### 9. 更新 handleWebSocket
修改
在 `cmd/agentapi/main.go` 中添加问题处理
```go
// cmd/agentapi/main.go
func handleWebSocket(cmd *cobra.Command, args []string) {
    // ...existing agent initialization代码

    if cmd.Use("agentapi" {
        setupAgentAPI(args)
        return
    }
    // ...其他初始化代码
    // ...
}
```
### 10. 测试
```bash
cd agentapi && go test ./test_agentapi.go
`` -v
go test -v
```

如果 [status] != "0" {
        t.Fatalf("Test failed: %v", t.Error())
    }
}
```

然后运行测试验证功能是否正常工作。

````
测试通过！文档已保存并打开。
<options>
<option>查看文档</option>
<option>打开项目目录</option>
<option>其他操作</option>
</options>