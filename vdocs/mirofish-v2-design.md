# MiroFish V2 设计方案

> **核心升级**：从"简单仿真 + 出报告"进化为"深度认知仿真 + 群体预测精炼"  
> **设计原则**：借鉴 Smallville（认知深度）、Project Sid（社会感知）、Concordia（GM 仲裁）、AgentSociety（模块化）、DeLLMphi（预测精炼）  
> **技术栈**：Go + OpenClaw Gateway + 百炼 Coding Plan（约束不变）

---

## 一、V1 vs V2 核心差异

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
graph LR
    subgraph V1["🐟 MiroFish V1"]
        direction TB
        V1A["Prompt 驱动 Agent<br/>(无记忆/无反思)"]
        V1B["OASIS 仿真"]
        V1C["直接生成报告"]
        V1A --> V1B --> V1C
    end

    subgraph V2["🐟 MiroFish V2"]
        direction TB
        V2A["三层认知 Agent<br/>(记忆流+反思+计划)"]
        V2B["GM 仲裁仿真<br/>(社会感知+涌现)"]
        V2C["德尔菲预测精炼<br/>(多轮辩论+校准)"]
        V2D["分析报告"]
        V2A --> V2B --> V2C --> V2D
    end

    V1 -->|"升级"| V2

    style V1 fill:#FFEBEE,stroke:#C62828,color:#1a1a1a
    style V2 fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
```

| 维度 | V1 | V2 | 借鉴来源 |
|:----:|:---:|:---:|:--------:|
| Agent 认知 | 单次 Prompt 决策 | 记忆流 + 反思 + 层级计划 | Smallville |
| 社会动力 | Agent 独立行动 | 声誉/影响力/好感度追踪 | Project Sid |
| 环境管理 | Agent 直接操控平台 | Game Master 仲裁 + 合理性检查 | Concordia |
| Agent 架构 | 硬编码 | Agent-Block-Action 可插拔 | AgentSociety |
| 预测输出 | 仿真 → 报告 | 仿真 → 德尔菲精炼 → 校准 → 报告 | DeLLMphi |
| 报告质量 | 单次 ReACT 生成 | 多视角辩论 + 统计校准 | Wisdom of Crowds |

---

## 二、总体架构

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '13px', 'fontFamily': 'Arial Black'}}}%%
graph TB
    subgraph Client["🌐 前端 (Vue 3)"]
        UI["六步流程 UI"]
        Viz["D3 仿真可视化"]
        Monitor["Agent 状态监控"]
    end

    subgraph GoBackend["⚙️ Go 后端"]
        Router["REST API + WebSocket"]
        Orch["编排器 (Orchestrator)"]

        subgraph Cognitive["🧠 认知引擎"]
            MemEngine["记忆流引擎"]
            ReflectEngine["反思引擎"]
            PlanEngine["计划引擎"]
        end

        subgraph SimEngine["🎮 仿真引擎"]
            GM["Game Master<br/>(环境仲裁)"]
            SocialGraph["社会关系图<br/>(声誉/影响力/好感度)"]
            Platform["虚拟社交平台<br/>(信息流+推荐)"]
            EventInjector["事件注入器"]
        end

        subgraph PredEngine["🔮 预测精炼引擎"]
            DelphiOrch["德尔菲编排器"]
            ExpertAgents["专家 Agent 团<br/>(多视角)"]
            Mediator["调解 Agent"]
            Calibrator["统计校准器<br/>(Platt+极端化)"]
        end

        subgraph ReactEngine["📊 报告引擎"]
            ReACT["ReACT 引擎"]
            Tools["工具集<br/>(图谱/仿真数据/预测结果)"]
        end
    end

    subgraph Gateway["🧠 OpenClaw Gateway"]
        MainL["Main Lane (max=4)"]
        SubL["Subagent Lane (max=8)"]
    end

    subgraph LLM["☁️ 百炼"]
        QP["qwen-plus"]
        QF["qwen-flash"]
    end

    subgraph Data["💾 数据层"]
        Neo["Neo4j<br/>(知识图谱+社会图)"]
        PG["PostgreSQL<br/>(项目/任务/仿真)"]
        Redis["Redis<br/>(记忆流/状态缓存)"]
    end

    Client -->|HTTP/WS| Router
    Router --> Orch
    Orch --> Cognitive & SimEngine & PredEngine & ReactEngine
    Cognitive --> MemEngine & ReflectEngine & PlanEngine
    SimEngine --> GM & SocialGraph & Platform

    ReactEngine --> MainL
    Orch -->|"单次推理"| MainL
    SimEngine -->|"sessions_spawn"| SubL
    PredEngine -->|"sessions_spawn"| SubL
    MainL & SubL --> LLM
    GoBackend --> Data

    style Client fill:#E8F4FD,stroke:#4A90D9,color:#1a1a1a
    style GoBackend fill:#FFF3E0,stroke:#E67E22,color:#1a1a1a
    style Cognitive fill:#FCE4EC,stroke:#C62828,color:#1a1a1a
    style SimEngine fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style PredEngine fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style ReactEngine fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
    style Gateway fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style LLM fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
    style Data fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
```

