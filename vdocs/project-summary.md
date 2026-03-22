# MiroFish 项目总结

> **MiroFish** — 简洁通用的群体智能引擎，预测万物
>
> A Simple and Universal Swarm Intelligence Engine, Predicting Anything

---

## 一、项目概述

MiroFish 是一款基于**多智能体（Multi-Agent）技术**的新一代 AI 预测引擎。它通过提取现实世界的种子信息（突发新闻、政策草案、金融信号、小说故事等），自动构建出高保真的**平行数字世界**。在该世界中，成千上万个具备独立人格、长期记忆与行为逻辑的智能体自由交互与社会演化，从而推演未来走向并生成预测报告。

**核心定位：**

| 维度 | 描述 |
|------|------|
| **宏观** | 决策者的预演实验室，让政策与公关在零风险中试错 |
| **微观** | 个人用户的创意沙盘，推演小说结局或探索脑洞 |

---

## 二、技术栈

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'fontSize': '16px', 'fontFamily': 'Arial Black', 'primaryColor': '#E8F4FD', 'primaryTextColor': '#1a1a1a', 'primaryBorderColor': '#4A90D9', 'lineColor': '#4A90D9', 'secondaryColor': '#FFF3E0', 'tertiaryColor': '#E8F5E9'}}}%%
block-beta
    columns 3
    block:frontend["🖥️ 前端"]:3
        columns 3
        Vue3["Vue 3"] ViteB["Vite 7"] D3["D3.js"]
        VueRouter["Vue Router 4"] Axios["Axios"] space
    end
    block:backend["⚙️ 后端"]:3
        columns 3
        Flask["Flask 3.x"] Python["Python 3.11-3.12"] OpenAI["OpenAI SDK"]
        ZepSDK["Zep Cloud SDK"] OASIS["OASIS (camel-ai)"] PyMuPDF["PyMuPDF"]
    end
    block:infra["🏗️ 基础设施"]:3
        columns 3
        Docker["Docker Compose"] uv["uv 包管理"] npm["npm"]
    end

    style frontend fill:#E8F4FD,stroke:#4A90D9,color:#1a1a1a
    style backend fill:#FFF3E0,stroke:#E67E22,color:#1a1a1a
    style infra fill:#E8F5E9,stroke:#27AE60,color:#1a1a1a
```

---

## 三、系统架构

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black', 'primaryColor': '#E8F4FD', 'primaryTextColor': '#1a1a1a', 'primaryBorderColor': '#4A90D9', 'lineColor': '#555'}}}%%
graph TB
    subgraph Client["🌐 前端 (Vue 3 + Vite)"]
        direction LR
        S1["Step1<br/>图谱构建"]
        S2["Step2<br/>环境搭建"]
        S3["Step3<br/>开始模拟"]
        S4["Step4<br/>报告生成"]
        S5["Step5<br/>深度互动"]
        S1 --> S2 --> S3 --> S4 --> S5
    end

    subgraph API["🔌 REST API 层 (Flask)"]
        direction LR
        GA["/api/graph"]
        SA["/api/simulation"]
        RA["/api/report"]
    end

    subgraph Services["⚙️ 核心服务层"]
        direction TB
        OG["OntologyGenerator<br/>本体生成"]
        GB["GraphBuilderService<br/>图谱构建"]
        ER["ZepEntityReader<br/>实体读取"]
        PG["OasisProfileGenerator<br/>人设生成"]
        SCG["SimulationConfigGenerator<br/>仿真配置"]
        SM["SimulationManager<br/>模拟管理"]
        SR["SimulationRunner<br/>模拟运行"]
        RAG["ReportAgent<br/>报告智能体"]
        ZT["ZepToolsService<br/>工具集"]
    end

    subgraph External["🔗 外部服务"]
        LLM["LLM API<br/>(Qwen / OpenAI兼容)"]
        ZEP["Zep Cloud<br/>(GraphRAG)"]
        OASIS_ENG["OASIS 引擎<br/>(camel-ai)"]
    end

    Client -->|HTTP| API
    GA --> OG & GB & ER
    SA --> PG & SCG & SM & SR
    RA --> RAG & ZT

    OG -->|调用| LLM
    GB -->|构建图谱| ZEP
    ER -->|读取实体| ZEP
    PG -->|生成人设| LLM
    SCG -->|生成配置| LLM
    SR -->|子进程+IPC| OASIS_ENG
    RAG -->|思考+工具调用| LLM
    ZT -->|搜索+洞察| ZEP

    style Client fill:#E8F4FD,stroke:#4A90D9,color:#1a1a1a
    style API fill:#F3E5F5,stroke:#8E24AA,color:#1a1a1a
    style Services fill:#FFF3E0,stroke:#E67E22,color:#1a1a1a
    style External fill:#E8F5E9,stroke:#27AE60,color:#1a1a1a
```

