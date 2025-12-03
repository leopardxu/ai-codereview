# Eino Gerrit Review Agent

这是一个基于 [Eino](https://github.com/cloudwego/eino) 框架开发的自动化代码评审代理服务。它结合了静态代码分析和大语言模型（LLM），为 Gerrit 上的代码变更提供智能评审建议。

## 功能特性

- **双模分析**：结合静态规则扫描和 LLM 智能分析。
- **上下文感知**：自动获取变更代码的相关上下文（函数、类、依赖等）以提高分析准确性。
- **Gerrit 集成**：直接读取 Gerrit 变更，并将评审意见回写到 Gerrit。
- **规则热重载**：支持动态更新静态分析规则，无需重启服务。
- **高性能**：并行数据获取与优化的处理流程。

## 快速开始

### 前置要求

- Go 1.21+
- Gerrit 账号（用于拉取代码和发表评论）
- OpenAI API Key（或兼容的 LLM 服务，如 DeepSeek, Moonshot 等）

### 1. 编译

```bash
go build -o review-agent cmd/server/main.go
```

### 2. 配置环境变量

在运行前，请设置以下环境变量：

| 变量名 | 必填 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `PORT` | 否 | `8000` | 服务监听端口 |
| `GERRIT_BASE_URL` | **是** | - | Gerrit 服务器地址 (如 `https://gerrit.example.com`) |
| `GERRIT_USER` | **是** | - | Gerrit 用户名 |
| `GERRIT_TOKEN` | **是** | - | Gerrit HTTP Password / Token |
| `OPENAI_API_KEY` | **是** | - | LLM API Key |
| `OPENAI_BASE_URL` | 否 | `https://api.openai.com/v1` | LLM API Base URL |
| `MODEL_NAME` | 否 | `gpt-4o` | 使用的模型名称 |
| `RULE_CONFIG_PATH` | 否 | - | 静态规则配置文件路径 (JSON) |
| `CONTEXT_FILE_LIMIT` | 否 | `10` | 上下文文件大小限制 (KB) |
| `CONTEXT_GRANULARITY`| 否 | `file` | 上下文粒度 (`file`, `function`, `class`, `dependency`) |

### 3. 运行服务

```bash
# 示例：Linux/macOS
export GERRIT_BASE_URL=""
export GERRIT_USER=""
export OPENAI_BASE_URL=""
export GERRIT_TOKEN=""
export OPENAI_API_KEY=""
export MODEL_NAME=""
export RULE_CONFIG_PATH="internal/config/examples/rules.json"
export CONTEXT_FILE_LIMIT=10
export CONTEXT_GRANULARITY="file"
./review-agent
```
```powershell
setx RULE_CONFIG_PATH "internal/config/examples/rules.json"
setx GERRIT_BASE_URL ""
setx GERRIT_USER ""
setx OPENAI_BASE_URL ""
setx GERRIT_TOKEN ""
setx OPENAI_API_KEY ""
setx MODEL_NAME ""
setx CONTEXT_FILE_LIMIT 10
setx CONTEXT_GRANULARITY "file"
```

## API 使用指南

## API 使用指南

### 1. 触发评审 (`POST /reviews/run`)

触发对指定 Change 和 Patchset 的评审。

**请求参数：**

```json
{
    "changeId": "I123456...",  // Gerrit Change ID
    "patchset": "1",           // Patchset Number
    "enableContext": true,     // 是否启用上下文获取
    "react": false             // 是否使用 ReAct 模式（高级编排）
}
```

**响应示例：**

```json
{
    "code": 0,
    "data": {
        "reviewId": "R173303...",
        "preview": { ... } // 评审建议预览
    }
}
```

### 2. 获取评审结果 (`GET /reviews/{id}`)

查看已生成的评审建议（不发布）。

**请求示例：**

```bash
curl "http://localhost:8000/reviews/R173303..."
```

### 3. 发布评审 (`POST /reviews/{id}/publish`)

将评审建议发布到 Gerrit。

**请求示例：**

```bash
curl -X POST "http://localhost:8000/reviews/R173303.../publish"
```

### 4. 重载规则 (`POST /config/rules/reload`)

热加载 `RULE_CONFIG_PATH` 指定的规则文件。

```bash
curl -X POST "http://localhost:8000/config/rules/reload"
```

## 静态规则配置示例 (`rules.json`)

```json
{
    "LinuxSpinSleep": true,
    "AndroidUiSleep": true,
    "FileTooLong": true,
    "FunctionLengthLimit": 50,
    "WhiteListFiles": ["generated.go"],
    "WhiteListFilesByLang": {
        "java": ["Test.java"]
    }
}
```