---

## 三、六步工作流（V1 五步 → V2 六步）

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
flowchart LR
    A["📄 Step 1<br/><b>图谱构建</b>"]
    B["🏗️ Step 2<br/><b>环境搭建</b>"]
    C["🚀 Step 3<br/><b>社会仿真</b>"]
    D["🔮 Step 4<br/><b>预测精炼</b><br/>(新增)"]
    E["📊 Step 5<br/><b>报告生成</b>"]
    F["💬 Step 6<br/><b>深度互动</b>"]

    A -->|"本体+图谱"| B
    B -->|"认知Agent+配置"| C
    C -->|"仿真数据+涌现信号"| D
    D -->|"精炼预测+置信度"| E
    E -->|"ReACT报告"| F

    style A fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style B fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
    style C fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style D fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style E fill:#FCE4EC,stroke:#C62828,color:#1a1a1a
    style F fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
```

**新增 Step 4（预测精炼）**：这是 V2 最重要的升级。仿真输出的原始行为数据需经过"德尔菲式"多 Agent 辩论和统计校准，才能生成高质量预测。

---

## 四、Agent 认知架构（核心升级 1）

### 4.1 三层认知模型

借鉴 Smallville 的记忆流 + 反思 + 计划，并加入 Project Sid 的社会感知，构建 V2 的 Agent 认知模型：

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
graph TB
    subgraph CogAgent["🧠 V2 认知 Agent"]
        direction TB

        subgraph L1["Layer 1: 记忆流 (Memory Stream)"]
            MS["自然语言记忆存储"]
            Score["三维评分<br/>recency × importance × relevance"]
            Retrieve["加权检索"]
        end

        subgraph L2["Layer 2: 反思 (Reflection)"]
            Trigger["触发条件<br/>(累积重要性 > 阈值)"]
            Questions["生成抽象问题"]
            Insights["产出高层洞察<br/>写回记忆流"]
        end

        subgraph L3["Layer 3: 社会感知 + 计划"]
            SocPerception["社会感知<br/>声誉/影响力/好感度"]
            DayPlan["日计划 → 小时计划"]
            Reactive["突发事件动态调整"]
        end

        subgraph Action["⚡ 行动决策"]
            Context["上下文组装<br/>= 当前观察<br/>+ Top-K 记忆<br/>+ 最新反思<br/>+ 社会关系<br/>+ 当前计划"]
            Decide["LLM 决策"]
            Execute["执行行为"]
        end
    end

    L1 --> L2
    L2 -->|"洞察"| L1
    L1 & L2 & L3 --> Action
    Execute -->|"新经历"| L1

    style CogAgent fill:#E8F4FD,stroke:#4A90D9,color:#1a1a1a
    style L1 fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style L2 fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style L3 fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style Action fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
```

### 4.2 记忆流实现

```go
// internal/cognitive/memory.go

type MemoryEntry struct {
    ID         string    `json:"id"`
    AgentID    string    `json:"agent_id"`
    Content    string    `json:"content"`
    Timestamp  time.Time `json:"timestamp"`
    Importance float64   `json:"importance"`
    Kind       string    `json:"kind"` // "observation" | "action" | "reflection" | "plan"
    Embedding  []float32 `json:"embedding,omitempty"`
}

type MemoryStream struct {
    store   MemoryStore     // Redis sorted set (by timestamp)
    embedder EmbeddingClient // 向量化
}

func (ms *MemoryStream) Retrieve(ctx context.Context, query string, k int) ([]MemoryEntry, error) {
    now := time.Now()
    candidates := ms.store.GetRecent(ctx, ms.agentID, 100)
    queryEmb := ms.embedder.Embed(ctx, query)

    scored := make([]scoredEntry, len(candidates))
    for i, entry := range candidates {
        recency := exponentialDecay(now.Sub(entry.Timestamp), decayRate)
        relevance := cosineSimilarity(queryEmb, entry.Embedding)
        importance := entry.Importance
        scored[i] = scoredEntry{
            entry: entry,
            score: recency*wRecency + relevance*wRelevance + importance*wImportance,
        }
    }

    sort.Slice(scored, func(i, j int) bool { return scored[i].score > scored[j].score })
    return top(scored, k), nil
}
```

### 4.3 反思机制

```go
// internal/cognitive/reflection.go

type ReflectionEngine struct {
    cc       *chatcompletions.Client
    memStore MemoryStore
}

func (r *ReflectionEngine) MaybeReflect(ctx context.Context, agent *Agent) error {
    recentMemories := r.memStore.GetSince(ctx, agent.ID, agent.LastReflectionTime)
    totalImportance := sumImportance(recentMemories)

    if totalImportance < reflectionThreshold {
        return nil // 不需要反思
    }

    // Step 1: 从近期记忆生成高层问题
    questions, _ := r.cc.Complete(ctx, chatcompletions.Request{
        Messages: []Message{
            {Role: "system", Content: reflectionQuestionPrompt},
            {Role: "user", Content: formatMemories(recentMemories)},
        },
    })

    // Step 2: 用记忆回答问题，产生洞察
    for _, q := range parseQuestions(questions) {
        relevantMemories, _ := agent.Memory.Retrieve(ctx, q, 10)
        insight, _ := r.cc.Complete(ctx, chatcompletions.Request{
            Messages: []Message{
                {Role: "system", Content: reflectionInsightPrompt},
                {Role: "user", Content: fmt.Sprintf("问题: %s\n相关记忆:\n%s", q, formatMemories(relevantMemories))},
            },
        })
        // Step 3: 洞察写回记忆流
        agent.Memory.Add(ctx, MemoryEntry{
            Content:    insight,
            Kind:       "reflection",
            Importance: highImportance,
        })
    }

    agent.LastReflectionTime = time.Now()
    return nil
}
```

