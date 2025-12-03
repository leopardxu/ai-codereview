# 项目技术架构图（更新版）

## 组件概览

- Web 层：`internal/web`
  - `handler_reviews.go` 提供评审触发、结果查询、发布接口
  - `handler_config.go` 提供规则热重载接口（已安全加固）

- Eino 编排：`internal/app/eino`
  - `eino_flow.go` 定义两套编排：ReviewGraph（批处理）与 ReactGraph（ReAct 工具调用）

- 工具与分析：`internal/app/tools`
  - `GerritTool` 访问 Gerrit 变更与文件
  - `DiffTool` 解析补丁
  - `CodeContextTool` 拉取上下文（函数/类/依赖/文件）
  - `StaticRuleTool` 执行静态规则
  - `LLMTool` 生成建议（轻量模型）
  - `FormatForGerrit` 合并并格式化输出

- 配置与规则：`internal/config`
  - `config.go` 读取运行时配置
  - `rule_manager.go` 规则模型与热重载（字段白名单 + 目录限制）

- 调度与策略：`internal/app/scheduler`, `internal/app/policies`
  - `worker_pool.go` 异步任务调度
  - `rate_limit.go` 速率限制

## 数据流（ReviewGraph）

1. Web 接口接收 `changeId/patchset`
2. Graph 节点顺序：Diff → Context → Analyze(Static + LLM) → Merge → Format
3. 产出 `preview`（结构化评审建议），可查询或发布到 Gerrit

![ReviewGraph 架构图](./images/review_graph.svg)

```mermaid
flowchart LR
    subgraph Client
        A[用户/调用方]
    end
    subgraph Web[Web 层]
        H1[RunReview]
        H2[GetReview]
        H3[PublishReview]
        H4[ReloadRules]
    end

    subgraph EinoGraph[ReviewGraph]
        D1[Diff]
        C1[Context]
        A1[Analyze: StaticRuleTool]
        A2[Analyze: LLMTool]
        M1[Merge]
        F1[FormatForGerrit]
    end

    subgraph Tools[工具层]
        T1[GerritTool]
        T2[DiffTool]
        T3[CodeContextTool]
        T4[StaticRuleTool]
        T5[LLMTool]
    end

    subgraph Config[配置/规则]
        R1[rule_manager]
        R2[config]
    end

    subgraph Policies[策略]
        P1[RateLimiter]
    end

    A --> H1
    H1 --> D1
    D1 -->|GetDiffs| T1
    D1 --> T2
    D1 --> C1
    C1 -->|Fetch| T3
    C1 --> A1
    C1 --> A2
    A1 --> M1
    A2 --> M1
    M1 --> F1
    F1 --> O[Preview]
    H2 --> O
    H3 -->|PostReview| T1
    H4 -->|LoadRuleConfig| R1
    R1 -.-> T4
    R1 -.-> T3
    R2 --> H1
    P1 -.-> T1
    P1 -.-> T3
```

## 安全加固要点

- 规则热重载：`RULE_CONFIG_PATH` 限制在 `internal/config` 目录，必须 `.json`
- 字段白名单：仅接受预定义字段，出现未知字段则拒绝重载
- 外部交互：Gerrit 使用速率限制与基础错误处理

## ReAct 编排（ReactGraph）

![ReactGraph 编排图](./images/react_graph.svg)

```mermaid
flowchart LR
    subgraph ReactGraph
        RT[ChatTemplate]
        RM[ChatModel]
        TN[ToolsNode]
        CV[Convert]
    end
    RT --> RM
    RM --> TN
    TN --> CV
    CV --> PRV[Preview]
```

## 时序图（ReviewGraph 典型一次调用）

![ReviewGraph 时序图](./images/sequence_review.svg)

```mermaid
sequenceDiagram
    participant U as User
    participant W as Web
    participant G as EinoGraph
    participant GT as GerritTool
    participant CT as CodeContextTool
    participant ST as StaticRuleTool
    participant LT as LLMTool

    U->>W: RunReview(changeId, patchset)
    W->>G: Start Graph
    G->>GT: GetDiffs
    GT-->>G: diffs
    G->>CT: Fetch Context
    CT-->>G: ctxs
    G->>ST: Analyze Static
    ST-->>G: static advices
    G->>LT: Generate LLM advices
    LT-->>G: llm advices
    G->>W: preview
    U->>W: PublishReview(id)
    W->>GT: PostReview(payload)
    GT-->>W: status
```