---

## 四、五步工作流程

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black', 'primaryColor': '#E8F4FD', 'primaryTextColor': '#1a1a1a', 'lineColor': '#4A90D9'}}}%%
flowchart LR
    A["📄 Step 1<br/><b>图谱构建</b>"]
    B["🏗️ Step 2<br/><b>环境搭建</b>"]
    C["🚀 Step 3<br/><b>开始模拟</b>"]
    D["📊 Step 4<br/><b>报告生成</b>"]
    E["💬 Step 5<br/><b>深度互动</b>"]

    A -->|本体+图谱| B -->|人设+配置| C -->|Agent交互| D -->|ReACT生成| E

    style A fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style B fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
    style C fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style D fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style E fill:#FCE4EC,stroke:#C62828,color:#1a1a1a
```

### 各步骤详解

| 步骤 | 名称 | 功能描述 | 核心服务 |
|:----:|:----:|---------|---------|
| **1** | **图谱构建** | 上传种子文档 + 描述预测需求 → LLM 分析生成本体（实体/关系类型） → Zep 构建知识图谱 | `OntologyGenerator` `GraphBuilderService` |
| **2** | **环境搭建** | 从图谱提取实体 → 生成 Agent 人设档案 → LLM 配置仿真参数（时间、事件、轮次） | `ZepEntityReader` `OasisProfileGenerator` `SimulationConfigGenerator` |
| **3** | **开始模拟** | 启动 OASIS 引擎 → Twitter/Reddit 双平台并行仿真 → Agent 自由交互 → 动态记录行为 | `SimulationRunner` `OASIS` |
| **4** | **报告生成** | ReportAgent 使用 ReACT 模式 → 调用 Zep 工具集检索 → 多轮思考 → 分段生成预测报告 | `ReportAgent` `ZepToolsService` |
| **5** | **深度互动** | 与 ReportAgent 对话追问 → 与模拟世界中的任意 Agent 进行对话采访 | `ReportAgent.chat()` `SimulationRunner.interview()` |

---

## 五、核心功能模块详解

### 5.1 图谱构建模块

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
sequenceDiagram
    participant U as 👤 用户
    participant FE as 🖥️ 前端
    participant API as 🔌 API
    participant OG as 🧠 OntologyGenerator
    participant LLM as 🤖 LLM
    participant GB as 📦 GraphBuilder
    participant ZEP as ☁️ Zep Cloud

    U->>FE: 上传文档 + 输入模拟需求
    FE->>API: POST /api/graph/ontology/generate
    API->>OG: 分析文档
    OG->>LLM: 提取实体类型 & 关系类型
    LLM-->>OG: 返回本体定义 (JSON)
    OG-->>FE: 展示本体结构

    U->>FE: 确认并启动图谱构建
    FE->>API: POST /api/graph/build
    API->>GB: 异步构建 (后台线程)
    GB->>GB: 文本分块 (TextProcessor)
    GB->>ZEP: 创建 Standalone Graph
    GB->>ZEP: 批量添加 Episodes
    ZEP-->>GB: 图谱构建完成
    GB-->>FE: 返回 graph_id
```

**实现要点：**
- `OntologyGenerator` 使用精心设计的系统提示词引导 LLM 输出实体类型（如个人、机构、媒体）和关系类型（如批评、合作、影响）
- `GraphBuilderService` 使用 `TextProcessor` 对长文本进行智能分块（默认 500 字符/块，50 字符重叠）
- 通过 Zep Cloud 的 Standalone Graph API 构建知识图谱，支持 GraphRAG 语义检索
- 图谱构建是异步任务，前端通过轮询 `GET /api/graph/task/<task_id>` 获取进度

### 5.2 环境搭建模块

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black', 'primaryColor': '#E8F4FD', 'primaryTextColor': '#1a1a1a', 'lineColor': '#555'}}}%%
flowchart TD
    A["从 Zep 图谱提取实体"]
    B["按本体类型过滤"]
    C["为每个实体生成<br/>Twitter/Reddit 人设"]
    D["LLM 生成仿真参数"]
    E["输出配置文件"]

    A --> B --> C --> D --> E

    C --> C1["twitter_profiles.csv"]
    C --> C2["reddit_profiles.json"]
    D --> D1["simulation_config.json"]

    style A fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style B fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style C fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style D fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style E fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
    style C1 fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
    style C2 fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
    style D1 fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
