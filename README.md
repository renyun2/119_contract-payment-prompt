# 合同台账与付款节点提醒系统

采购/销售/服务/工程合同集中登记、多阶段**付款或收款**计划、到期颜色预警、实际流水核销，以及**销售合同**逾期催收跟进与统计报表。通过 **Docker Compose** 一键拉起 **PostgreSQL 16 + Go API + Vue 前端（Nginx）**，浏览器访问：**http://localhost:8080**（若本机 `8080` 已被占用，请修改 `docker-compose.yml` 中 `frontend` 的端口映射）。

---

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.23、Gin、GORM、PostgreSQL 驱动 |
| 前端 | Vue 3、Vite 5、Vue Router、Element Plus、ECharts、Axios |
| 数据库 | PostgreSQL 16（Alpine 镜像） |
| 部署 | 多阶段 Dockerfile + Docker Compose |

---

## 一键启动

```bash
cd 119_contract-payment
docker compose up -d --build
```

- 前端（Nginx）：映射 **`8080:80`**，静态资源并反向代理 **`/api/`** 到后端。  
- API：容器内监听 **`:8080`**（仅内部网络，由前端服务反代）。  
- 数据库：初始化时自动执行项目根目录 **`init.sql`**（建表 + **50 份合同** + 多节点 + 历史付款 + 部分催收记录）。

健康检查：**GET** `http://localhost:8080/api/health`

### 清空演示数据并重建

数据在卷 `contract_pay_pg` 中持久化；需要完全重来时：

```bash
docker compose down -v
docker compose up -d --build
```

---

## 业务功能（与需求对应）

1. **合同录入**：合同编号、名称、签订日、类型（`purchase`/`sales`/`service`/`engineering`）、对方单位、总金额、期限、摘要、状态（`active` 履行中、`completed` 已完成、`terminated` 已终止）。
2. **付款/收款节点**：每份合同多节点（示例：预付款/进度款/尾款/质保金）；字段含节点名、触发条件、金额、计划日、是否已触发、是否已付（核销后）。
3. **到期预警**：**应付**节点在计划付款日前 **15 天**橙色、**7 天及以内（含逾期）**红色；**销售（应收）**节点**逾期未收回**红色，未到期沿用 15/7 天梯度。工作台汇总**本月**内到期且**未核销**的节点清单。
4. **实际付款 / 回款流水**：录入合同 + 节点 + 实际付款日 + 金额 + 银行流水号 + 付款账户；保存后节点状态置为**已付**并生成 `actual_payments` 记录。
5. **应收与催收**：销售合同提供「催收跟进」页签：**跟进人、日期、沟通内容、承诺付款日**（可选关联节点）。
6. **合同报表**：合同总数/履行中/已完成；本月应付侧「计划未付 / 本期已付款」与应收侧「计划未回 / 本期已回款」；按类型的合同金额分布（饼图）；**当年**销售收款与非销售付款对比（柱状图）。

---

## 数据表（`init.sql`）

- `contracts` — 合同主表  
- `payment_nodes` — 付款/收款计划节点  
- `actual_payments` — 实际付款凭证  
- `collection_followups` — 催收跟进记录  

---

## 本地开发（可选）

### 后端

```bash
cd backend
go run ./cmd/server
```

默认在未设置 `DATABASE_DSN` 时尝试连接：`127.0.0.1:5432`、库名 `contract_payment`、用户/密码见 `DATABASE_DSN` 文档串（与 Compose 一致）。表结构请以 **`init.sql`** 为准；生产环境建议使用迁移工具统一管理。

### 前端

```bash
cd frontend
npm install
npm run dev
```

`vite.config.js` 已将 **`/api` 代理到 `http://127.0.0.1:8080`**（与本地后端默认端口一致）。

---

## 故障排除

构建镜像时若 **`go mod download` 出现网络流错误**，多为 Go 代理瞬时故障，请直接重试 `docker compose build`；或在 `backend/Dockerfile` 构建阶段增加适合本机网络环境的 `GOPROXY`、使用公司内网镜像等。

---

## 生产与安全建议

- 修改数据库口令，勿使用演示账号；  
- 在上层网关启用 HTTPS；  
- 按需增加登录鉴权与审计字段；  
- 大规模环境请拆分迁移与 CI/CD，替代一次性 `init.sql` 写入。

---

## 目录结构（摘要）

```
119_contract-payment/
├── docker-compose.yml
├── init.sql                 # Schema + 50 份演示合同与种子数据
├── README.md
├── backend/
│   ├── Dockerfile
│   ├── cmd/server/main.go
│   └── internal/…          # models / database / handlers
└── frontend/
    ├── Dockerfile
    ├── nginx.conf            # /api 反代
    └── src/                  # Vue 页面与路由
```