### 4.4 社会感知图

借鉴 Project Sid，每个 Agent 维护一个社会关系图：

```go
// internal/cognitive/social.go

type SocialPerception struct {
    AgentID     string
    TargetID    string
    Reputation  float64   // [-1, 1] 声誉评价
    Influence   float64   // [0, 1] 影响力权重
    Likability  float64   // [-1, 1] 好感度
    LastUpdated time.Time
}

type SocialGraph struct {
    store *neo4j.Driver
}

// Agent 行为后更新社会感知
func (sg *SocialGraph) UpdateAfterInteraction(ctx context.Context, observer, target string, interaction InteractionRecord) error {
    delta := evaluateInteraction(interaction) // LLM 评估交互的正/负影响
    current, _ := sg.Get(ctx, observer, target)
    current.Likability = clamp(current.Likability+delta.Likability, -1, 1)
    current.Reputation = clamp(current.Reputation+delta.Reputation, -1, 1)
    current.Influence = updateInfluence(current.Influence, interaction)
    return sg.Set(ctx, current)
}
```

---

## 五、Game Master 仲裁引擎（核心升级 2）

借鉴 Concordia，引入 Game Master 作为环境仲裁者，而不是让 Agent 直接操控平台：

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
sequenceDiagram
    participant Agent as 🤖 Agent
    participant GM as 🎲 Game Master
    participant Platform as 📱 社交平台
    participant SocGraph as 👥 社会关系图
    participant MemStream as 📝 记忆流

    Note over Agent: 感知 → 检索记忆 → 反思(可选) → 计划

    Agent->>GM: "我想发表一条关于政策影响的帖子"<br/>(自然语言意图)

    GM->>GM: 合理性检查<br/>1. 该Agent的人设会这样做吗?<br/>2. 当前时间/情境合理吗?<br/>3. 内容与Agent记忆一致吗?

    alt 合理
        GM->>Platform: 执行: CreatePost(agent_id, content)
        Platform-->>GM: 成功, post_id=123
        GM->>SocGraph: 更新: 关注者看到此帖
        GM->>MemStream: 写入Agent记忆: "我发了一条帖子..."
        GM-->>Agent: "你成功发布了帖子，获得了3个点赞"
    else 不合理
        GM-->>Agent: "你作为一名保守派学者，<br/>不太可能发布这样的激进言论。<br/>你可以选择其他行动。"
        Agent->>GM: "那我转发并评论表达温和看法"
        GM->>Platform: 执行: Repost + Comment
    end
```

### GM 的关键职责

```go
// internal/simulation/game_master.go

type GameMaster struct {
    cc       *chatcompletions.Client
    platform *Platform
    social   *SocialGraph
    memory   *MemoryStreamManager
}

type ActionProposal struct {
    AgentID     string `json:"agent_id"`
    Intent      string `json:"intent"`      // 自然语言意图
    ActionType  string `json:"action_type"` // "post" | "reply" | "like" | "follow" | "repost"
    Content     string `json:"content,omitempty"`
    TargetID    string `json:"target_id,omitempty"`
}

type GMVerdict struct {
    Approved    bool   `json:"approved"`
    Reason      string `json:"reason"`
    ModifiedAction *ActionProposal `json:"modified_action,omitempty"`
}

