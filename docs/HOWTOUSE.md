## ipaas-agent 使用指南

这份文档是"怎么把它跑起来并确认它在工作"的完整说明。所有行为描述都对应仓库里的实际代码
（含函数名与文件位置），如果和 README 有冲突，以最后的"与 README 的差异"一节为准。

- 适用版本：v0.9.11 及之后的 main 分支
- 运行形态：单个 Go 可执行文件，无命令行参数，无图形界面
- 支持平台：linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64

---

## 1. 它是什么，解决什么问题

`ipaas-agent`（内部叫"本地网关"）是装在**你自己的内网里**的一个常驻进程。它主动向外
发起一条到钉钉开放平台的 WebSocket 长连接，钉钉的连接平台（iPaaS）需要访问你内网的
HTTP 接口或数据库时，把请求顺着这条已有的连接送进来，由网关在本地执行后原路返回。

因此：

- 你的内网**不需要对公网开放任何端口**，只需要能出网访问 `https://api.dingtalk.com`。
- 数据库的账号密码、内网地址**只留在本地配置文件里**，不随请求发到平台。
- 网关本身不做鉴权、不做缓存、不做格式转换，它是一个"执行器"：收到什么动作就在本地执行什么。

数据流概括：

```text
连接平台的动作 ──(钉钉云端)──> 长连接 ──> 本地网关 ──> 内网 HTTP / MySQL / PG / SQL Server / Oracle
                                  <───────── 执行结果 ─────────┘
```

---

## 2. 开始之前需要准备的东西

