# nanoserver 通信协议文档

## 协议总览

nanoserver 采用 **双协议层** 架构：

| 协议层            | 传输方式             | 序列化 | 加密           | 用途                   |
| ----------------- | -------------------- | ------ | -------------- | ---------------------- |
| **WebSocket RPC** | WebSocket (Nano框架) | JSON   | XXTEA + Base64 | 实时游戏通信           |
| **HTTP REST**     | HTTP/1.1             | JSON   | 无             | 登录、支付、查询、管理 |

---

## 协议规范

### 命名规范

| 类别              | 规范                                              | 示例                                             |
| ----------------- | ------------------------------------------------- | ------------------------------------------------ |
| **协议结构体**    | PascalCase，以 Request/Response/Info/Message 结尾 | `CreateDeskRequest`, `LoginToGameServerResponse` |
| **推送路由**      | lowerCamelCase，以 "on" 前缀                      | `"onPlayerEnter"`, `"onMoPai"`                   |
| **操作码常量**    | PascalCase，Optype 前缀                           | `OptypeChu`, `OptypePeng`, `OptypeHu`            |
| **番型常量**      | PascalCase，FanXing 前缀                          | `FanXingQingYiSe`, `FanXingQiDui`                |
| **房间/比赛常量** | PascalCase，类型后缀                              | `MatchTypeClassic`, `RoomTypeDailyMatch`         |
| **退出类型**      | PascalCase，ExitType 前缀                         | `ExitTypeSelfRequest`, `ExitTypeDissolve`        |
| **胡牌类型**      | PascalCase，HuType 前缀                           | `HuTypeZiMo`, `HuTypeDianPao`                    |
| **JSON 标签**     | snake_case                                        | `json:"deskId"`, `json:"shouPaiIds"`             |
| **HTTP 路径**     | 小写 + 斜线分段                                   | `/v1/user/login/3rd`, `/v1/order/notify/wechat`  |

### 错误码规范

- **成功**: `code = 0`
- **失败**: `code ≠ 0`，错误信息在 `error` 字段
- 响应格式：`{ "code": 0, "data": {...} }` 或 `{ "code": -1, "error": "..." }`

---

## 一、WebSocket RPC 协议

### 1.1 加密管道

所有 WebSocket 消息经过 Nano 框架的 Pipeline 处理：

1. **入站**（客户端→服务器）：Base64 解码 → XXTEA 解密
2. **出站**（服务器→客户端）：XXTEA 加密 → Base64 编码
3. **密钥**: `"7AEC4MA152BQE9HWQ7KB"`

### 1.2 路由约定

- **请求/响应路由**: `ComponentName.MethodName`（Nano 自动路由）
  - 例：`Manager.Login`、`DeskManager.CreateDesk`
- **推送路由**（服务器→客户端）：lowerCamelCase，以 `on` 前缀
  - 路由常量定义于 `protocol/route.go`：
    - `RouteOpTypeHint = "onOpTypeHint"`
    - `RouteTypeDo = "onOpTypeDo"`

### 1.3 Component: Manager（登录管理）

| 路由                 | 方向  | 请求类型                   | 响应类型                    | 说明           |
| -------------------- | ----- | -------------------------- | --------------------------- | -------------- |
| `Manager.Login`      | C→S→C | `LoginToGameServerRequest` | `LoginToGameServerResponse` | 登录游戏服务器 |
| `Manager.CheckOrder` | C→S→C | `CheckOrderReqeust`        | `CheckOrderResponse`        | 检查订单状态   |

#### `LoginToGameServerRequest`

```json
{
  "name": "玩家昵称",
  "uid": 10001,
  "headUrl": "http://...",
  "sex": 1,
  "fangKa": 10,
  "ip": "127.0.0.1"
}
```

#### `LoginToGameServerResponse`

```json
{
  "uid": 10001,
  "nickname": "玩家昵称",
  "headUrl": "http://...",
  "sex": 1,
  "fangKa": 10
}
```

### 1.4 Component: DeskManager（桌子管理）

#### 请求路由