func (gm *GameMaster) Arbitrate(ctx context.Context, proposal ActionProposal, agent *Agent) (*GMVerdict, error) {
    profile := agent.Profile
    recentMemory, _ := agent.Memory.Retrieve(ctx, proposal.Intent, 5)
    socialContext, _ := gm.social.GetRelationships(ctx, agent.ID)

    verdict, _ := gm.cc.Complete(ctx, chatcompletions.Request{
        Messages: []Message{
            {Role: "system", Content: gmArbitratePrompt},
            {Role: "user", Content: formatArbitrateInput(proposal, profile, recentMemory, socialContext)},
        },
    })

    return parseVerdict(verdict), nil
}
```

### 为什么需要 GM

| 没有 GM (V1) | 有 GM (V2) |
|:------------:|:----------:|
| Agent 直接执行行为，可能出现不符合人设的行为 | GM 检查一致性，确保行为合理 |
| 无法处理 Agent 间冲突 | GM 仲裁冲突（如两人同时争抢资源） |
| 环境反馈单一（成功/失败） | GM 用叙事描述环境变化，丰富 Agent 感知 |
| 行为质量依赖单次 Prompt | GM 额外检查，双重保障行为质量 |

**成本优化**：GM 不需要对每个行为都调用 LLM。可以用规则引擎处理简单行为（点赞/关注），仅对复杂行为（发帖/评论）调用 LLM 仲裁：

```go
func (gm *GameMaster) ShouldLLMArbitrate(proposal ActionProposal) bool {
    switch proposal.ActionType {
    case "like", "follow", "repost":
        return false // 简单行为用规则引擎
    case "post", "reply", "comment":
        return true  // 涉及内容生成的行为需 LLM 仲裁
    default:
        return true
    }
}
```

---

## 六、预测精炼引擎（核心升级 3 — V2 新增）

这是 V2 最关键的新增模块，借鉴 DeLLMphi 的德尔菲法和群体预测研究。

### 6.1 整体流程

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
graph TB
    subgraph Input["📥 仿真输出"]
        SimData["仿真行为日志"]
        Trends["涌现趋势信号"]
        SocState["社会关系状态"]
    end

    subgraph Delphi["🔮 德尔菲预测精炼"]
        direction TB
        subgraph Experts["👨‍💼 多视角专家 Agent"]
            E1["乐观派分析师<br/>(sessions_spawn)"]
            E2["悲观派分析师<br/>(sessions_spawn)"]
            E3["数据派分析师<br/>(sessions_spawn)"]
            E4["领域专家<br/>(sessions_spawn)"]
            E5["博弈论专家<br/>(sessions_spawn)"]
        end

        subgraph Round["🔄 多轮迭代 (3-5轮)"]
            R1["Round 1: 独立预测<br/>各专家基于仿真数据独立分析"]
            R2["调解 Agent 综合分歧<br/>识别关键争议点"]
            R3["Round 2-N: 修正预测<br/>专家看到分歧后更新判断"]
            R4["收敛判定<br/>预测变化 < 阈值"]
        end
    end

    subgraph Calibrate["📐 统计校准"]
        Aggregate["聚合预测<br/>Median / 加权平均"]
        Platt["Platt 缩放<br/>(矫正LLM保守倾向)"]
        Extremize["极端化修正"]
        Confidence["置信区间估算"]
    end

    subgraph Output["📊 输出"]
        Prediction["精炼预测结果<br/>+ 置信度<br/>+ 关键论据<br/>+ 分歧点"]
    end

    Input --> Delphi
    Experts --> Round
    Round --> Calibrate
    Calibrate --> Output

    style Input fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style Delphi fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style Experts fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style Round fill:#FCE4EC,stroke:#C62828,color:#1a1a1a
    style Calibrate fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
    style Output fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
```

### 6.2 详细时序

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '13px', 'fontFamily': 'Arial Black'}}}%%
sequenceDiagram
    participant Go as ⚙️ 预测精炼引擎
    participant TI as 🔧 toolsinvoke
    participant SL as 📋 Subagent Lane
    participant CC as 🔌 chatCompletions

    Note over Go: 从仿真引擎获取: 行为日志 + 趋势 + 社会状态

    loop 每轮 (Round 1 ~ N)
        par 并发 spawn 5 个专家 Agent
            Go->>TI: sessions_spawn(乐观派分析师, 仿真摘要+上轮反馈)
            Go->>TI: sessions_spawn(悲观派分析师, 仿真摘要+上轮反馈)
            Go->>TI: sessions_spawn(数据派分析师, 仿真摘要+上轮反馈)
            Go->>TI: sessions_spawn(领域专家, 仿真摘要+上轮反馈)
            Go->>TI: sessions_spawn(博弈论专家, 仿真摘要+上轮反馈)
        end

        TI->>SL: 5 tasks ≤ maxConcurrent=8
        Note over SL: 全部并行执行

        Go->>Go: 轮询收集 5 个专家预测结果

        Go->>CC: 调解 Agent 综合分歧<br/>(chatCompletions + function calling)
        CC-->>Go: 分歧分析 + 结构化反馈

        Go->>Go: 收敛判定 (标准差 < 阈值?)
        alt 已收敛
            Note over Go: 退出循环
        else 未收敛
            Note over Go: 将调解反馈作为下轮输入
        end
    end

    Go->>Go: 聚合预测 + Platt校准 + 极端化修正
    Go->>Go: 输出: 精炼预测 + 置信度 + 论据 + 分歧
```

### 6.3 核心代码

```go
// internal/prediction/delphi.go

type DelphiEngine struct {
    ti          *toolsinvoke.Client
    cc          *chatcompletions.Client
    calibrator  *Calibrator
}

type ExpertPerspective struct {
    Name       string `json:"name"`
    SystemPrompt string `json:"system_prompt"`
}

var defaultExperts = []ExpertPerspective{
    {Name: "optimist", SystemPrompt: "你是一位乐观派分析师，倾向于看到积极信号和增长机会..."},
    {Name: "pessimist", SystemPrompt: "你是一位风险分析师，专注于识别负面因素和潜在风险..."},
    {Name: "quant", SystemPrompt: "你是一位数据驱动的量化分析师，只关注可量化的证据..."},
    {Name: "domain", SystemPrompt: "你是该领域的资深专家，拥有丰富的历史案例知识..."},
    {Name: "strategist", SystemPrompt: "你是一位博弈论专家，分析各方利益博弈和策略互动..."},
}

