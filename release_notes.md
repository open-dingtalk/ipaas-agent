## 更新日志

本地网关目前处于公测阶段，如有问题可以通过[工单](https://open.dingtalk.com/document/connector/what-is-the-connection-platform#title-4sd-7no-7f0)联系我们

“钉钉连接平台自动化官方互助交流群”群的钉钉群号： 109135000489

### 2021-12-02

- 增强 HTTP 插件的请求处理逻辑，支持更灵活的请求参数和响应结构；
    - 新增单元测试以验证 HTTP 请求和响应的正确性
- 修正了 `mysql` 插件读取字段类型问题

### 2021-12-01

- 支持 `mssql` 和 `oracledb`、`pgsql` 数据库插件
- 支持 `auth.allow_remote` 配置
- 优化了日志输出