# CryptoSentinel

加密货币定投监控机器人 - 基于多因子策略的 BTC/ETH 投资建议系统

## 功能特性

- **多因子策略引擎**：结合 AHR999、MVRV-Z Score、Pi 周期、MA 乘数等指标
- **风控熔断机制**：杠杆率监控、逃顶信号检测
- **BTC/ETH 独立策略**：针对不同资产的差异化投资建议
- **定时推送**：通过 Telegram 定时发送投资建议报告
- **Docker 部署**：支持容器化部署，非 root 用户运行

## 策略逻辑

### 风控规则（最高优先级）
| 条件 | 动作 |
|------|------|
| 杠杆率 > 1.5 | 熔断停止，禁止买入 |
| Pi 周期死叉 或 MA 突破红线 | 逃顶警报，禁止买入 |

### BTC 策略（基于 AHR999）
| AHR999 区间 | 动作 | 资金倍率 |
|-------------|------|----------|
| < 0.45 | 贪婪买入 | 1.5x |
| 0.45 - 1.20 | 正常定投 | 1.0x |
| 1.20 - 5.00 | 持有观望 | 0x |
| > 5.00 | 逐步卖出 | 0x |

### MVRV 辅助验证
- 若 MVRV-Z Score > 6.0（极度过热），强制改为谨慎持有

### ETH 独立策略（回归带）
| ETH 位置 | 动作 |
|----------|------|
| 低估区 | 重仓买入 |
| 中间区 | 跟随 BTC |
| 高估区 | 卖出/换 BTC |

## 快速开始

### 环境要求
- Go 1.21+
- Docker & Docker Compose（可选）

### 本地运行

1. 克隆项目
```bash
git clone https://github.com/EurekaO-O/CryptoSentinel.git
cd CryptoSentinel
```

2. 安装依赖
```bash
go mod tidy
```

3. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 填入 Telegram Bot Token 和 Chat ID
```

4. 运行
```bash
go run ./cmd/bot
```

### Docker 部署

1. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 填入配置
```

2. 启动容器
```bash
docker-compose up -d
```

3. 查看日志
```bash
docker-compose logs -f
```

## 配置说明

### 环境变量

| 变量名 | 必填 | 说明 | 默认值 |
|--------|------|------|--------|
| TELEGRAM_BOT_TOKEN | 是 | Telegram Bot Token | - |
| TELEGRAM_CHAT_ID | 是 | 目标聊天 ID | - |
| AHR999_URL | 否 | AHR999 数据接口 | 内置默认值 |
| CRON_SPEC | 否 | 定时表达式 | `0 0 9 * * 1` (每周一9点) |
| LEVERAGE | 否 | 当前杠杆率 | 1.0 |
| RUN_ON_START | 否 | 启动时执行一次 | false |
| TZ | 否 | 时区 | Asia/Shanghai |

### Cron 表达式格式
```
秒 分 时 日 月 周
0  0  9  *  *  1   # 每周一上午9点
0  0  */6 *  *  *  # 每6小时
```

## 项目结构

```
CryptoSentinel/
├── cmd/
│   └── bot/
│       └── main.go              # 程序入口
├── configs/
│   └── config.yaml              # 配置模板
├── internal/
│   ├── collector/               # 数据采集模块
│   │   ├── collector.go         # 基础采集器
│   │   └── collector_v2.go      # 多指标采集器
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── model/
│   │   └── indicators.go        # 数据模型定义
│   ├── notifier/
│   │   ├── telegram.go          # Telegram 通知
│   │   └── formatter.go         # 报告格式化
│   └── strategy/
│       ├── strategy.go          # 基础策略
│       ├── strategy_v2.go       # 多因子策略
│       └── *_test.go            # 单元测试
├── Dockerfile                   # Docker 构建文件
├── docker-compose.yml           # Docker Compose 配置
├── .env.example                 # 环境变量示例
└── go.mod
```

## 通知示例

```
🛡️ CryptoSentinel 周报 [2025-01-07]

1. 风控检查
- 杠杆率: 1.00x (✅ 安全)
- 逃顶信号: Pi周期[✅ 正常] / MA状态[正常]

2. 核心指标
- 📏 AHR999: 0.7000 -> 🔵 定投区
- 🌡️ MVRV-Z: 2.50
- 💎 ETH位置: 🟡 中间区

3. 执行建议
- 🚀 BTC 操作: 📈 正常定投
- 💰 倍率: 1.0x
- Ξ ETH 操作: 👉 跟随BTC

"看着资产像树苗一样慢慢长高，本身就是一件枯燥的事情。"
```

## 运行测试

```bash
go test ./... -v
```

## 安全说明

- 敏感信息（Token、Chat ID）通过环境变量传入，不写入代码
- Docker 容器使用非 root 用户运行
- 支持只读文件系统运行
- 资源限制防止容器占用过多资源

## License

MIT