```

**实现要点：**
- `ZepEntityReader` 从 Zep 图谱读取实体节点，并按照本体定义的类型进行过滤
- `OasisProfileGenerator` 为每个实体生成两套人设（Twitter 简介 + Reddit 简介），包含性格、身份、行为倾向
- `SimulationConfigGenerator` 根据预测需求，由 LLM 自动配置仿真时间跨度、轮次数、事件注入等参数
- 所有生成过程均为异步任务，前端通过 `POST /api/simulation/prepare/status` 轮询进度

### 5.3 OASIS 仿真引擎

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black', 'primaryColor': '#E8F4FD', 'primaryTextColor': '#1a1a1a', 'lineColor': '#555'}}}%%
graph TB
    subgraph Runner["SimulationRunner (主进程)"]
        direction TB
        Start["启动模拟"]
        Monitor["状态监控"]
        Log["行为记录<br/>actions.jsonl"]
        IPC["IPC 客户端"]
    end

    subgraph OASIS["OASIS 子进程"]
        direction TB
        TW["Twitter 平台仿真"]
        RD["Reddit 平台仿真"]

        subgraph Agents["智能体群"]
            A1["Agent 1<br/>独立人格+记忆"]
            A2["Agent 2<br/>独立人格+记忆"]
            A3["Agent N<br/>独立人格+记忆"]
        end

        TW --> Agents
        RD --> Agents
    end

    subgraph Actions["Agent 行为类型"]
        direction LR
        P["发帖 CREATE_POST"]
        L["点赞 LIKE_POST"]
        R["转发 REPOST"]
        C["评论 REPLY"]
        F["关注 FOLLOW"]
    end

    Start -->|subprocess| OASIS
    Monitor -->|轮询 run_state.json| OASIS
    IPC -->|命令通信| OASIS
    Agents --> Actions
    Actions -->|记录| Log

    style Runner fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style OASIS fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style Agents fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
    style Actions fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
```

**实现要点：**
- `SimulationRunner` 通过 `subprocess` 启动 OASIS 引擎（`run_parallel_simulation.py`），以独立进程运行
- 支持 **Twitter + Reddit 双平台并行仿真**，Agent 在两个社交平台上同时活动
- 每个 Agent 拥有独立的人格、记忆和行为逻辑，由 LLM 驱动决策
- 通过 `SimulationIPCClient` 实现主进程与 OASIS 子进程的进程间通信（IPC），支持 Agent 采访等交互命令
- `ZepGraphMemoryManager` 可选地将仿真中产生的 Agent 行为回写到 Zep 图谱，实现**动态时序记忆更新**
- 所有 Agent 行为记录在 `actions.jsonl` 中，运行状态通过 `run_state.json` 轮询

### 5.4 ReportAgent (ReACT 模式)

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black', 'primaryColor': '#E8F4FD', 'primaryTextColor': '#1a1a1a', 'lineColor': '#555'}}}%%
flowchart TD
    Start["开始生成报告"]
    Plan["📋 规划阶段<br/>LLM 生成报告目录"]
    Loop["🔄 分段生成循环"]

    subgraph ReACT["ReACT 生成每一节"]
        direction TB
        Think["🤔 思考 (Thought)<br/>分析需要什么信息"]
        Act["🔧 工具调用 (Action)<br/>search / insight_forge / panorama"]
        Observe["👁️ 观察 (Observation)<br/>获取工具返回结果"]
        Reflect["🪞 反思 (Reflection)<br/>信息是否充分?"]
        Write["✍️ 生成 (Write)<br/>撰写该节内容"]
        Think --> Act --> Observe --> Reflect
        Reflect -->|需要更多信息| Think
        Reflect -->|信息充分| Write
    end

    Tools["Zep 工具集"]
    T1["search<br/>语义搜索"]
    T2["insight_forge<br/>深度洞察"]
    T3["panorama<br/>全景概览"]
    T4["interview<br/>Agent 采访"]

    Start --> Plan --> Loop --> ReACT
    Act --> Tools
    Tools --> T1 & T2 & T3 & T4

    style Start fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style Plan fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style Loop fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style ReACT fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
    style Tools fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