type PredictionResult struct {
    Question      string            `json:"question"`
    Probability   float64           `json:"probability"`
    Confidence    float64           `json:"confidence"`
    KeyArguments  []string          `json:"key_arguments"`
    Disagreements []string          `json:"disagreements"`
    ExpertViews   map[string]float64 `json:"expert_views"`
    Rounds        int               `json:"rounds"`
}

func (d *DelphiEngine) Refine(ctx context.Context, simSummary string, questions []string) ([]PredictionResult, error) {
    results := make([]PredictionResult, len(questions))

    for qi, question := range questions {
        var prevFeedback string
        var expertPredictions map[string]float64

        for round := 1; round <= maxRounds; round++ {
            // 1. 并发 spawn 专家 Agent
            expertPredictions = make(map[string]float64)
            runIDs := make(map[string]string)

            for _, expert := range defaultExperts {
                resp, _ := d.ti.Invoke(ctx, toolsinvoke.Request{
                    Tool: "sessions_spawn",
                    Args: map[string]any{
                        "task": buildExpertPrompt(expert, question, simSummary, prevFeedback),
                        "model": "dashscope:qwen-plus",
                        "label": fmt.Sprintf("delphi-r%d-%s", round, expert.Name),
                    },
                })
                runIDs[expert.Name] = parseRunID(resp.Result)
            }

            // 2. 收集专家预测
            for name, runID := range runIDs {
                result := waitForResult(ctx, d.ti, runID)
                expertPredictions[name] = parseProbability(result)
            }

            // 3. 收敛检查
            if stdDev(values(expertPredictions)) < convergenceThreshold {
                break
            }

            // 4. 调解 Agent 综合分歧
            prevFeedback, _ = d.mediate(ctx, question, expertPredictions)
        }

        // 5. 聚合 + 校准
        raw := median(values(expertPredictions))
        calibrated := d.calibrator.Calibrate(raw)

        results[qi] = PredictionResult{
            Question:    question,
            Probability: calibrated,
            Confidence:  1.0 - stdDev(values(expertPredictions)),
            ExpertViews: expertPredictions,
        }
    }

    return results, nil
}
```

### 6.4 统计校准器

```go
// internal/prediction/calibrator.go

type Calibrator struct {
    // Platt 缩放参数 (可从历史数据训练)
    A float64 // 默认 -1.0
    B float64 // 默认 0.0
}

func (c *Calibrator) Calibrate(rawProb float64) float64 {
    // Step 1: Platt 缩放 — 矫正 LLM 的系统性偏差
    logit := math.Log(rawProb / (1.0 - rawProb))
    plattProb := 1.0 / (1.0 + math.Exp(c.A*logit+c.B))

    // Step 2: 极端化修正 — LLM 天然偏保守(趋向50%), 往外推
    extremized := extremize(plattProb, extremizationFactor)

    return clamp(extremized, 0.01, 0.99)
}

func extremize(p float64, factor float64) float64 {
    logOdds := math.Log(p / (1.0 - p))
    extremeLogOdds := logOdds * factor // factor > 1 往外推
    return 1.0 / (1.0 + math.Exp(-extremeLogOdds))
}
```

---

## 七、Agent-Block-Action 可插拔架构（核心升级 4）

借鉴 AgentSociety，将 Agent 能力模块化：

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
graph TB
    subgraph AgentV2["🤖 V2 Agent 结构"]
        direction TB
        Profile["📋 Agent Profile<br/>(人设/目标/偏好)"]

        subgraph Blocks["🧩 可插拔 Blocks"]
            MemBlock["MemoryBlock<br/>记忆流管理"]
            ReflBlock["ReflectionBlock<br/>反思引擎"]
            PlanBlock["PlanningBlock<br/>计划生成"]
            SocBlock["SocialBlock<br/>社会感知"]
            PostBlock["PostingBlock<br/>发帖策略"]
            ReactBlock["ReactionBlock<br/>互动策略"]
        end

        Dispatcher["🎯 Block 调度器<br/>根据当前情境选择激活哪些 Block"]
        Output["⚡ 行动输出"]
    end

    Profile --> Dispatcher
    Dispatcher --> Blocks
    Blocks --> Output

    style AgentV2 fill:#E8F4FD,stroke:#4A90D9,color:#1a1a1a
    style Blocks fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
```

```go
// internal/agent/block.go

type Block interface {
    Name() string
    ShouldActivate(ctx *StepContext) bool
    Execute(ctx *StepContext) (*BlockOutput, error)
}

type Agent struct {
    ID       string
    Profile  AgentProfile
    Blocks   []Block
    Memory   *MemoryStream
    Social   *SocialPerception
}

// 每步执行流: 遍历所有 Block, 激活符合条件的
func (a *Agent) Step(ctx context.Context, stepCtx *StepContext) (*ActionProposal, error) {
    var outputs []*BlockOutput

    for _, block := range a.Blocks {
        if block.ShouldActivate(stepCtx) {
            out, err := block.Execute(stepCtx)
            if err != nil {
                continue
            }
            outputs = append(outputs, out)
        }
    }

    return a.decideAction(ctx, stepCtx, outputs)
}
```

