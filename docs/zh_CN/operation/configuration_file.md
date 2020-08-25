# 基本配置
Kuiper 的配置文件位于 `$ kuiper / etc / kuiper.yaml` 中。 配置文件为 yaml 格式。

## 日志级别

```yaml
basic:
  # true|false, with debug level, it prints more debug info
  debug: false
  # true|false, if it's set to true, then the log will be print to console
  consoleLog: false
  # true|false, if it's set to true, then the log will be print to log file
  fileLog: true
```
## Cli 端口
```yaml
basic:
  # CLI port
  port: 20498
```
CLI 服务器监听端口

## REST 服务配置

```yaml
basic:
  # REST service port
  restPort: 9081
  restTls:
    certfile: /var/https-server.crt
    keyfile: /var/https-server.key
```

#### restPort
REST http 服务器监听端口

#### restTls
TLS 证书 cert 文件和 key 文件位置。如果 restTls 选项未配置，则 REST 服务器将启动为 http 服务器，否则启动为 https 服务器。

## Prometheus 配置

如果 `prometheus` 参数设置为 true，Kuiper 将把运行指标暴露到 prometheus。Prometheus 将运行在 `prometheusPort` 参数指定的端口上。

```yaml
basic:
  prometheus: true
  prometheusPort: 20499
```
在如上默认配置中，Kuiper 暴露于 Prometheusd 运行指标可通过 `http://localhost:20499/metrics` 访问。

## Sink configurations

```yaml
  #The cache persistence threshold size. If the message in sink cache is larger than 10, then it triggers persistence. If you find the remote system is slow to response, or sink throughput is small, then it's recommend to increase below 2 configurations.More memory is required with the increase of below 2 configurations.

  # If the message count reaches below value, then it triggers persistence.
  cacheThreshold: 10
  # The message persistence is triggered by a ticker, and cacheTriggerCount is for using configure the count to trigger the persistence procedure regardless if the message number reaches cacheThreshold or not. This is to prevent the data won't be saved as the cache never pass the threshold.
  cacheTriggerCount: 15

  # Control to disable cache or not. If it's set to true, then the cache will be disabled, otherwise, it will be enabled.
  disableCache: false
```