| 路由                              | 请求类型                     | 响应类型                 | 说明                       |
| --------------------------------- | ---------------------------- | ------------------------ | -------------------------- |
| `DeskManager.CreateDesk`          | `CreateDeskRequest`          | `CreateDeskResponse`     | 创建房间                   |
| `DeskManager.Join`                | `JoinDeskRequest`            | `JoinDeskResponse`       | 加入房间                   |
| `DeskManager.UnCompleteDesk`      | `[]byte`(空)                 | `UnCompleteDeskResponse` | 查询未完成对局             |
| `DeskManager.Ready`               | `[]byte`(空)                 | `error`                  | 准备                       |
| `DeskManager.QiPaiFinished`       | `[]byte`(空)                 | `error`                  | 理牌完成                   |
| `DeskManager.DingQue`             | `DingQue`                    | `error`                  | 定缺                       |
| `DeskManager.OpChoose`            | `OpChooseRequest`            | `error`                  | 选择操作（出/碰/杠/胡/过） |
| `DeskManager.Exit`                | `ExitRequest`                | `error`                  | 退出房间                   |
| `DeskManager.Dissolve`            | `[]byte`(空)                 | `error`                  | 申请解散                   |
| `DeskManager.DissolveStatus`      | `DissolveStatusRequest`      | `error`                  | 解散投票                   |
| `DeskManager.ReConnect`           | `ReConnect`                  | `error`                  | 断线重连                   |
| `DeskManager.ReJoin`              | `ReJoinDeskRequest`          | `ReJoinDeskResponse`     | 重新加入                   |
| `DeskManager.ReEnter`             | `ReEnterDeskRequest`         | `error`                  | 重入未完成对局             |
| `DeskManager.Pause`               | `[]byte`(空)                 | `error`                  | 切后台                     |
| `DeskManager.Resume`              | `[]byte`(空)                 | `error`                  | 回前台                     |
| `DeskManager.VoiceMessage`        | `[]byte`                     | `error`                  | 语音消息                   |
| `DeskManager.RecordingVoice`      | `RecordingVoice`             | `error`                  | 录音完成                   |
| `DeskManager.ClientInitCompleted` | `ClientInitCompletedRequest` | `error`                  | 客户端初始化完成           |

#### 房间选项 (`DeskOptions`)

```json
{
  "mode": 3,
  "maxRound": 8,
  "maxFan": 3,
  "zimo": "",
  "menqing": false,
  "jiangdui": false,
  "jiaxin": false,
  "pengpeng": false,
  "pinghu": false,
  "yaojiu": false
}
```

| 字段       | 类型   | 说明                 |
| ---------- | ------ | -------------------- |
| `mode`     | int    | 模式：3=三人、4=四人 |
| `maxRound` | int    | 最大局数：1/4/8/16   |
| `maxFan`   | int    | 封顶番数             |
| `zimo`     | string | 自摸加番/加底        |
| `menqing`  | bool   | 门清                 |
| `jiangdui` | bool   | 将对                 |
| `jiaxin`   | bool   | 夹心五               |
| `pengpeng` | bool   | 碰碰胡               |
| `pinghu`   | bool   | 平胡                 |
| `yaojiu`   | bool   | 幺九                 |

#### 创建房间 (`CreateDeskRequest` / `CreateDeskResponse`)

```json
// Request
{
  "option": { ... },
  "club": 0
}

// Response
{
  "code": 0,
  "error": "",
  "desk": {
    "deskId": "123456",
    "title": "亲友圈",
    "desc": "血战到底",
    "mode": 3
  },
  "tableInfo": { ... }
}
```

#### 加入房间 (`JoinDeskRequest` / `JoinDeskResponse`)

```json
// Request
{
  "deskNo": "123456"
}

// Response
{
  "code": 0,
  "tableInfo": {
    "deskNo": "123456",
    "createdAt": 1500000000,
    "creator": 10001,
    "title": "...",
    "desc": "...",
    "status": 3,
    "round": 1,
    "mode": 3
  },
  "players": [
    {
      "deskPos": 0,
      "uid": 10001,
      "nickname": "玩家A",
      "isReady": true,
      "sex": 1,
      "isExit": false,
      "headUrl": "http://...",
      "score": 0,
      "ip": "127.0.0.1",
      "offline": false
    }
  ]
}
```