**好处**：可以根据场景替换 Block。比如：
- 金融预测场景 → 加 `MarketBlock`（关注涨跌信号）
- 舆情预测场景 → 加 `EmotionBlock`（情感传播）
- 选举预测场景 → 加 `VotingBlock`（投票意向）

---

## 八、仿真引擎 V2 改造

### 8.1 V2 仿真每轮流程

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '13px', 'fontFamily': 'Arial Black'}}}%%
sequenceDiagram
    participant Orch as ⚙️ 编排器
    participant Cog as 🧠 认知引擎
    participant GM as 🎲 Game Master
    participant Platform as 📱 平台
    participant Social as 👥 社会图
    participant TI as 🔧 toolsinvoke

    loop 每轮 (Round)
        Note over Orch: Phase 1: 认知处理 (并发 spawn)

        par 为每组 Agent 并发执行认知
            Orch->>TI: spawn(GroupA认知: 记忆检索→反思→计划→意图)
            Orch->>TI: spawn(GroupB认知: ...)
            Orch->>TI: spawn(GroupC认知: ...)
            Orch->>TI: spawn(GroupD认知: ...)
            Orch->>TI: spawn(GroupE认知: ...)
        end

        Orch->>Orch: 收集所有Agent意图(ActionProposals)

        Note over Orch: Phase 2: GM 仲裁 (批量)

        Orch->>GM: 批量仲裁 ActionProposals
        GM->>GM: 规则引擎过滤简单行为
        GM->>TI: spawn(LLM仲裁复杂行为)
        GM-->>Orch: 通过/拒绝/修改

        Note over Orch: Phase 3: 执行 + 更新

        Orch->>Platform: 执行通过的行为
        Platform-->>Orch: 行为结果(点赞数/评论回复...)
        Orch->>Social: 更新社会关系图
        Orch->>Cog: 将结果写入各Agent记忆流

        Note over Orch: Phase 4: 涌现信号检测
        Orch->>Orch: 分析本轮涌现模式<br/>(极化趋势/意见领袖变化/话题热度)
    end
```

### 8.2 V1 vs V2 每轮对比

| 阶段 | V1 | V2 |
|:----:|:---:|:---:|
| Agent 决策 | 单次 Prompt | 记忆检索 → 反思 → 计划 → 意图 |
| 行为验证 | 无 | GM 仲裁（规则 + LLM） |
| 社会关系 | 不追踪 | 每轮更新声誉/影响力/好感度 |
| 涌现检测 | 无 | 每轮分析涌现信号 |
| LLM 调用/轮 | 5 次 (5组×1) | ~12 次 (5组认知 + 2 GM + 5社会更新) |

---

## 九、调用量分析 & 成本控制

### 完整流程调用量

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '13px', 'fontFamily': 'Arial Black'}}}%%
pie title V2 LLM 调用分布 (30 Agent, 40 轮仿真)
    "本体分析 + 图谱" : 2
    "人设生成" : 5
    "仿真认知 (40轮×5组)" : 200
    "仿真GM仲裁 (~40轮×2)" : 80
    "社会感知更新 (~40轮×3)" : 120
    "反思触发 (~20次)" : 20
    "预测精炼 (5专家×4轮)" : 20
    "调解Agent (4轮)" : 4
    "ReACT报告 (~20)" : 20
    "互动/采访 (~10)" : 10
```

| 步骤 | 路径 | 调用次数 | 模型 |
|:----:|:----:|:-------:|:----:|
| 本体分析 | chatCompletions | 1 | qwen-plus |
| 图谱构建 | 纯代码 | 0 | — |
| 人设生成 ×5 组 | sessions_spawn | 5 | qwen-plus |
| 仿真认知 40轮×5组 | sessions_spawn | **200** | qwen-flash |
| GM 仲裁 40轮×~2 | sessions_spawn | **80** | qwen-flash |
| 社会感知 40轮×~3 | chatCompletions | **120** | qwen-flash |
| 反思 ~20 次 | chatCompletions | 20 | qwen-plus |
| **预测精炼** 5专家×4轮 | sessions_spawn | **20** | qwen-plus |
| 调解 Agent 4轮 | chatCompletions | 4 | qwen-plus |
| 校准 | 纯计算 | 0 | — |
| ReACT 报告 | chatCompletions | ~20 | qwen-plus |
| 互动/采访 | 混合 | ~10 | qwen-plus |
| **总计** | | **~481** | |

**Pro 套餐 5h 额度 6000 次 → 481 次仅占 8%，安全可控。**

### 成本优化策略

