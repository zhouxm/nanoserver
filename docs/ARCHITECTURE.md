# nanoserver 项目架构分析

## 项目概述

**nanoserver** 是一个用 Go 编写的**四川麻将"血战到底"游戏服务器**，支持三人/四人两种玩法。本项目作为 [`Nano`](https://github.com/lonnng/nano) 游戏服务器框架的示例开源，采用房卡（room card）模式运营，支持俱乐部功能。

**不包含客户端逻辑**。本项目是纯后端服务器，客户端为独立的 iOS/Android 原生应用，与本仓库分开发布。

---

## 目录结构

```
nanoserver/
├── back/                    # 独立的后台管理模块（独立 go.mod）
├── build/                   # 构建脚本/资源
├── cmd/mahjong/             # 主入口（mahjong 服务器二进制）
│   ├── main.go              # 启动入口
│   └── game/                # 核心游戏服务器（Nano 组件）
│       ├── game.go          # 游戏服务器初始化
│       ├── manager.go       # Manager 组件：玩家/会话管理
│       ├── desk_manager.go  # DeskManager 组件：桌子/房间管理
│       ├── desk.go          # Desk 结构体 & 游戏主循环
│       ├── player.go        # Player 结构体 & 玩家操作
│       ├── club_manager.go  # ClubManager 组件：俱乐部管理
│       ├── rest_api.go      # GM 管理接口（踢人/广播/充值等）
│       ├── dissolve_context.go # 解散房间状态机
│       ├── prepare_context.go  # 准备/定缺状态
│       ├── crypto.go        # XXTEA 加密管道
│       ├── dice.go          # 骰子
│       ├── constants.go     # 游戏常量
│       ├── helper.go        # 工具函数
│       └── mahjong/         # 麻将核心引擎
│           ├── tile.go      # 牌表示
│           ├── mahjong.go   # 牌组操作
│           ├── meta.go      # 核心类型
│           ├── algorithm.go # 胡牌检测
│           ├── indexes.go   # 索引标记系统
│           ├── base.go      # 番数计算 & 计分
│           └── heler.go     # 辅助算法
│       └── history/         # 游戏回放/快照
│           └── history.go
│   └── web/                 # HTTP Web 服务器
│       ├── web.go           # 路由注册 & 启动
│       ├── gm.go            # GM 管理接口
│       ├── stats.go         # 统计接口
│       └── api/             # REST API 实现
│           ├── login.go     # 登录（微信/游客）
│           ├── order.go     # 订单 & 支付
│           ├── desk.go      # 桌子查询
│           ├── history.go   # 历史记录查询
│           └── provider/    # 第三方服务集成
│               └── wechat.go # 微信支付
├── configs/                 # TOML 配置文件
├── db/                      # 数据库层
│   ├── model.go             # XORM 引擎初始化
│   ├── user.go              # 用户 CRUD
│   ├── desk.go              # 桌子 CRUD
│   ├── history.go           # 历史记录 CRUD
│   ├── club.go              # 俱乐部 CRUD
│   ├── order.go             # 订单 CRUD
│   ├── trade.go             # 交易 CRUD
│   ├── consume.go           # 房卡消耗统计
│   ├── online.go            # 在线统计
│   ├── third_account.go     # 第三方账号
│   ├── logger.go            # XORM 日志适配器
│   ├── common.go            # SQL 工具函数
│   ├── const.go             # 数据库常量
│   └── model/               # 数据模型定义
│       └── struct.go
├── protocol/                # 通信协议定义（结构体）
│   ├── login.go             # 登录协议
│   ├── desk.go              # 桌子协议
│   ├── club.go              # 俱乐部协议
│   ├── order.go             # 订单协议
│   ├── history.go           # 历史协议
│   ├── stats.go             # 统计协议
│   ├── route.go             # 路由常量
│   ├── const.go             # 操作码常量
│   ├── common.go            # 通用类型
│   ├── req.go               # 请求/响应类型
│   └── ...                  # 其他协议定义
├── pkg/                     # 共享工具包
│   ├── algoutil/            # 算法工具（MD5/RSA/HTTP/编码）
│   ├── async/               # 异步任务运行器
│   ├── constant/            # 游戏常量
│   ├── crypto/              # RSA 加解密
│   ├── errutil/             # 错误码定义
│   ├── room/                # 房间号生成
│   ├── security/            # 验证工具
│   ├── set/                 # 线程安全集合
│   └── whitelist/           # IP 白名单
├── docs/                    # 文档
├── tools/                   # 工具
└── dist/                    # 构建输出
```

---

## 架构总览

```
┌──────────────────────────────────────────────────────────────────┐
│                        nanoserver                                 │
│                                                                   │
│  ┌───────────────────────────┐  ┌──────────────────────────────┐  │
│  │   Game Server (Nano)      │  │   HTTP Web Server (nex)      │  │
│  │   WebSocket Port 33251    │  │   HTTP Port 12307             │  │
│  │                           │  │                               │  │
│  │  Component: Manager       │  │  POST /v1/user/login/...     │  │
│  │  Component: DeskManager   │  │  POST /v1/order/...          │  │
│  │  Component: ClubManager   │  │  GET  /v1/history/...        │  │
│  │                           │  │  GET  /v1/desk/...           │  │
│  │  加密管道: XXTEA+Base64   │  │  GET  /v1/gm/... (管理)      │  │
│  │  序列化: JSON             │  │  GET  /v1/stats/... (统计)   │  │
│  └───────────┬───────────────┘  └───────────────┬──────────────┘  │
│              │                                   │                │
│              │         ┌─────────────────┐       │                │
│              └────────>│   MySQL 数据库  │<──────┘                │
│                        │   (XORM ORM)    │                       │
│                        └─────────────────┘                        │
└──────────────────────────────────────────────────────────────────┘
         │                          │
   WebSocket                   HTTP/JSON
         │                          │
  ┌───────────┐           ┌──────────────────┐
  │ 移动客户端 │           │   管理后台/GM     │
  │ (iOS/Android)          │   (浏览器/脚本)   │
  └───────────┘           └──────────────────┘
```

### 进程架构

- 一个单进程运行两个服务器：WebSocket 游戏服务器 + HTTP Web 服务器
- 两者通过共享内存（全局变量/Channel）进行内部通信
- `sync.WaitGroup` 确保任一服务器崩溃时进程退出

### 并发模型

- **每桌一个 goroutine**：`desk.play()` 方法在自己的 goroutine 中运行，处理该桌所有游戏逻辑
- **玩家操作通过 Channel 通信**：`player.chOperation` 传递用户操作（出牌/碰/杠/胡/过）
- **异步数据库写入**：通过带缓冲的 Channel (`chWrite`/`chUpdate`) 异步写入
- **Nano 框架**处理 WebSocket 连接、会话管理和路由分发

---

## 通信协议

### WebSocket（实时游戏）

- 基于 Nano 框架的 RPC 风格通信
- 加密管道：XXTEA 加密 → Base64 编码（入站反向）
- 序列化：JSON
- 心跳：Nano 内置心跳机制

### HTTP REST（业务接口）

- 基于 `nex` + `gorilla/mux` 框架
- 接口包括：登录、订单支付、历史查询、GM 管理、统计

---

## 功能清单

### 游戏核心
1. 四川麻将"血战到底"规则实现（胡牌后继续，直至只剩一人未胡）
2. 三人模式（72 张牌，无万）和四人模式（108 张牌）
3. 完整牌操作：摸牌、出牌、碰、杠（明杠/暗杠/补杠）、胡（自摸/点炮）
4. 复杂番型计算：清一色、大对子、七对（龙七对/双龙七对/豪华龙七对）、将对、金钩钩、幺九、中张、夹心五等
5. 番数上限配置（MaxFan）
6. 杠上花/杠上炮/抢杠胡
7. 转雨（杠后点炮，杠分转移）
8. 过手胡检测（guo-shou hu）
9. 查叫/赔叫（非听牌玩家赔偿）
10. 双响炮

### 房间系统
11. 创建房间/加入房间（房间号 6 位数字）
12. 房卡消耗模式（配置映射：局数→房卡数）
13. 俱乐部开房（俱乐部独立房卡余额）
14. 解散房间（投票超时机制，300s→150s）
15. 准备/定缺流程
16. 计时器：每 5 分钟清理已销毁房间

### 用户系统
17. 微信登录（OAuth2）
18. 游客登录（IMEI+AppID）
19. 微信支付（统一下单、回调处理）
20. 房卡充值

### 网络功能
21. 断线重连（网络切换/杀进程/关机均可恢复）
22. 前后台切换检测
23. 语音消息转发
24. 系统广播消息
25. 热更补丁下载

### 管理功能
26. GM 管理接口（踢人、重置、充值、广播、房卡配置）
27. IP 白名单（GM 接口仅限本地访问）
28. 统计数据：注册/活跃/留存/在线/房卡消耗

### 数据功能
29. 首次运行自动建表
30. 游戏回放（JSON 快照录制）
31. 结构化日志（logrus）
32. 异步数据库写入

---

## 数据库模型

主要表（XORM ORM，自动建表）：

| 表名 | 用途 |
|------|------|
| `user` | 用户账号（房卡、角色、状态、密钥） |
| `desk` | 游戏桌子记录（玩家、分数、模式、局数） |
| `history` | 游戏历史（JSON 快照 + 元信息） |
| `order` | 支付订单 |
| `trade` | 交易记录 |
| `recharge` | 充值记录 |
| `register` | 注册记录 |
| `login` | 登录日志 |
| `online` | 在线人数快照 |
| `third_account` | 第三方账号绑定 |
| `card_consume` | 房卡消耗记录 |
| `club` | 俱乐部（余额、名称、人数上限） |
| `user_club` | 俱乐部成员（申请/同意状态） |
| `agent` | 代理（房卡数、折扣、等级） |
| `uuid` | UUID 映射 |

---

## 主要依赖

| 库 | 用途 |
|----|------|
| `github.com/lonng/nano` | 游戏服务器框架（WebSocket/会话/组件） |
| `github.com/lonng/nex` | HTTP REST 框架 |
| `xorm.io/xorm` | ORM 数据库操作 |
| `github.com/go-sql-driver/mysql` | MySQL 驱动 |
| `github.com/gorilla/mux` | HTTP 路由 |
| `github.com/gorilla/websocket` | WebSocket （Nano 内部使用） |
| `github.com/spf13/viper` | 配置管理（TOML） |
| `github.com/urfave/cli` | 命令行参数 |
| `github.com/sirupsen/logrus` | 结构化日志 |
| `github.com/xxtea/xxtea-go` | 网络流量加密 |
| `gopkg.in/chanxuehong/wechat.v2` | 微信登录/支付 |
| `golang.org/x/crypto` | RSA 加解密 |