```

**实现要点：**
- `ReportAgent` 采用 **ReACT（Reasoning + Acting）** 模式，先思考再行动
- 首先由 LLM 规划报告目录结构，然后逐节生成
- 每一节的生成经历多轮 Thought → Action → Observation → Reflection 循环
- **工具集（ZepToolsService）** 提供四大能力：
  - `search`：基于 Zep GraphRAG 的语义搜索
  - `insight_forge`：深度洞察分析，发现隐藏模式
  - `panorama`：全景概览，获取图谱统计信息
  - `interview`：对仿真中的 Agent 进行采访
- 详细日志通过 `ReportLogger` 记录在 `agent_log.jsonl`，前端可实时展示 Agent 思考过程

### 5.5 深度互动模块

用户可与两类对象进行对话：

| 对话对象 | 实现方式 | 用途 |
|:--------:|---------|------|
| **ReportAgent** | `ReportAgent.chat()` — 带工具调用的对话 | 追问报告细节、请求补充分析 |
| **仿真 Agent** | `SimulationRunner.interview_agent()` — IPC 通信 | 采访仿真世界中的任意角色，了解其想法 |

---

## 六、数据流总览

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'fontSize': '13px', 'fontFamily': 'Arial Black', 'primaryColor': '#E8F4FD', 'primaryTextColor': '#1a1a1a', 'lineColor': '#555'}}}%%
flowchart TB
    subgraph Input["📥 输入"]
        Doc["种子文档<br/>(PDF/TXT/MD)"]
        Req["预测需求<br/>(自然语言)"]
    end

    subgraph Phase1["Step 1: 图谱构建"]
        OG["OntologyGenerator<br/>→ 实体/关系本体"]
        GB["GraphBuilder<br/>→ 知识图谱"]
    end

    subgraph Phase2["Step 2: 环境搭建"]
        ER["EntityReader<br/>→ 实体列表"]
        PG["ProfileGenerator<br/>→ Agent 人设"]
        SCG["ConfigGenerator<br/>→ 仿真参数"]
    end

    subgraph Phase3["Step 3: 仿真运行"]
        SIM["OASIS 引擎<br/>Twitter + Reddit"]
        ACT["行为日志<br/>actions.jsonl"]
        MEM["图谱记忆更新"]
    end

    subgraph Phase4["Step 4: 报告生成"]
        RA["ReportAgent<br/>ReACT 多轮推理"]
        RPT["预测报告<br/>(Markdown)"]
    end

    subgraph Phase5["Step 5: 互动"]
        CHAT["对话互动"]
        ITV["Agent 采访"]
    end

    Doc & Req --> Phase1
    OG --> GB
    GB -->|graph_id| Phase2
    ER --> PG --> SCG
    Phase2 -->|profiles + config| Phase3
    SIM --> ACT
    SIM --> MEM
    Phase3 -->|仿真数据| Phase4
    RA --> RPT
    Phase4 --> Phase5

    style Input fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style Phase1 fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
    style Phase2 fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style Phase3 fill:#FCE4EC,stroke:#C62828,color:#1a1a1a
    style Phase4 fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style Phase5 fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
```

---

## 七、API 接口总览

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'fontSize': '13px', 'fontFamily': 'Arial Black', 'primaryColor': '#E8F4FD', 'primaryTextColor': '#1a1a1a', 'lineColor': '#555'}}}%%
graph LR
    subgraph GraphAPI["/api/graph"]
        G1["POST /ontology/generate"]
        G2["POST /build"]
        G3["GET /task/:id"]
        G4["GET /data/:graph_id"]
        G5["GET /project/:id"]
    end

    subgraph SimAPI["/api/simulation"]
        S1["POST /create"]
        S2["POST /prepare"]
        S3["POST /start"]
        S4["POST /stop"]
        S5["GET /:id/run-status"]
        S6["GET /:id/posts"]
        S7["POST /interview"]
    end

    subgraph ReportAPI["/api/report"]
        R1["POST /generate"]
        R2["GET /generate/status"]
        R3["GET /:id"]
        R4["POST /chat"]
        R5["GET /:id/sections"]
    end

    style GraphAPI fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style SimAPI fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style ReportAPI fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