### 1.5 推送事件（服务器→客户端）

| 推送路由                  | 数据类型                 | 说明             |
| ------------------------- | ------------------------ | ---------------- |
| `"onPlayerEnter"`         | `PlayerEnterDesk`        | 玩家进出房间通知 |
| `"onDeskBasicInfo"`       | `DeskBasicInfo`          | 房间基本信息     |
| `"onDuanPai"`             | `DuanPai`                | 发牌完成         |
| `"onDingQueHint"`         | `DingQue`                | 定缺提示         |
| `"onDingQue"`             | `[]QueItem`              | 定缺广播         |
| `"onMoPai"`               | `MoPai`                  | 摸牌             |
| `"onOpTypeHint"`          | `Hint`                   | 操作提示         |
| `"onOpTypeDo"`            | `OpTypeDo`               | 操作执行广播     |
| `"onHuScoreChange"`       | `HuInfo`                 | 胡牌分数变动     |
| `"onGangScoreChange"`     | `GangPaiScoreChange`     | 杠牌分数变动     |
| `"onRoundEnd"`            | `RoundOverStats`         | 单局结束         |
| `"onGameEnd"`             | `DestroyDeskResponse`    | 整场结束         |
| `"onDissolveAgreement"`   | `DissolveResponse`       | 解散通知         |
| `"onDissolveStatus"`      | `DissolveStatusResponse` | 解散状态更新     |
| `"onDissolveFailure"`     | `DissolveResult`         | 解散失败         |
| `"onDissolveSuccess"`     | `EmptyMessage`           | 解散成功         |
| `"onDissolve"`            | `ExitResponse`           | 房间已解散       |
| `"onPlayerExit"`          | `ExitResponse`           | 玩家退出         |
| `"onSyncDesk"`            | `SyncDesk`               | 桌面同步（重连） |
| `"onBroadcast"`           | `StringMessage`          | 系统广播         |
| `"onCoinChange"`          | `CoinChangeInformation`  | 房卡变动         |
| `"onPlayerOfflineStatus"` | `PlayerOfflineStatus`    | 玩家上下线状态   |
| `"onVoiceMessage"`        | `[]byte`                 | 语音消息         |
| `"onRecordingVoice"`      | `PlayRecordingVoice`     | 播放录音         |

### 1.6 核心游戏数据结构

#### 发牌 (`DuanPai`)

```json
{
  "markerID": 10001,
  "dice1": 3,
  "dice2": 5,
  "accountInfo": [
    { "uid": 10001, "onHand": [1, 5, 12, ...] },
    { "uid": 10002, "onHand": [...] }
  ]
}
```

#### 操作提示 (`Hint`)

```json
{
  "uid": 10001,
  "ops": [
    { "type": 2, "tileIDs": [12, 12, 12] },
    { "type": 3, "tileIDs": [12] },
    { "type": 4, "tileIDs": [12] },
    { "type": 5, "tileIDs": [] }
  ],
  "tings": null
}
```

操作类型：1=出牌、2=碰、3=杠、4=胡、5=过

#### 桌面同步（重连）(`SyncDesk`)

```json
{
  "status": 3,
  "players": [
    {
      "uid": 10001,
      "handTiles": [1, 5, 12, ...],
      "chuTiles": [3, 8, ...],
      "pGTiles": [4, 4, 4],
      "latestTile": 12,
      "isHu": false,
      "huPai": 0,
      "huType": 0,
      "que": 0,
      "score": 0
    }
  ],
  "scoreInfo": [...],
  "markerUid": 10001,
  "lastMoPaiUid": 10002,
  "restCount": 48,
  "dice1": 3,
  "dice2": 5,
  "lastTileId": 12,
  "lastChuPaiUid": 10003
}
```

---

## 二、HTTP REST 协议

### 2.1 基础规范

- **基础路径**: 无统一前缀，各端点独立
- **请求格式**: `x-www-form-urlencoded`（GET）或 JSON（POST）
- **响应格式**: JSON `{ "code": 0, "data": {...} }`
- **版本前缀**: `/v1/`

### 2.2 登录接口 (`/v1/user/`)

