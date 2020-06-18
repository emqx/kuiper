Kuiper 实现了几个插件。

## 源（Sources）

| 名称                  | 描述                                                  |
| --------------------- | ------------------------------------------------------------ |
| [zmq](sources/zmq.md)| 该插件监听Zero Mq消息并发送到Kuiper流中 |
| [random](sources/random.md) | 该插件按照指定模式生成消息   |



## 动作（Sinks/Actions）



| 名称                  | 描述                                                  |
| --------------------- | ------------------------------------------------------------ |
| [file](sinks/file.md) | 该插件将分析结果保存到某个指定到文件系统中 |
| [zmq](sinks/zmq.md)   | 该插件将分析结果发送到Zero Mq的主题中    |
| [influxdb](sinks/influxdb.md)   | 该插件将分析结果发送到InfluxDB中    |




## 函数（Functions）

...