| 策略 | 说明 | 节省 |
|:----:|------|:----:|
| **GM 规则引擎** | 简单行为（点赞/关注/转发）不调 LLM | ~50% GM 调用 |
| **批量社会更新** | 一次 LLM 调用处理多个 Agent 的社会感知更新 | ~60% 社会感知调用 |
| **反思阈值控制** | 只有累积重要性超阈值才触发反思 | 按需触发 |
| **qwen-flash 优先** | 仿真认知/GM 等高频低复杂度用 flash | 降低 token 成本 |
| **预测精炼提前收敛** | 标准差 < 阈值即停止迭代 | 3-4 轮即可收敛 |

应用全部优化后，实际调用量预计 **~350 次**。

---

## 十、项目结构 V2

```
swarm-predict-v2/
├── cmd/server/main.go
├── internal/
│   ├── api/                              # HTTP Handler
│   │   ├── router.go
│   │   ├── graph_handler.go
│   │   ├── simulation_handler.go
│   │   ├── prediction_handler.go         # 🆕 预测精炼 API
│   │   └── report_handler.go
│   ├── openclaw/                         # OpenClaw 客户端
│   │   ├── clients.go
│   │   ├── spawner.go
│   │   └── poller.go
│   ├── cognitive/                        # 🆕 认知引擎
│   │   ├── memory.go                     # 记忆流
│   │   ├── reflection.go                # 反思引擎
│   │   ├── planning.go                  # 计划引擎
│   │   └── social.go                    # 社会感知图
│   ├── agent/                            # 🆕 Agent 架构
│   │   ├── agent.go                     # Agent 核心
│   │   ├── block.go                     # Block 接口
│   │   ├── blocks/                      # 可插拔 Blocks
│   │   │   ├── memory_block.go
│   │   │   ├── reflection_block.go
│   │   │   ├── planning_block.go
│   │   │   ├── social_block.go
│   │   │   ├── posting_block.go
│   │   │   └── reaction_block.go
│   │   └── dispatcher.go               # Block 调度器
│   ├── simulation/                       # 仿真引擎 V2
│   │   ├── engine.go                    # 编排器
│   │   ├── game_master.go              # 🆕 Game Master
│   │   ├── grouper.go
│   │   ├── platform.go
│   │   ├── feed.go
│   │   ├── emergence.go                # 🆕 涌现信号检测
│   │   └── logger.go
│   ├── prediction/                       # 🆕 预测精炼引擎
│   │   ├── delphi.go                    # 德尔菲编排
│   │   ├── expert.go                    # 专家 Agent 定义
│   │   ├── mediator.go                  # 调解 Agent
│   │   └── calibrator.go               # 统计校准
│   ├── react/                            # ReACT 报告引擎
│   │   ├── engine.go
│   │   ├── tools.go
│   │   └── report.go
│   ├── graph/                            # Neo4j 图谱
│   ├── model/
│   └── store/
├── configs/
│   ├── openclaw.json
│   └── experts.yaml                     # 🆕 专家 Agent 配置
├── go.mod
└── docker-compose.yml
```

---

## 十一、部署架构

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
graph TB
    subgraph Docker["Docker Compose"]
        FE["frontend :3000"]
        BE["go-backend :8080"]
        OC["openclaw-gateway :18789"]
        PG["postgres :5432"]
        NEO["neo4j :7687"]
        RD["redis :6379"]
    end

    BL["☁️ 百炼 Coding Plan"]

    FE -->|"/api"| BE
    BE --> PG & NEO & RD
    BE -->|"chatCompletions (Main)<br/>toolsinvoke → spawn (Subagent)"| OC
    OC --> BL

    style Docker fill:#E8F4FD,stroke:#4A90D9,color:#1a1a1a
    style BL fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
```

与 V1 部署拓扑一致，新增模块全在 Go 后端内部，无额外服务。

---

## 十二、实现路线图

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '13px', 'fontFamily': 'Arial Black'}}}%%
gantt
    title MiroFish V2 实现路线图
    dateFormat  YYYY-MM-DD
    axisFormat  %m/%d

    section Phase 1: 基础框架
    Go 项目脚手架 + OpenClaw 集成     :a1, 2026-03-17, 5d
    Neo4j + Redis 数据层              :a2, after a1, 3d
    Agent-Block 架构骨架              :a3, after a1, 4d

    section Phase 2: 认知引擎
    记忆流实现 (Redis Sorted Set)     :b1, after a3, 4d
    反思引擎                          :b2, after b1, 3d
    社会感知图 (Neo4j)                :b3, after b1, 3d
    计划引擎                          :b4, after b2, 2d

    section Phase 3: 仿真引擎V2
    Game Master 实现                  :c1, after b4, 5d
    仿真编排器 (spawn分组)            :c2, after c1, 4d
    涌现信号检测                      :c3, after c2, 3d

    section Phase 4: 预测精炼 (核心)
    德尔菲引擎 + 专家Agent            :d1, after c2, 5d
    调解 Agent                        :d2, after d1, 3d
    统计校准器                        :d3, after d2, 2d

    section Phase 5: 报告 + 前端
    ReACT 报告引擎 V2                 :e1, after d3, 4d
    前端六步流程改造                   :e2, after e1, 5d
    Agent 监控可视化                   :e3, after e2, 4d

    section Phase 6: 测试 + 优化
    端到端测试                        :f1, after e3, 5d
    成本优化调参                      :f2, after f1, 3d
```