| 方法 | 路径                   | 说明                 |
| ---- | ---------------------- | -------------------- |
| POST | `/v1/user/login/query` | 查询游客登录是否可用 |
| POST | `/v1/user/login/3rd`   | 第三方（微信）登录   |
| POST | `/v1/user/login/guest` | 游客登录             |
| GET  | `/v1/user/club`        | 查询玩家俱乐部列表   |

#### 微信登录 (`POST /v1/user/login/3rd`)

```json
// Request
{
  "platform": "wechat",
  "appID": "wx...",
  "channelID": "appstore",
  "device": { "imei": "...", "os": "ios", "model": "iPhone", "ip": "...", "remote": "" },
  "name": "微信昵称",
  "openID": "o...",
  "accessToken": "..."
}

// Response
{
  "code": 0,
  "name": "昵称",
  "uid": 10001,
  "headUrl": "http://...",
  "fangKa": 10,
  "sex": 1,
  "ip": "127.0.0.1",
  "port": 33251,
  "playerIP": "客户端IP",
  "config": {
    "version": "1.0.0",
    "iosDownloadURL": "...",
    "androidDownloadURL": "...",
    "heartbeatInterval": 5,
    "share": { "title": "...", "desc": "...", "url": "..." },
    "contact": "客服微信号",
    "voipApp": "...",
    "voipToken": "..."
  },
  "messages": ["公告1", "公告2"],
  "clubList": [
    { "id": 1, "name": "俱乐部A", "balance": 100 }
  ],
  "debug": 0
}
```

### 2.3 订单/支付 (`/v1/order/`)

| 方法 | 路径                      | 说明                |
| ---- | ------------------------- | ------------------- |
| GET  | `/v1/order/`              | 创建支付订单        |
| POST | `/v1/order/notify/wechat` | 微信支付回调（XML） |
| GET  | `/v1/order/console/`      | 管理后台订单列表    |

### 2.4 历史记录 (`/v1/history/`)

| 方法 | 路径                         | 说明         |
| ---- | ---------------------------- | ------------ |
| GET  | `/v1/history/lite/{desk_id}` | 精简历史列表 |
| GET  | `/v1/history/{id}`           | 完整历史详情 |

### 2.5 桌子查询 (`/v1/desk/`)

| 方法 | 路径                   | 说明             |
| ---- | ---------------------- | ---------------- |
| GET  | `/v1/desk/player/{id}` | 查询玩家桌子列表 |
| GET  | `/v1/desk/{id}`        | 查询桌子详情     |

### 2.6 GM 管理 (`/v1/gm/`)

| 方法 | 路径                 | 说明               |
| ---- | -------------------- | ------------------ |
| GET  | `/v1/gm/reset`       | 重置玩家未完成对局 |
| GET  | `/v1/gm/consume`     | 设置房卡消耗配置   |
| GET  | `/v1/gm/broadcast`   | 系统广播           |
| GET  | `/v1/gm/kick`        | 踢玩家下线         |
| GET  | `/v1/gm/online`      | 在线统计           |
| GET  | `/v1/gm/recharge`    | 玩家充值           |
| GET  | `/v1/gm/query/user/` | 查询玩家信息       |

> 注意：GM 接口（除 `/v1/gm/query/user/` 外）仅限本地访问（127.0.0.1）。

### 2.7 统计 (`/v1/stats/`)

| 方法 | 路径                        | 说明         |
| ---- | --------------------------- | ------------ |
| GET  | `/v1/stats/user/register`   | 注册用户统计 |
| GET  | `/v1/stats/user/activation` | 活跃用户统计 |
| GET  | `/v1/stats/online`          | 在线人数统计 |
| GET  | `/v1/stats/retention`       | 留存率统计   |
| GET  | `/v1/stats/consume`         | 房卡消耗统计 |

---

## 三、常量定义

### 3.1 操作类型 (`Optype`)

| 常量             | 值   | 说明     |
| ---------------- | ---- | -------- |
| `OptypeIllegal`  | 0    | 非法操作 |
| `OptypeChu`      | 1    | 出牌     |
| `OptypePeng`     | 2    | 碰       |
| `OptypeGang`     | 3    | 杠       |
| `OptypeHu`       | 4    | 胡       |
| `OptypePass`     | 5    | 过       |
| `OptyMoPai`      | 500  | 摸牌     |
| `OptypeAnGang`   | 1004 | 暗杠     |
| `OptypeMingGang` | 1014 | 明杠     |
| `OptypeBaGang`   | 1024 | 巴杠     |