```

---

## 八、项目目录结构

```
MiroFish/
├── backend/                          # Python Flask 后端
│   ├── app/
│   │   ├── api/                      # REST API 路由
│   │   │   ├── graph.py              # 图谱相关 API
│   │   │   ├── simulation.py         # 仿真相关 API
│   │   │   └── report.py             # 报告相关 API
│   │   ├── config.py                 # 配置管理
│   │   ├── models/                   # 数据模型
│   │   │   ├── project.py            # 项目模型 + ProjectManager
│   │   │   └── task.py               # 异步任务模型 + TaskManager
│   │   ├── services/                 # 核心业务逻辑
│   │   │   ├── ontology_generator.py # 本体生成 (LLM)
│   │   │   ├── graph_builder.py      # 图谱构建 (Zep)
│   │   │   ├── zep_entity_reader.py  # 实体读取
│   │   │   ├── oasis_profile_generator.py  # Agent 人设生成
│   │   │   ├── simulation_config_generator.py  # 仿真配置生成
│   │   │   ├── simulation_manager.py # 仿真管理
│   │   │   ├── simulation_runner.py  # 仿真运行器 (OASIS)
│   │   │   ├── simulation_ipc.py     # IPC 通信客户端
│   │   │   ├── report_agent.py       # ReACT 报告智能体
│   │   │   ├── zep_tools.py          # Zep 工具集
│   │   │   └── zep_graph_memory_updater.py  # 图谱记忆更新
│   │   └── utils/                    # 工具类
│   │       ├── llm_client.py         # LLM 统一客户端
│   │       ├── file_parser.py        # 文件解析器
│   │       └── logger.py             # 日志模块
│   ├── scripts/                      # OASIS 仿真脚本
│   └── run.py                        # 后端入口
├── frontend/                         # Vue 3 前端 SPA
│   ├── src/
│   │   ├── api/                      # API 客户端封装
│   │   │   ├── index.js              # Axios 实例
│   │   │   ├── graph.js              # 图谱 API
│   │   │   ├── simulation.js         # 仿真 API
│   │   │   └── report.js             # 报告 API
│   │   ├── components/               # 步骤组件
│   │   │   ├── Step1GraphBuild.vue   # 图谱构建
│   │   │   ├── Step2EnvSetup.vue     # 环境搭建
│   │   │   ├── Step3Simulation.vue   # 仿真运行
│   │   │   ├── Step4Report.vue       # 报告生成
│   │   │   └── Step5Interaction.vue  # 深度互动
│   │   ├── views/                    # 页面视图
│   │   ├── App.vue                   # 根组件
│   │   └── main.js                   # 前端入口
│   └── vite.config.js                # Vite 配置
├── .env.example                      # 环境变量模板
├── docker-compose.yml                # Docker 部署
├── package.json                      # 根包 (concurrently)
└── README.md                         # 项目文档
```

---

## 九、关键设计模式

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black', 'primaryColor': '#E8F4FD', 'primaryTextColor': '#1a1a1a', 'lineColor': '#555'}}}%%
mindmap
    root(("**MiroFish<br/>设计模式**"))
        **异步任务**
            后台线程执行耗时操作
            前端轮询获取进度
            TaskManager 统一管理
        **ReACT Agent**
            Thought 思考
            Action 工具调用
            Observation 观察
            Reflection 反思
        **IPC 通信**
            主进程与 OASIS 子进程通信
            支持 Agent 采访
            命令/响应模式
        **GraphRAG**
            Zep Cloud 知识图谱
            语义搜索 + 结构化查询
            动态记忆更新
        **多平台仿真**
            Twitter 社交平台
            Reddit 论坛平台
            并行模拟
```

| 模式 | 说明 |
|------|------|
| **异步任务模式** | 图谱构建、环境搭建、报告生成等耗时操作均在后台线程执行，前端通过轮询获取任务状态 |
| **ReACT Agent** | 报告生成采用 Reasoning + Acting 循环模式，Agent 自主决定何时调用工具、何时生成内容 |
| **IPC 进程间通信** | `SimulationIPCClient` 实现主进程与 OASIS 子进程的命令通信，支持 Agent 采访等交互场景 |
| **GraphRAG** | 基于 Zep Cloud 的知识图谱，结合语义搜索与图结构查询，为 ReportAgent 提供丰富的上下文 |
| **文件存储** | 项目、仿真、报告数据以文件形式存储在 `backend/uploads/` 下，通过 ProjectManager/TaskManager 管理 |

---

## 十、部署方式

| 方式 | 命令 | 端口 |
|:----:|------|:----:|
| **源码部署** | `npm run setup:all && npm run dev` | 前端 3000 / 后端 5001 |
| **Docker** | `docker compose up -d` | 前端 3000 / 后端 5001 |

**环境变量要求：**

| 变量 | 必需 | 用途 |
|------|:----:|------|
| `LLM_API_KEY` | ✅ | LLM API 密钥 (推荐阿里百炼 Qwen) |
| `LLM_BASE_URL` | ✅ | LLM API 地址 |
| `LLM_MODEL_NAME` | ✅ | LLM 模型名称 |
| `ZEP_API_KEY` | ✅ | Zep Cloud API 密钥 |
| `LLM_BOOST_*` | ❌ | 可选的加速 LLM 配置 |
