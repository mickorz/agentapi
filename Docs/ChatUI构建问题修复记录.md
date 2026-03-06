# Chat UI 构建问题修复记录

## 问题描述

打开 Chat UI 页面时，显示原始的 Next.js 构建文件内容，而不是渲染后的 HTML 页面：

```
page.js
page.js.nft.json
page_client-reference-manifest.js
```

## 问题原因

### 根本原因

`build.ps1` 脚本使用了错误的输出目录。Next.js 15 在使用 `output: "export"` 配置时，静态文件输出到 `chat/out/` 目录，而服务器端渲染文件输出到 `.next/server/app/` 目录。

### 错误的构建逻辑

```powershell
# 错误: 优先使用 .next/server/app/ 目录
$nextServerApp = "chat/.next/server/app"
if (Test-Path $nextServerApp) {
    Copy-Item -Recurse -Force "chat\.next\server\app\*" lib\httpapi\chat\
}
```

这个逻辑会优先检查 `.next/server/app/` 是否存在，如果存在就复制该目录的内容。但该目录包含的是服务器端渲染的中间文件，而不是静态 HTML 文件。

### 目录结构对比

**错误目录** `.next/server/app/` (服务器端文件):
```
.next/server/app/
├── page.js                        # 服务器端 JS bundle
├── page.js.nft.json              # Node File Tracing 信息
├── page_client-reference-manifest.js
├── index.html
├── embed.html
├── favicon.ico/                   # 目录形式
└── ...
```

**正确目录** `chat/out/` (静态导出文件):
```
chat/out/
├── index.html                     # 静态 HTML
├── embed.html
├── 404.html
├── favicon.ico                    # 文件形式
├── _next/
│   └── static/
│       ├── chunks/
│       └── css/
└── ...
```

## 解决方案

### 修改 build.ps1

```powershell
# 修复: 使用正确的静态导出目录
$nextOut = "chat/out"
if (Test-Path $nextOut) {
    Remove-Item -Recurse -Force lib\httpapi\chat -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Path lib\httpapi\chat -Force | Out-Null
    # Remove 404 directory to avoid path issues on Windows
    Remove-Item -Recurse -Force "chat\out\404" -ErrorAction SilentlyContinue
    Copy-Item -Recurse -Force "chat\out\*" lib\httpapi\chat\
} else {
    Write-Host "ERROR: Next.js static export output not found at $nextOut" -ForegroundColor Red
    Write-Host "Make sure next.config.ts has output: 'export' and build completed successfully" -ForegroundColor Red
    exit 1
}
```

### 关键修改点

1. **使用 `chat/out/` 目录**: 这是 Next.js `output: "export"` 配置的正确输出位置
2. **添加错误处理**: 如果目录不存在，显示明确的错误信息并退出
3. **移除 fallback 逻辑**: 不再尝试使用 `.next/server/app/` 作为备选

## 技术背景

### Next.js 静态导出

当 `next.config.ts` 配置了 `output: "export"` 时：

```typescript
const nextConfig: NextConfig = {
  output: "export",  // 启用静态导出
  images: { unoptimized: true },
  basePath,
  trailingSlash: true,
};
```

Next.js 会执行以下构建流程：

1. `next build` - 编译应用
2. 自动触发静态导出
3. 将所有页面生成为静态 HTML 文件到 `out/` 目录

### 目录用途说明

| 目录 | 用途 | 是否用于静态服务 |
|------|------|-----------------|
| `chat/out/` | 静态导出输出 | **是** - 包含完整的静态 HTML/JS/CSS |
| `.next/` | 构建缓存和中间产物 | 否 |
| `.next/static/` | 静态资源 (被复制到 out/) | 否 |
| `.next/server/` | 服务器端渲染产物 | 否 |
| `.next/server/app/` | SSR 页面组件 | 否 - 会导致问题 |

## 验证修复

修复后，`lib/httpapi/chat/` 目录结构：

```
lib/httpapi/chat/
├── index.html          # 正确
├── embed/
├── _next/
│   └── static/
├── favicon.ico         # 文件
└── 404.html            # 文件
```

访问 `http://localhost:3284/chat/` 返回正确的 HTML 页面，而不是原始 JS 文件。

## 提交记录

```
commit ee99778
fix: use correct Next.js static export output directory

Changed build.ps1 to copy from chat/out/ instead of .next/server/app/.
When using output: "export" in next.config.ts, Next.js generates static
files to the out/ directory.
```

## 相关文件

- `build.ps1` - Windows 构建脚本
- `build.sh` - Linux/Mac 构建脚本 (本身是正确的)
- `chat/next.config.ts` - Next.js 配置
- `lib/httpapi/embed.go` - 静态文件嵌入逻辑

## 教训总结

1. **理解构建工具的输出目录**: Next.js 在不同模式下输出到不同目录
2. **优先级逻辑要谨慎**: 不要随意使用 fallback 目录，可能导致错误行为
3. **静态导出 vs SSR**: `output: "export"` 和默认 SSR 模式的输出完全不同
