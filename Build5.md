# FWAlizer 功能构建计划（Build5：文档体系重构）

> **文档定位：** 本文档是 FWAlizer 的**当前构建方案**（依据 AGENTS.md §12.1：Build 文档为详细构建方案，非强规则），承接已存档的 [Build1-4](HistoryDocs/)（见 `HistoryDocs/`，仅核查）。
> - 设计记录：[Design4.md](./Design4.md)（当前设计记录；与 AGENTS.md 或用户决策冲突时以用户确认为准）
> - 编码指令：[AGENTS.md](./AGENTS.md)（**唯一强要求**：简单轻量化、不过度防御、内部使用导向、中文注释、log/slog、增量添加+精确删除）
> - 问题追踪：[Issue4.md](./Issue4.md)（当前问题记录）
> - 历史构建与问题记录：见 [HistoryDocs/](./HistoryDocs/)（Build1-4、Issue1-3、Design1-3，均已存档，仅核查）
>
> **执行原则（与 Build1-4 一致）：**
> - 每一步完成后均可编译、可测试。不跳步、不并行多步。
> - AI 执行指令：每次仅执行一个 Step，完成后运行验收命令，确认通过后再进入下一步。
> - **排序原则：先修复后构建、先安全后优化、先依赖后独立**。
> - 每步的新增逻辑必须配套单元测试（用户决策）。

---

## 一、构建进度追踪

| Step | 内容 | 设计依据 | 状态 |
|------|------|---------|------|
| 1 | 文档体系重构（HistoryDocs 迁移 + Design4/Build5/Issue4 创建 + AGENTS/README/.env.example/docker-compose.yml.example 更新） | 用户决策（2026-08-02） | ✅ 验收通过（2026-08-02：10 文档存档横幅齐全；AGENTS/README 引用核验无失效；go build + npm run build 全绿） |
| 2 | 后续候选构建项（见五、候选列表，待用户决策后逐项转为 Step） | Design4 §三 | ☐ 未开始 |

> 状态标记：☐ 未开始 / ◧ 进行中 / ✅ 验收通过

---

## 二、构建概要（文件清单总览）

| Step | 涉及文件 | 要点 |
|------|---------|------|
| 1 | `HistoryDocs/`（新建，移入 Build1-4/Issue1-3/Design1-3 共 10 个文档并加存档横幅）、`Design4.md`/`Build5.md`/`Issue4.md`（新建）、`AGENTS.md`、`README.md`、`.env.example`、`docker-compose.yml.example` | 文档体系迁移与重建：根目录仅保留活跃文档；AGENTS/README 引用更新与 Documents 目录说明；README 改为 WebUI 优先介绍 |

---

## 三、构建顺序依赖图

```
Step 1 (文档体系重构) ── 一次性完成（目录迁移 → 存档横幅 → 新文档 → 引用更新 → 用户文档）
Step 2 (候选构建项) ── 依赖用户对 Design4 §三 候选的决策，逐项转 Step 后执行
```

---

## 四、分步构建计划

---

### Step 1：文档体系重构

**目标：** 将历史文档（Build1-4、Issue1-3、Design1-3）移入 `HistoryDocs/` 并标注存档；创建 Design4/Build5/Issue4 作为当前活跃文档；更新 AGENTS.md 引用体系与 Documents 目录说明；README 改为 WebUI 模式优先介绍（零基础友好），同步更新 .env.example 与 docker-compose.yml.example。

**前置条件：** 无

**产出文件与操作：**

#### 1.1 历史文档归档

- 新建 `HistoryDocs/` 目录，移入 10 个文档：Build1-4、Issue1-3、Design1-3；
- 每个文档标题下插入存档横幅：

```
> ⚠️ **已存档**：本文档已移入 `HistoryDocs/`，仅作历史记录与核查参考，**不再用于构建**。
> 当前活跃文档：编码指令 [AGENTS.md](../AGENTS.md)、设计记录 [Design4.md](../Design4.md)、构建方案 [Build5.md](../Build5.md)、问题记录 [Issue4.md](../Issue4.md)。
```

