# ipaas-gateway

`ipaas-gateway` 是一个基于钉钉开放平台stream能力构建的系统。它允许用户在本地部署客户端代理（agent），以提供反向代理服务，连接内部服务与钉钉开放平台。

## 配置文件说明

`ipaas-gateway` 使用 YAML 格式的配置文件。下面是配置文件各字段的详细说明：

### client 配置

- `client_id`: 你的钉钉开放平台应用的客户端ID（AppKey），可以在钉钉开发者后台获取。
- `client_secret`: 该应用的客户端密钥（AppSecret），同样在钉钉开发者后台获取。

这两个字段是必须的，因为它们用于身份验证和与钉钉开放平台的通信。

### mysql 配置

配置文件中 `mysql` 部分是用来定义数据库连接的，它可以包含多个数据库配置。每个数据库配置包括以下字段：

- `host`: 数据库服务器的主机名或IP地址。
- `addr`: 数据库服务器的地址和端口，通常格式为 `hostname:port`。
- `username`: 用于连接数据库的用户名称。
- `password`: 用于连接数据库的密码。
- `database`: 要连接的数据库名。
- `config`: 额外的数据库连接配置，如：
  - `max_open_conns`: 数据库连接池允许的最大打开连接数。
  - `max_idle_conns`: 数据库连接池允许的最大空闲连接数。
- `config_key`: 此数据库配置的引用键名，用于在代码中引用特定数据库配置。

## How to Use

在你的项目目录中添加一个名为 `config.yml` 的配置文件，根据上述字段填写对应的信息。例如：

```yaml
client:
  client_id: dingeypapfxxx
  client_secret: 3YZT7mQjlIxxx
mysql:
  - host: localhost
    addr: localhost:3306
    username: root
    password: root
    database: example
    config:
      max_open_conns: 10
      max_idle_conns: 5
    config_key: default
```

确保 `client` 部分包含有效的 `client_id` 和 `client_secret`，`mysql` 部分包含正确的数据库连接信息。

## 部署

1. 安装必要的依赖。
2. 在你的系统上部署 `ipaas-gateway` 代理。
3. 将 `config.yml` 配置文件放在代理可访问的位置。
4. 启动代理，并确保它可以连接到钉钉开放平台以及配置的内部服务。

## 日志

如果需要更改日志的配置，请在目录下新建 `/config/log.yaml` 文件，并填写对应字段。例如：

```yaml
log:
  level: debug
  path: tmp/log
  name: ipaas-agent.log
  maxsize: 100
  maxage: 7
  maxbackups: 10
  compress: true
```

以下是各个参数的说明：

- level: 日志级别。可选的值包括debug、info、warn、error等。debug级别会打印所有的日志，error级别只会打印错误日志。

- path: 日志文件的存储路径。例如，tmp/log表示日志文件将被存储在tmp/log目录下。

- name: 日志文件的名称。例如，ipaas-agent.log表示日志文件的名称将是ipaas-agent.log。

- maxsize: 每个日志文件的最大大小（以MB为单位）。例如，100表示每个日志文件的最大大小是100MB。

- maxage: 日志文件的最大存储时间（以天为单位）。例如，7表示每个日志文件会被保留7天。

- maxbackups: 保留的最大日志文件数量。例如，10表示最多会保留10个日志文件。

- compress: 是否压缩日志文件。true表示会压缩日志文件，false表示不会压缩日志文件。

请确保你的log配置包含了所有必要的参数，并且这些参数的值是正确的。

## 需要帮助？

如果你有任何疑问或需要帮助，可以查阅钉钉开放平台的文档，或者在项目的 issue 区提问。

