# AgentAPI --allowed-origins 参数说明

## 什么是 CORS？

CORS (Cross-Origin Resource Sharing) 是浏览器的安全策略，用于控制一个域名的网页是否可以访问另一个域名的资源。

## --allowed-origins 的作用

用于控制哪些前端网页可以访问你的 agentapi 服务器。

### 为什么需要它？

当你的前端页面（如 `http://localhost:3000`）尝试访问 agentapi 服务器（`http://localhost:3284`）时，浏览器会检查服务器是否允许这个"跨域"请求。

### 场景示例

| 场景 | 命令 |
|------|------|
| 用自带 Chat UI | 不需要设置（默认已包含） |
| 自定义前端 (React/Vue) | 设置你的前端地址 |
| 多个前端应用 | 用逗号分隔多个地址 |

### 使用方法

```powershell
# 允许特定来源
-o "http://localhost:3000,http://localhost:8080"

# 允许所有来源（开发时方便，生产环境不推荐）
-o "*"

# 生产环境
-o "https://myapp.com,https://www.myapp.com"
```

### 默认值

```
http://localhost:3284, http://localhost:3000, http://localhost:3001
```

## 总结

- 如果只是用 agentapi 自带的 Chat UI，**不需要管这个参数**
- 如果开发自定义前端，需要设置你的前端地址
- 生产环境建议设置具体的域名，不要用 `*`

## 相关链接

- [MDN - CORS](https://developer.mozilla.org/zh-CN/docs/Web/HTTP/CORS)
- [CORS 详解](https://www.ruanyifeng.com/blog/2016/cors.html)
