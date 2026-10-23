# ipaas-gateway

`ipaas-gateway` 是一个基于钉钉开放平台stream能力构建的系统。它允许用户在本地部署客户端代理（agent），以提供反向代理服务，连接内部服务与钉钉开放平台。

完整的上手步骤、配置字段逐项说明、启动成功判据、故障排查表和长期运行建议见 [使用指南 docs/HOWTOUSE.md](docs/HOWTOUSE.md)。本页保留配置字段的速查说明。

## 配置文件说明

`ipaas-gateway` 使用 YAML 格式的配置文件来配置代理的行为。配置文件中包含以下部分：

### 一个完整的配置文件示例

```yaml
client:
  client_id: dingeypapxxxxxxxxxxx
  client_secret: xxxxxxxxxx7mQjlIF7q6YiFitxxxxxxxxxxxxxxxxxxxx
plugins:
  mssql:
    - host: localhost
      port: 1433
      address: localhost:1433
      user: sa
      password: sa123456A
      database: TestDB
      config_key: sqlServer
  mysql:
    - host: localhost
      port: 3306
      address: localhost:1433
      user: root
      password: root
      database: example
      config_key: default
  oracledb:
    - host: localhost
      port: 1521
      address: localhost:1521
      user: system
      password: example
      sid: FREE
      config_key: oracaldb
auth:
  mssql:
    allow_remote: true
  mysql:
    allow_remote: true
  pgsql:
    allow_remote: true
  # 只在需要走代理出网时填写
  proxy: ""
timeout:
  http_default_ms: 30000
  http_max_ms: 0
  db_ms: 60000
```

### client 配置

- `client_id`: 你的钉钉开放平台应用的客户端ID（AppKey），可以在钉钉开发者后台获取。
- `client_secret`: 该应用的客户端密钥（AppSecret），同样在钉钉开发者后台获取。

这两个字段是**必须的**，因为它们用于身份验证和与钉钉开放平台的通信。

### mysql 配置

配置文件中 `plugin.mysql` 部分是用来定义数据库连接的，它可以包含多个数据库配置（列表格式）。每个数据库配置包括以下字段：

- `host`: 数据库服务器的主机名或IP地址。
- `port`: 数据库服务器的端口号。
- `addr`: 数据库服务器的地址和端口，通常格式为 `hostname:port`。可选字段，如果未提供，则使用 `host` 和 `port` 拼接。
- `username`: 用于连接数据库的用户名称。
- `password`: 用于连接数据库的密码。
- `database`: 要连接的数据库名。
- `config_key`: 此数据库配置的引用键名，用于在代码中引用特定数据库配置。

鉴权配置：

- `auth.allow_remote`: 是否允许远程配置。如果设置为 `true`，则允许连接平台传入临时配置；如果设置为 `false`，则只允许本地配置文件的设置。

### mssql 配置

配置文件中 `plugin.mssql` 部分是用来定义数据库连接的，它可以包含多个数据库配置（列表格式）。每个数据库配置包括以下字段：

- `host`: 数据库服务器的主机名或IP地址。
- `port`: 数据库服务器的端口号。
- `address`: 数据库服务器的地址和端口，通常格式为 `hostname:port`。可选字段，如果未提供，则使用 `host` 和 `port` 拼接。
- `user`: 用于连接数据库的用户名称。
- `password`: 用于连接数据库的密码。
- `database`: 要连接的数据库名。
- `config_key`: 此数据库配置的引用键名，用于在代码中引用特定数据库配置。

鉴权配置：

- `auth.allow_remote`: 是否允许远程配置。如果设置为 `true`，则允许连接平台传入临时配置；如果设置为 `false`，则只允许本地配置文件的设置。

### oracledb 配置

配置文件中 `plugin.oracledb` 部分是用来定义数据库连接的，它可以包含多个数据库配置（列表格式）。每个数据库配置包括以下字段：