### 里程碑

| Phase | 里程碑 | 预计 |
|:-----:|--------|:----:|
| 1 | 基础框架可运行，OpenClaw 连通 | Week 1-2 |
| 2 | 认知 Agent 可记忆、反思、计划 | Week 3-4 |
| 3 | GM 仿真可运行，产出行为日志 | Week 5-6 |
| **4** | **预测精炼可运行，输出校准预测** | **Week 7-8** |
| 5 | 完整六步流程 + 前端 | Week 9-10 |
| 6 | 端到端测试通过，成本达标 | Week 11-12 |

---

## 十三、V1 → V2 → V3 演进路线

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
graph LR
    subgraph V1_Box["V1 (当前)"]
        direction TB
        V1_1["Python Flask"]
        V1_2["OASIS 仿真"]
        V1_3["Zep Cloud GraphRAG"]
        V1_4["ReACT 报告"]
    end

    subgraph V2_Box["V2 (本方案)"]
        direction TB
        V2_1["Go + OpenClaw"]
        V2_2["认知Agent + GM仿真"]
        V2_3["Neo4j + 社会图"]
        V2_4["德尔菲预测精炼"]
        V2_5["ReACT报告 V2"]
    end

    subgraph V3_Box["V3 (远景)"]
        direction TB
        V3_1["分布式 Agent (Actor模型)"]
        V3_2["多模态感知<br/>(图像/视频/音频)"]
        V3_3["预测市场集成<br/>(Polymarket...)"]
        V3_4["实时信息流<br/>(新闻/社媒 API)"]
        V3_5["OpenProse 动态工作流"]
    end

    V1_Box -->|"Go重写+认知升级<br/>+预测精炼"| V2_Box
    V2_Box -->|"分布式+多模态<br/>+实时数据"| V3_Box

    style V1_Box fill:#FFEBEE,stroke:#C62828,color:#1a1a1a
    style V2_Box fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
    style V3_Box fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
```

---

## 十四、总结

```mermaid
%%{init: {'theme': 'default', 'themeVariables': {'fontSize': '14px', 'fontFamily': 'Arial Black'}}}%%
flowchart TB
    subgraph Summary["MiroFish V2 核心升级"]
        direction TB

        subgraph Upgrade1["升级1: Agent认知"]
            U1["记忆流 + 反思 + 计划<br/>(借鉴 Smallville)"]
        end

        subgraph Upgrade2["升级2: 社会感知"]
            U2["声誉/影响力/好感度<br/>(借鉴 Project Sid)"]
        end

        subgraph Upgrade3["升级3: GM仲裁"]
            U3["Game Master 验证行为合理性<br/>(借鉴 Concordia)"]
        end

        subgraph Upgrade4["升级4: 模块化"]
            U4["Agent-Block-Action 可插拔<br/>(借鉴 AgentSociety)"]
        end

        subgraph Upgrade5["升级5: 预测精炼 ⭐"]
            U5["德尔菲法 + 统计校准<br/>(借鉴 DeLLMphi)"]
        end
    end

    Upgrade1 & Upgrade2 & Upgrade3 & Upgrade4 --> Upgrade5

    subgraph Result["🎯 最终效果"]
        R1["预测准确率 ↑↑"]
        R2["Agent 行为真实性 ↑↑"]
        R3["涌现现象丰富度 ↑"]
        R4["成本可控 (~350-481 次/仿真)"]
    end

    Upgrade5 --> Result

    style Summary fill:#E8F4FD,stroke:#4A90D9,color:#1a1a1a
    style Upgrade1 fill:#E3F2FD,stroke:#1976D2,color:#1a1a1a
    style Upgrade2 fill:#FFF3E0,stroke:#F57C00,color:#1a1a1a
    style Upgrade3 fill:#F3E5F5,stroke:#7B1FA2,color:#1a1a1a
    style Upgrade4 fill:#FFFDE7,stroke:#F9A825,color:#1a1a1a
    style Upgrade5 fill:#FCE4EC,stroke:#C62828,color:#1a1a1a
    style Result fill:#E8F5E9,stroke:#388E3C,color:#1a1a1a
```

| 维度 | 值 |
|------|------|
| **全部 LLM 调用** | 通过 OpenClaw Gateway → 百炼 Coding Plan ✅ |
| **Agent 认知** | 三层认知模型（记忆流 + 反思 + 计划 + 社会感知） |
| **环境管理** | Game Master 仲裁，规则引擎 + LLM 双层检查 |
| **预测精炼** | 5 专家 × 4 轮德尔菲法 → Platt 校准 → 极端化修正 |
| **模块化** | Agent-Block-Action，按场景插拔能力模块 |
| **调用量** | 优化后 ~350 次/仿真 (Pro 额度 6000，占 ~6%) |
| **技术栈** | Go + OpenClaw + Neo4j + PostgreSQL + Redis |
| **工期** | ~12 周 (6 Phase) |