#### 1.2 新建当前活跃文档

- `Design4.md`：当前设计记录（功能全景 + 设计决策记录 + 后续候选 + 变更流程）；
- `Build5.md`：当前构建方案（本文档，Step 1 记录本次重构）；
- `Issue4.md`：当前问题记录（进行中问题 + 已知遗留/候选事项）。

#### 1.3 AGENTS.md 引用体系更新

- 文件头引用块、§一 文档定位、§12.1 文档定位表、§12.2 文档清单全部改为：当前文档（Design4/Build5/Issue4）+ `HistoryDocs/` 汇总引用；
- 新增 `Documents/` 目录说明：存放各云平台 API 使用要求文档（参数格式、字段长度限制、频率限制）。

#### 1.4 README.md 重写（WebUI 优先）

- 结构改为：简介 → 快速上手（WebUI 模式分步引导，零基础）→ 运行模式（WebUI 推荐 / .env 备用进阶）→ 配置说明 → RULES 语法 → 备份恢复 → 告警 → 多云 API 权限与 Documents 说明 → Docker（WebUI 优先）→ 开发指南 → FAQ；
- 补充 `Documents/` 目录介绍。

#### 1.5 环境示例文件更新

- `.env.example`：标注 .env 模式为备用模式；`SYNC_ENABLED` 注释同步「模拟测试」措辞；
- `docker-compose.yml.example`：标注 WebUI 模式为默认推荐、.env 模式为备用。

**测试与验收：**
```bash
# 1. 目录结构正确：根目录仅剩活跃文档
ls *.md        # 应含 AGENTS.md / Design4.md / Build5.md / Issue4.md / README.md / FutureDesktopDevelop.md
ls HistoryDocs/ # 应含 Build1-4 / Issue1-3 / Design1-3 共 10 个
# 2. 存档文档横幅齐全（10 个文件均含"已存档"）
grep -l "已存档" HistoryDocs/*.md | wc -l   # 应为 10
# 3. AGENTS.md 无对根目录旧文档的失效引用（仅 HistoryDocs/ 路径）
grep -n "HistoryDocs\|Design4\|Build5\|Issue4" AGENTS.md
# 4. README 无失效引用；文档链接均指向存在的文件
grep -n "\./\(Design[1-3]\|Build[1-4]\|Issue[1-3]\)\.md" README.md AGENTS.md || echo "OK: 无失效引用"
# 5. 项目构建不受影响
go build ./... && cd webui/frontend && npm run build
```

---

## 五、候选构建项（待用户决策，逐项转 Step）

| # | 候选 | 说明 | 来源 |
|---|------|------|------|
| 1 | DNS 失败写入历史记录 | `EventDNSFailed` 落库（`result=failed`），提升故障可见性；需 `main.go` 增加订阅 | Design4 §三-1 |
| 2 | 日志页「暂停输出」按钮 | 高日志量时前端暂停渲染 | Design4 §三-2 |
| 3 | 主题设置项加入全局设置页 | 复用 `theme` 键，双入口管理 | Design4 §三-3 |
| 4 | 运行测试「上次执行」持久化 | 刷新后保留结果（当前内存态） | Design4 §三-4 |

> 候选转 Step 流程：用户确认 → 在本文件追加 Step（含目标/前置/参考代码/验收命令）→ 按序执行。

---

## 六、变更记录

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2026-08-02 | 文档体系重构：Build5 创建（承接已存档 Build1-4），Step 1 记录文档体系重构 |
| v1.1 | 2026-08-02 | **Step 1 验收通过**：HistoryDocs 迁移完成（10 文档 + 存档横幅）、Design4/Build5/Issue4 创建、AGENTS/README 引用与 Documents 说明更新、README WebUI 优先重写、.env.example/docker-compose.yml.example 同步 |