- `host`: 数据库服务器的主机名或IP地址。
- `port`: 数据库服务器的端口号。
- `address`: 数据库服务器的地址和端口，通常格式为 `hostname:port`。可选字段，如果未提供，则使用 `host` 和 `port` 拼接。
- `user`: 用于连接数据库的用户名称。
- `password`: 用于连接数据库的密码。
- `service_name`: 数据库的服务名。与 `sid` 二选一。
- `sid`: 数据库的SID。与 `service_name` 二选一。
- `config_key`: 此数据库配置的引用键名，用于在代码中引用特定数据库配置。

鉴权配置：

- `auth.allow_remote`: 是否允许远程配置。如果设置为 `true`，则允许连接平台传入临时配置；如果设置为 `false`，则只允许本地配置文件的设置。

### timeout 配置

本地网关侧的超时兜底，单位统一为毫秒，所有字段都可以省略（括号内为默认值）。

- `timeout.http_default_ms`（默认 `30000`）：连接平台没有下发 `timeout` 时，一次 HTTP 动作允许的时长。这个值此前写死在代码里是 5 秒，后端稍慢就会失败。
- `timeout.http_max_ms`（默认 `0`，即不限制）：对平台下发的 `timeout` 做上限截断，避免一条动作长时间占住回调。
- `timeout.db_ms`（默认 `60000`）：一次数据库动作从建立连接到取完结果的总时长。

优先级：平台下发的 `timeout` 优先，平台没给值时用 `http_default_ms`，设置了 `http_max_ms` 时再按它截断。

> 单位说明：协议 1.0 报文里的 `timeout` 字段是**秒**，协议 2.0 是**毫秒**，两个协议共用同一个字段名，上表配置始终是毫秒。

### 代理与重连

- `auth.proxy`（或环境变量 `IPAAS_AGENT_AUTH_PROXY`）：连接钉钉开放平台使用的代理地址，例如 `http://127.0.0.1:7890`、`socks5://127.0.0.1:1080`。仅在"内网只能经代理出网"时填写，留空表示直连。
  启动或运行期间连不上钉钉时，网关不再退出进程，而是按 2 秒起、上限 1 分钟的指数退避持续重连，网络恢复后自动接上，无需人工重启。
- 连接空闲时按 30 秒发送一次 WebSocket 心跳，避免被 NAT 或防火墙静默掐断后一直挂在死连接上。

## 如何使用

在你的项目目录中添加一个名为 `config.yml` 的配置文件，根据上述字段填写对应的信息。例如：

```yaml
client:
  client_id: dingeypapxxxxxxxxxxx
  client_secret: xxxxxxxxxx7mQjlIF7q6YiFitxxxxxxxxxxxxxxxxxxxx
plugins:
  mssql:
    - host: localhost
      port: 1433
      address: localhost:1433
      user: sa
      password: sa123456A
      database: TestDB
      config_key: sqlServer
```

> 确保 `client` 部分包含有效的 `client_id` 和 `client_secret`，`mysql` 部分包含正确的数据库连接信息。

## 部署

1. 进入[钉钉开放平台](https://open-dev.dingtalk.com/#/)，创建一个新的应用，并获取 `client_id` 和 `client_secret`。
2. 进入[Release](https://github.com/open-dingtalk/ipaas-agent/releases)页面，下载最新版本的 `ipaas-agent` 本地网关。请选择适合你操作系统的版本。
3. 将 `config.yaml` 配置文件放在本地网关程序的根目录下。
4. 启动代理，并确保它可以连接到钉钉开放平台以及配置的内部服务。

## 日志

在同一目录下，会产生 `log1.txt` 和 `log2.txt` 两个日志文件。

## 开发

### 运行测试

`go test ./...` 默认只跑不依赖外部服务的用例。四个数据库插件的用例需要按上面的配置准备好真实实例以及固定的库名、账号和表结构，默认跳过；准备好之后显式开启：

```powershell
$env:IPAAS_AGENT_DB_TEST = "1"
go test ./...
```

端口在监听不等于那就是用例需要的实例：库名、账号、表结构缺任何一项都会以失败报出来，而不是被跳过。

## 需要帮助？

如果你有任何疑问或需要帮助，可以查阅钉钉开放平台的文档，或者在项目的 issue 区提问。

