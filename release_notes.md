## 更新日志

本地网关目前处于公测阶段，如有问题可以通过[工单](https://open.dingtalk.com/document/connector/what-is-the-connection-platform#title-4sd-7no-7f0)联系我们

“钉钉连接平台自动化官方互助交流群”群的钉钉群号： 109135000489

### 2026-10-23

- 启动时连不上钉钉不再退出进程，改为按指数退避在后台持续重连，网络恢复后自动接上（对应 issue #2）
- 支持 `auth.proxy` 配置，内网只能经代理出网时可以正常建连
- 连接空闲时按 30 秒发送 WebSocket 心跳，避免被 NAT 或防火墙静默掐断后一直挂在死连接上
- 新增 `timeout.http_default_ms`、`timeout.http_max_ms`、`timeout.db_ms` 超时配置（对应 issue #1）
    - HTTP 动作原本在没有下发 `timeout` 时写死 5 秒，现在按本地配置兜底，且数据库查询也有超时
- 坏报文不再导致进程 panic；`mysql` 旧版代理不再因为一个坏配置 `panic`，并修掉错误分支的连接泄漏
- 升级依赖，修掉一个被实际调用到的 WebSocket 弱随机问题（GHSA / GO-2026-6278）
- 修好发布流水线：原有的 artifact action 已退役会导致构建失败
- `go test ./...` 在没有数据库实例的机器上不再报失败，数据库插件的用例改为显式开启

### 2024-12-13

- MSSQL 插件支持 `auth.mssql.less_common_parameters` 配置, 用于配置不常用的参数。可用于兼容TLS1.0、TLS1.1等特殊情况


### 2024-12-10

- 增强数据库插件的日志输出

### 2024-12-02

- 增强 HTTP 插件的请求处理逻辑，支持更灵活的请求参数和响应结构；
    - 新增单元测试以验证 HTTP 请求和响应的正确性
- 修正了 `mysql` 插件读取字段类型问题

### 2024-12-01

- 支持 `mssql` 和 `oracledb`、`pgsql` 数据库插件
- 支持 `auth.allow_remote` 配置
- 优化了日志输出