1. **一个钉钉开放平台应用**，拿到 `Client ID`（AppKey）和 `Client Secret`（AppSecret）。
   在[钉钉开发者后台](https://open-dev.dingtalk.com/#/)创建，应用需要开通 Stream 模式
   相关的连接能力。
2. **一台能常驻运行的机器**（Linux/Windows/macOS 都行），要求：
   - 能出网访问 `api.dingtalk.com:443`；
   - 能连通你要代理的那些内网地址和端口；
   - 一个可写目录（网关会在运行目录里读配置、写日志）。
3. **连接平台侧的一个连接器/动作**，用来触发到网关的调用。这部分在钉钉控制台操作，
   入口和字段名以控制台为准，本文档只说明网关这一侧会收到什么、怎么回应。

如果这台机器只能经代理出网，还需要一个可用的 HTTP 或 SOCKS5 代理，见 4.3。

---

## 3. 安装

### 3.1 直接下载发布产物（推荐）

从 [Releases](https://github.com/yhan1203/ipaas-agent/releases) 下载对应平台的包，
解压到一个固定目录，例如：

```text
/opt/ipaas-agent/
├── ipaas-agent              # 或 ipaas-agent.exe
└── config.yaml              # 需要你自己创建
```

发布产物命名为 `ipaas-agent-<版本号>-<goos>-<goarch>`，Windows 带 `.exe` 后缀。

### 3.2 从源码构建

```bash
git clone https://github.com/yhan1203/ipaas-agent.git
cd ipaas-agent
go build -o ipaas-agent .
```

带上版本信息（发布流水线用的就是这条，见 `.github/workflows/action.yml`）：

```bash
go build -ldflags "-X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -X main.GitCommit=$(git rev-parse --short HEAD) \
  -X main.Version=v0.9.11" -o ipaas-agent .
```

不传 `-ldflags` 时三个值都是 `DEV`，只影响日志里打印的版本号，不影响功能。

### 3.3 一个必须知道的坑：一切路径都以"进程工作目录"为准

这是最常见的启动失败原因。代码里：

- 配置查找用的是相对路径 `.` 和 `./config`（`pkg/config/load.go` 的 `LoadConfig`）；
- 日志用的是相对路径 `log1.txt`、`log2.txt`（`pkg/logger/logger.go`）。

两者都相对于**进程的当前工作目录**，不是相对于可执行文件所在目录。也就是说：

```bash
cd /opt/ipaas-agent && ./ipaas-agent      # 正确：能找到 config.yaml
/opt/ipaas-agent/ipaas-agent              # 错误：会在当前目录下找配置
```

第二种写法会报 `读取配置文件出错: Config File "config" Not Found`。用 systemd、任务计划
程序、容器等托管启动时，务必显式设置工作目录（见第 8 节）。

---

## 4. 配置文件

### 4.1 文件名与查找顺序

- 固定叫 `config`，扩展名 `.yaml` 或 `.yml` 都能被加载（实测两者均可，建议统一用 `config.yaml`）。
- 查找顺序：先当前工作目录，再 `./config` 子目录。
- **没有命令行参数**，不能用 `-c` 指定配置路径，文件名也不可改。

### 4.2 完整示例

下面是一份"什么都能跑通"的配置，实际使用时删掉用不到的部分即可：

```yaml
# ---------- 认证：必填 ----------
auth:
  clientID: dingeypapxxxxxxxxxxx
  clientSecret: xxxxxxxxxx7mQjlIF7q6YiFitxxxxxxxxxxxxxxxxxxxx
  # 以下两项都可省略
  openAPIHost: https://api.dingtalk.com   # 专有云/预发环境才需要改
  proxy: ""                               # 只能经代理出网时填写
  # 以下各库的 allow_remote 决定"能否由平台临时下发连接参数"，都可省略，见 4.6
  mysql:
    allow_remote: false
  pgsql:
    allow_remote: false
  mssql:
    allow_remote: false
    # SQL Server 的兼容性参数，原样拼进连接串尾部
    less_common_parameters: "encrypt=disable;trustServerCertificate=true"
  oracledb:
    allow_remote: false

# ---------- 超时兜底：可省略，单位毫秒 ----------
timeout:
  http_default_ms: 30000   # 平台没下发 timeout 时，一次 HTTP 动作允许的时长
  http_max_ms: 0           # 对平台下发的 timeout 做上限截断，0 表示不限制
  db_ms: 60000             # 一次数据库动作（连接 + 查询）的总时长

# ---------- 数据源：按需 ----------
plugins:
  mysql:
    - host: 127.0.0.1
      port: 3306
      user: root
      password: root
      database: example
      config_key: default
  pgsql:
    - host: 127.0.0.1
      port: 5432
      user: postgres
      password: postgres
      database: example
      config_key: pg-main
  mssql:
    - host: 127.0.0.1
      port: 1433
      user: sa
      password: sa123456A
      database: TestDB
      config_key: sqlServer
  oracledb:
    - host: 127.0.0.1
      port: 1521
      user: system
      password: example
      service_name: FREE    # 与 sid 二选一
      config_key: oracle-main
```

这份示例可以直接整份复制使用：注释只是标注用途，删掉用不到的数据源即可。注意 YAML 里
顶层键不可重复，`auth` 下的各库开关必须像上面那样写在同一个 `auth:` 块内。

### 4.3 `auth` 认证

| 键 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `auth.clientID` | 是 | 无 | 应用的 Client ID / AppKey |
| `auth.clientSecret` | 是 | 无 | 应用的 Client Secret / AppSecret |
| `auth.openAPIHost` | 否 | `https://api.dingtalk.com` | 开放平台地址，专有云或预发环境才需要改 |
| `auth.proxy` | 否 | 空（直连） | 出网代理，如 `http://127.0.0.1:7890`、`socks5://127.0.0.1:1080` |
| `auth.<库名>.allow_remote` | 否 | `false` | 是否允许平台临时下发连接参数，见 4.6 |
| `auth.mssql.less_common_parameters` | 否 | 空 | 追加到 SQL Server 连接串尾部的参数 |

两个兼容写法，读代码时容易困惑，这里说明清楚（`pkg/config/load.go` 的 `GetAuthClientConfig`）：

- `client.client_id` / `client.client_secret` 是旧版字段名，仍然可用；**同名时 `auth.*` 优先**。
- `client.proxy` 同样可用，但优先级低于 `auth.proxy`。

两个环境变量（`pkg/config/load.go` 里显式 `BindEnv`）：

```bash
export IPAAS_AGENT_AUTH_OPEN_API_HOST=https://api.dingtalk.com
export IPAAS_AGENT_AUTH_PROXY=http://127.0.0.1:7890
```

**只支持这两个环境变量**，其它配置项不能通过环境变量注入；环境变量的值优先于文件。

`clientID` 或 `clientSecret` 为空时进程不会立刻报错退出，而是一直建不起连接并在后台重试，
`log1.txt` 里会持续出现连接失败。排查连接问题时先看这两项是否为空。

### 4.4 `plugins.*` 数据源

四个数据源的配置格式一致，都是**列表**，一个 `config_key` 对应一套库。字段定义在
`pkg/plugins/sql_executor.go` 的 `Body` 结构体：

| 字段 | 适用 | 说明 |
| --- | --- | --- |
| `host` | 全部 | 主机名或 IP |
| `port` | 全部 | **必填，没有默认端口**。整数或字符串都行（`FlexInt` 兼容两种写法） |
| `user` | 全部 | 数据库账号。注意是 `user`，不是 `username` |
| `password` | 全部 | 密码 |
| `database` | MySQL / PG / SQL Server | 库名 |
| `service_name` | Oracle | 服务名，与 `sid` 二选一 |
| `sid` | Oracle | SID，与 `service_name` 二选一 |
| `config_key` | 全部 | 这套配置的引用名，平台侧就靠它指定用哪个库 |
| `address` | — | 会被解析，但**不参与任何连接串**，写不写都不生效 |
| `connection_str` | — | 同上，当前代码未使用 |

各数据库实际生成的连接串（摘自代码）：

```text
MySQL       user:password@tcp(host:port)/database
PostgreSQL  host=... port=... user=... password=... dbname=... sslmode=disable
SQL Server  server=...;port=...;user id=...;password=...;database=...;<less_common_parameters>
Oracle      go_ora.BuildUrl(host, port, service_name, user, password, {SID: sid})
```

由此可得几个实际约束：

- **PostgreSQL 强制 `sslmode=disable`**，不能通过配置开启 TLS。
- SQL Server 的 TLS 相关参数只能通过 `auth.mssql.less_common_parameters` 追加，
  这是兼容只支持 TLS 1.0/1.1 的旧 SQL Server 的唯一入口。
- Oracle 同时填 `service_name` 和 `sid` 时行为由驱动决定，请只填一个。
- 数据库连接是**每次动作临时建立、用完立刻关闭**（各插件 `DoSQLExecute` 里 `defer db.Close()`），
  没有连接池。所以网络抖动表现为单次失败，不会因为连接池失效而持续报错。

### 4.5 密码里的特殊字符

密码含 `#`、`:`、`{`、`&`、`*` 或前导空格时用引号包起来，否则会被 YAML 当成注释或锚点，
表现为"密码明明是对的但一直连不上"：

```yaml
password: "p@ss#word:123"
```

### 4.6 `allow_remote`：谁来决定连哪个库

判断逻辑在四个插件的 `HandleMessage` 里，完全一致：

```go
if remoteConf.ConfigKey == "" && p.AllowRemote {
    // 直接用平台报文里带来的 host/port/user/password/database
} else {
    // 用 config_key 去本地配置里找，找不到就失败
}
```

展开成三条结论：

- 平台**带了** `config_key` → 一定走本地配置，`allow_remote` 此时不起作用。
- 平台**没带** `config_key` 且 `allow_remote: true` → 用平台传来的连接参数。
- 平台**没带** `config_key` 且 `allow_remote: false` → 报 `未找到配置或不允许远程配置`。

默认为 `false`（`auth.mysql.allow_remote`、`auth.mssql.allow_remote` 在代码里显式
`SetDefault`，另外两个读不到值时同样按 `false` 处理）。它本质是一个安全开关：打开意味着
允许云端指定一个任意的内网地址去连，请按需权衡。

### 4.7 `timeout`：超时兜底

单位统一为**毫秒**。这三个键对应上游 issue #1（原本 HTTP 超时写死 5 秒、数据库查询永不超时），
实现见 `pkg/config/v1/timeout.go`。

| 键 | 默认 | 作用 |
| --- | --- | --- |
| `timeout.http_default_ms` | `30000` | 平台没下发 `timeout` 时，一次 HTTP 动作允许的时长 |
| `timeout.http_max_ms` | `0`（不限制） | 对平台下发的 `timeout` 做上限截断 |
| `timeout.db_ms` | `60000` | 一次数据库动作从建连到取完结果的总时长 |

生效顺序：平台下发的值优先 → 平台没给则用 `http_default_ms` → 设置了 `http_max_ms` 再截断。
填 `0` 或负数会回落到内置默认值，不会出现"永不超时"。

**协议版本会影响报文里 `timeout` 字段的单位**，这是个容易踩的坑：

- 协议 `1.0` 的 `timeout` 是**秒**（`pkg/plugins/v1/proxy_http.go` 会 `*1000` 换算）。
- 协议 `2.0` 的 `timeout` 是**毫秒**。
- 本文档的本地配置**始终是毫秒**。

### 4.8 协议 1.0 的 MySQL 代理：走顶层 `mysql:` 键

历史遗留：协议 1.0 的 MySQL 代理读的是**顶层 `mysql:`**，和 2.0 的 `plugins.mysql` 是两套
互不相通的配置（`pkg/plugins/v1/proxy_mysql.go` 的 `findConfigByKey` 里
`viper.UnmarshalKey("mysql", ...)`）。字段名也不同：

```yaml
# 只有需要兼容协议 1.0 的 MySQL 连接器时才需要这一段
mysql:
  - addr: 127.0.0.1:3306      # 注意：addr 是 host:port 合写
    username: root            # 注意：这里是 username，不是 user
    password: root
    database: example
    config_key: legacy-v1
```

如果连接平台上用的是旧版 MySQL 连接器，而配置只写了 `plugins.mysql`，日志里会出现
`mysql config xxx not found`，平台侧拿到的是一个空的 `null` 而不是明确报错——这是最容易
被误判成"网络问题"的一种失败。

### 4.9 配置热更新

网关通过 `viper.WatchConfig` 监听配置变化，保存即生效，不用重启。但生效范围有限：

- **立即生效**：`plugins.*` 数据源、`auth.*.allow_remote`、`timeout.*`（超时是每次调用现读的）。
- **需要重启**：`auth.clientID`、`auth.clientSecret`、`auth.openAPIHost`、`auth.proxy`。
  这几项只在建立连接的那一刻读取，改了要重启进程才会用新值重连。

---

## 5. 启动，以及怎么确认真的连上了

```bash
cd /opt/ipaas-agent
./ipaas-agent
```

终端里会看到 pterm 渲染的"IPaaS 本地网关"标题栏，然后：

```text
✔ 初始化成功
```

看到这一行才叫起来了。如果是这一行，说明进程活着但还没连上钉钉：

```text
✘ 初始化完成，与服务器的连接将在后台重试
```

这种状态下网关不会退出，会按 **2 秒起、每次翻倍、上限 1 分钟**的退避持续重连，网络恢复后
自动接上，不需要人工重启（对应上游 issue #2；修复前的行为是直接退出进程）。

`log1.txt` 里对应的成功标志是这两行：

```text
info    启动程序        buildTime=... gitCommit=... version=v0.9.11
info    成功连接到服务器
```

`version=DEV` 说明这个二进制构建时没传 `-ldflags`，属正常。

连接建立后，网关会在**空闲 30 秒**时发送一次 WebSocket 心跳（SDK 默认 120 秒）。这是为了让
NAT/防火墙静默掐断连接时能被发现，而不是挂在一个死连接上一直不自知。

---

## 6. 在连接平台侧怎么用

控制台的操作路径以钉钉控制台为准，这里只讲网关这一侧的对应关系，方便你从一次失败的调用反查配置。

### 6.1 协议版本怎么区分

报文里的 `specVersion` 决定走哪套解析（`pkg/plugins/v1/wrapper.go`）：

- `2.0`：`specVersion` 和 `pluginName` 都在 `data` 里，由 `pluginName` 直接指定插件。
- `1.0`：`specVersion` 在 `headers` 里，靠 `headers.type`（`http` / `mysql`）路由。
- 识别不出版本的一律按 HTTP 处理。

### 6.2 内置的插件名

| 名字 | 作用 |
| --- | --- |
| `http_plugin` | HTTP 反向代理，转发到内网接口 |
| `mysql_plugin` | 在 MySQL 执行 SQL |
| `pgsql_plugin` | 在 PostgreSQL 执行 SQL |
| `mssql_plugin` | 在 SQL Server 执行 SQL |
| `oracledb_plugin` | 在 Oracle 执行 SQL |
| `proxy_mysql_plugin` | 协议 1.0 的旧版 MySQL 代理，读顶层 `mysql:` |
| `version_plugin` | 回报协议版本，返回 `{"name":"version_plugin","protocol_version":"2.0"}` |

协议 1.0 的 HTTP 报文还有一个二次路由：带上 header `x-ipaas-plugin-name: mysql_plugin`
（或其它插件名），可以让 1.0 的 HTTP 通道去调用数据库插件，结果会被包装成 HTTP 响应的形状
（`pkg/plugins/http_plugin.go` 的 `handleV1`）。这是内部约定，普通 HTTP 代理不要设置它。

### 6.3 数据库动作的报文与返回

平台请求里带的字段就是 `Body` 结构体的字段，最少是：

```json
{ "config_key": "default", "sql": "select id, name from demo limit 10" }
```

`config_key` 留空且 `allow_remote: true` 时，可以把 `host`/`port`/`user`/`password`/`database` 一起带进来。

返回结构固定为：

```json
{
  "response": {
    "result": [ { "id": 1, "name": "demo" } ],
    "columns": ["id", "name"],
    "message": "success"
  }
}
```

`message` 成功时恒为 `success`，失败时是具体的错误文本（例如 `Unknown database 'xxx'`），
平台侧看到的报错就来自这里。

关于 SQL 有三点实际约束，都源于代码走的是 `db.QueryContext` 且没有打开驱动的多语句开关：

- 面向 **`SELECT`** 设计。`INSERT`/`UPDATE`/`DELETE` 会被执行，但响应里只有空结果集，
  **取不到受影响行数**。
- **一次只能一条语句**，`select 1; select 2` 这种分号多语句会被驱动拒绝。
- MySQL 插件结果里的二进制字段会统一转成字符串，`BLOB` 内容需要自行再解码。

### 6.4 怎么确认"平台到网关"这条链路通了

1. 网关日志有 `成功连接到服务器`，说明回调 topic `/v1.0/ipaas/proxy/callback` 注册成功。
2. 在连接平台触发一次调用，网关 `log1.txt` 里应出现 `收到服务器消息:` 一行。
3. 有第 2 步的日志但结果不对 → 问题在本地执行（配置、可达性、SQL），看紧随其后的错误行。
4. 完全没有第 2 步的日志 → 请求根本没到网关，问题在应用凭证、连接平台侧配置或该 topic 的订阅。

---

## 7. 故障排查

先看 `log1.txt`（业务日志），`log2.txt` 是 SDK 的连接层日志。两个文件都在**工作目录**下，
**不轮转**，长期运行需自行清理（8.3）。日志等级固定 `Info`，代码里没有开关。

| 现象 / 日志关键字 | 原因 | 处理 |
| --- | --- | --- |
| `读取配置文件出错: Config File "config" Not Found` | 工作目录不对（3.3），或文件没命名成 `config.yaml` | 启动前 `cd` 到配置所在目录；托管启动时设置工作目录 |
| `连接到服务器失败`，不断退避重试 | 凭证为空或错误、出网被拦、该走代理没配 | 核对 `auth.*`；`curl -v https://api.dingtalk.com`；需要时配 `auth.proxy` |
| 持续的 401 / 鉴权类错误 | `clientID`/`clientSecret` 与应用不匹配，或应用未开通 Stream 能力 | 到开发者后台核对凭证与能力开通情况 |
| `未找到配置或不允许远程配置: xxx` | 平台没带 `config_key`，或带了但本地没有同名条目 | 让平台带上正确的 `config_key`，或按 4.6 打开 `allow_remote` |
| `mysql config xxx not found`，平台拿到 `null` | 用的是协议 1.0 的 MySQL 连接器，但没配顶层 `mysql:` | 按 4.8 补顶层 `mysql:` 配置 |
| HTTP 动作约 5 秒就失败 | 老版本写死的超时（v0.9.11 已修） | 升级到 v0.9.11，或按 4.7 配 `http_default_ms` |
| `context deadline exceeded` | 触发了 `timeout.*` 兜底 | 判断是后端确实慢还是值偏小，按 4.7 调整 |
| `connection refused` / 连接超时 | 数据库地址或端口不可达 | 在**网关所在机器**上 `telnet host port` 验证，不要用办公网笔记本验证 |
| `Access denied for user` | 账号、密码或库名不对 | 注意密码里的 YAML 特殊字符（4.5）；先在网关本机用同名账号连一次 |
| 连接过一会儿断，之后又自愈 | 此前是 NAT 静默掐断（v0.9.11 已加 30 秒心跳） | 升级版本；若仍频繁断开，检查中间设备的会话超时 |
| `解析 ... 配置出错` | YAML 结构不符合 `plugins.{库名}` 的列表格式 | 检查缩进；列表项必须以 `-` 开头 |
| 平台侧长时间无响应 | 数据库查询或 HTTP 卡住 | v0.9.11 起有 `db_ms`/`http_*` 兜底；升级后配一个合理上限 |

排查连接层细节时，直接在 `log2.txt` 里找 SDK 的心跳与重连记录，它比 `log1.txt` 更贴近底层。

---

## 8. 长期运行

> 本节的服务化写法是通用模板，本仓库未做真机验证。可以依赖的事实只有两条：网关必须**先设好
> 工作目录**（3.3）；可靠的观测面是 `log1.txt` / `log2.txt`，终端输出只是给人看一眼。

### 8.1 Linux：systemd

```ini
# /etc/systemd/system/ipaas-agent.service
[Unit]
Description=DingTalk iPaaS local gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=nobody
WorkingDirectory=/opt/ipaas-agent
ExecStart=/opt/ipaas-agent/ipaas-agent
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable --now ipaas-agent
systemctl status ipaas-agent
```

`WorkingDirectory` 是关键的一行，缺了它就等于踩进 3.3 那个坑。

### 8.2 Windows：任务计划程序

```powershell
$action  = New-ScheduledTaskAction -Execute 'D:\ipaas-agent\ipaas-agent.exe' `
                                  -WorkingDirectory 'D:\ipaas-agent'
$trigger = New-ScheduledTaskTrigger -AtStartup
$setting = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries `
          -DontStopIfGoingOnBatteries -RestartCount 999 `
          -RestartInterval (New-TimeSpan -Minutes 1)
Register-ScheduledTask -TaskName ipaas-agent -Action $action -Trigger $trigger `
          -Settings $setting -User 'SYSTEM' -LogonType ServiceAccount -RunLevel Highest
Start-ScheduledTask -TaskName ipaas-agent
```

`-WorkingDirectory` 同样是必填项。图形界面里对应的字段是任务属性的"起始于(可选)"。

### 8.3 日志、升级与回滚

- 日志不轮转，建议挂 logrotate 或计划任务按大小/天数切分。网关以追加方式打开文件，在 Linux
  上轮转时直接删除正在写的文件不会让进程出错，但新内容会写进已删除的 inode，所以轮转后需要
  重启一次才会重新打开。
- 网关没有连接池和其它常驻状态，随时可以直接结束进程再启动，不影响平台侧已有配置。
- 升级：备份二进制和 `config.yaml` → 替换二进制 → 启动 → 确认 `log1.txt` 出现
  `成功连接到服务器`。回滚就是把二进制换回去。
- 跨版本注意：`timeout.*` 与 `auth.proxy` 是 v0.9.11 新增的，回滚到旧版本时它们会被忽略
  （不报错），但超时行为会退回"写死 5 秒 / 数据库不超时"。

---

## 9. 开发与测试

```bash
go build ./...
go vet ./...
go test ./...
```

`go test ./...` 默认只跑不依赖外部服务的用例，四个数据库插件的用例会**跳过**而不是失败，
这样在没有数据库实例的机器上（包括 CI）结果是干净的。准备好真实实例后显式开启：

```bash
export IPAAS_AGENT_DB_TEST=1
go test ./...
```

开启后如果库、账号、表结构缺任何一项，会以**失败**明确报出来，不会被静默跳过。

---

## 10. 与 README 的差异，以及当前已知限制

以下几处 README 的说法与代码不一致，以代码和本文档为准：

1. 字段名：README 的 MySQL 一节写的是 `username`，实际字段是 **`user`**（只有协议 1.0 的顶层
   `mysql:` 才用 `username`）。
2. `addr` / `address`：README 说"可选，未提供则用 host 和 port 拼接"，实际上这两个字段
   **被解析但不参与任何连接串**，填了也不生效。
3. 认证字段：README 只介绍 `client.client_id` / `client.client_secret`，实际推荐使用
   `auth.clientID` / `auth.clientSecret`，旧字段仍兼容但优先级较低。
4. README 没有提到工作目录这一层，而它恰好是最常见的启动失败原因（3.3）。

当前确实存在、但按最小改动原则未改动的限制：

- 数据库插件每次动作新建连接、用完关闭，没有连接池；高频调用时建连开销会体现在耗时日志里。
- PostgreSQL 固定 `sslmode=disable`，无法开启 TLS。
- 日志不轮转、等级固定 `Info`、没有输出到 stdout 的开关。
- 配置只认工作目录下的 `config.yaml`，不支持命令行指定路径。
- 数据库动作面向查询设计，写操作拿不到受影响行数（6.3）。

需要帮助时可以在 [issue 区](https://github.com/yhan1203/ipaas-agent/issues) 提问，并附上
`log1.txt` 中相关时间段的片段（记得先把密码、Secret 打码）。