### 3.2 胡牌类型 (`HuPaiType`)

| 常量            | 值  | 说明 |
| --------------- | --- | ---- |
| `HuTypeDianPao` | 0   | 点炮 |
| `HuTypeZiMo`    | 1   | 自摸 |
| `HuTypePei`     | 2   | 赔付 |

### 3.3 房间状态 (`DeskStatus`)

| 常量                     | 值  | 说明           |
| ------------------------ | --- | -------------- |
| `DeskStatusCreate`       | 0   | 创建           |
| `DeskStatusZb`           | 1   | 准备中         |
| `DeskStatusDq`           | 2   | 定缺           |
| `DeskStatusPlaying`      | 3   | 游戏中         |
| `DeskStatusEnded`        | 4   | 结束           |
| `DeskStatusDuanPai`      | 5   | 发牌           |
| `DeskStatusQiPai`        | 6   | 理牌           |
| `DeskStatusRoundOver`    | 7   | 单局结束       |
| `DeskStatusCleaned`      | 8   | 清理           |
| `DeskStatusInterruption` | 9   | 中断（解散中） |
| `DeskStatusDestory`      | 10  | 已销毁         |

### 3.4 番型 (`FanXing`)

| 常量                   | 值  | 说明     |
| ---------------------- | --- | -------- |
| `FanXingQingYiSe`      | 1   | 清一色   |
| `FanXingQingQiDui`     | 2   | 清七对   |
| `FanXingQingDaDui`     | 3   | 清大对   |
| `FanXingQingDaiYao`    | 4   | 清带幺   |
| `FanXingQingJiangDui`  | 5   | 清将对   |
| `FanXingSuFan`         | 6   | 素番     |
| `FanXingQiDui`         | 7   | 七对     |
| `FanXingDaDui`         | 8   | 大对子   |
| `FanXingQuanDaiYao`    | 9   | 全带幺   |
| `FanXingJiangDui`      | 10  | 将对     |
| `FanXingYaoJiuQiDui`   | 11  | 幺九七对 |
| `FanXingQingLongQiDui` | 12  | 清龙七对 |
| `FanXingLongQiDui`     | 13  | 龙七对   |

### 3.5 退出类型 (`ExitType`)

| 常量                           | 值  | 说明         |
| ------------------------------ | --- | ------------ |
| `ExitTypeExitDeskUI`           | -1  | 退出房间界面 |
| `ExitTypeSelfRequest`          | 0   | 自己请求退出 |
| `ExitTypeClassicCoinNotEnough` | 1   | 金币不足     |
| `ExitTypeDailyMatchEnd`        | 2   | 每日比赛结束 |
| `ExitTypeNotReadyForStart`     | 3   | 未准备开始   |
| `ExitTypeChangeDesk`           | 4   | 换桌         |
| `ExitTypeRepeatLogin`          | 5   | 重复登录     |
| `ExitTypeDissolve`             | 6   | 房间解散     |

---

## 四、协议文件索引

所有协议定义位于 `protocol/` 目录：

| 文件                  | 内容               |
| --------------------- | ------------------ |
| `protocol/route.go`   | 路由常量           |
| `protocol/const.go`   | 所有枚举常量       |
| `protocol/common.go`  | 通用共享类型       |
| `protocol/login.go`   | 登录相关协议       |
| `protocol/desk.go`    | 桌子、游戏操作协议 |
| `protocol/req.go`     | 请求/响应类型      |
| `protocol/club.go`    | 俱乐部协议         |
| `protocol/order.go`   | 订单/支付协议      |
| `protocol/history.go` | 历史记录协议       |
| `protocol/stats.go`   | 统计协议           |
| `protocol/users.go`   | 用户管理协议       |
| `protocol/agent.go`   | 代理协议           |
| `protocol/apps.go`    | 应用管理协议       |
| `protocol/web.go`     | 版本信息           |
| `protocol/test.go`    | 测试协议           |
