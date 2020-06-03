# `Dockerfile` links

- [emqx/kuiper](https://github.com/emqx/kuiper/blob/master/docker/Dockerfile)

# Quick reference

- **Where to get help**:

  **<u>Web</u>**

  - https://github.com/emqx/kuiper

  **<u>Documents</u>**

  - [Getting started](https://github.com/emqx/kuiper/blob/master/docs/en_US/getting_started.md) 
  - [Reference guide](https://github.com/emqx/kuiper/blob/master/docs/en_US/reference.md)

- **Where to file issues:**

  https://github.com/emqx/kuiper/issues

- **Supported architectures**

  `amd64`, `arm64v8`,  `arm32v7`, `i386`, `ppc64le`

- **Supported Docker versions**:

  [The latest release](https://github.com/docker/docker-ce/releases/latest)

# Image Variants

The `emqx/kuiper` images come in many flavors, each designed for a specific use case.

## `emqx/kuiper:<tag>`

This is the defacto image, which is based on Debian and it also includes a Golang build environment. If you are unsure about what your needs  are, you probably want to use this one. It is designed to be used both as a throw away container (mount your source code, compile plugins for Kuiper,  and start the  container to run your app), as well as the base to build other images.

Notice: This image is the equivalent to development image of `x.x.x-dev` in 0.3.x versions.

## `emqx/kuiper:<tag>-slim`

This image is also based on Debian, and only contains the minimal packages needed to run `kuiper`. The difference between with previous image (`emqx/kuiper:<tag>`) is that this image does not include Golang development environment. The typical usage of this image would be deploy the plugins compiled in previous Docker image instances.

## `emqx/kuiper:<tag>-alpine`

This image is based on the popular [Alpine Linux project](http://alpinelinux.org), available in [the `alpine` official image](https://hub.docker.com/_/alpine). Alpine Linux is much smaller than most distribution base images (~5MB), and thus leads to much slimmer images in general. 

This variant is highly recommended when final image size being as  small as possible is desired. The main caveat to note is that it does  use [musl libc](http://www.musl-libc.org) instead of [glibc and friends](http://www.etalabs.net/compare_libcs.html), so certain software might run into issues depending on the depth of  their libc requirements. However, most software doesn't have an issue  with this, so this variant is usually a very safe choice. See [this Hacker News comment thread](https://news.ycombinator.com/item?id=10782897) for more discussion of the issues that might arise and some pro/con comparisons of using Alpine-based images.

To minimize image size, it's uncommon for additional related tools (such as `git` or `bash`) to be included in Alpine-based images. Using this image as a base, add the things you need in your own Dockerfile (see the [`alpine` image description](https://hub.docker.com/_/alpine/) for examples of how to install packages if you are unfamiliar).

# What is Kuiper

EMQ X Kuiper is an edge lightweight IoT data analytics / streaming software implemented by Golang, and it can be run at all kinds of resource constrained edge devices. One goal of Kuiper is to migrate the cloud streaming software frameworks (such as [Apache Spark](https://spark.apache.org)，[Apache Storm](https://storm.apache.org) and [Apache Flink](https://flink.apache.org)) to edge side.  Kuiper references these cloud streaming frameworks, and also considered special requirement of edge analytics, and introduced **rule engine**, which is based on ``Source``, ``SQL (business logic)`` and ``Sink``, rule engine is used for developing streaming applications at edge side.

![Kuiper architect](https://github.com/emqx/kuiper/raw/master/docs/resources/arch.png)

**User scenarios**

It can be run at various IoT edge use scenarios, such as real-time processing of production line data in the IIoT; Gateway of Connected Vehicle analyze the data from data-bus in real time; Real-time analysis of urban facility data in smart city scenarios. Kuiper processing at the edge can reduce system response latency, save network bandwidth and storage costs, and improve system security.

**Features**

- Lightweight

  - Core server package is only about 3MB, initial memory usage is about 10MB
- Cross-platform
  - CPU Arch：X86 AMD * 32, X86 AMD * 64; ARM * 32, ARM * 64; PPC
  - The popular Linux distributions, MacOS and Docker
  - Industrial PC, Raspberry Pi, industrial gateway, home gateway, MEC edge cloud server
- Data analysis support
  - Support data extract, transform and filter through SQL 
  - Data order, group, aggregation and join
  - 60+ functions, includes mathematical, string, aggregate and hash etc
  - 4 time windows
- Highly extensibile

  Plugin system is provided,  and it supports to extend at Source, SQL functions and Sink.
  - Source: embedded support for MQTT, and provide extension points for sources
  - Sink: embedded support for MQTT and HTTP, and provide extension points for sinks
  - UDF functions: embedded support for 60+ functions, and provide extension points for SQL functions
- Management
  - Stream and rule management through CLI
  - Stream and rule management through REST API (In planning)
  - Easily be integrate with [KubeEdge](https://github.com/kubeedge/kubeedge) and [K3s](https://github.com/rancher/k3s), which bases Kubernetes
- Integration with EMQ X Edge
  Seamless integration with EMQ X Edge, and provided an end to end solution from messaging to analytics. 


# How to use this image

### Run kuiper

Execute some command under this docker image

```
docker run -d -v `pwd`:$somewhere emqx/kuiper:$tag $somecommand
```

For example

```
docker run -d --name kuiper -e MQTT_BROKER_ADDRESS=$MQTT_BROKER_ADDRESS emqx/kuiper:$tag
```

#### 5 minutes quick start

1. Set Kuiper source to an MQTT server. This sample uses server locating at ``tcp://broker.emqx.io:1883``. ``broker.emqx.io`` is a public MQTT test server hosted by [EMQ](https://www.emqx.io).

   ```shell
   docker run -d --name kuiper -e MQTT_BROKER_ADDRESS=tcp://broker.emqx.io:1883 emqx/kuiper:$tag
   ```

2. Create a stream - the stream is your stream data schema, similar to table definition in database. Let's say the temperature & humidity data are sent to ``broker.emqx.io``, and those data will be processed in your **LOCAL RUN** Kuiper docker instance.  Below steps will create a stream named ``demo``, and data are sent to ``devices/device_001/messages`` topic, while ``device_001`` could be other devices, such as ``device_002``, all of those data will be subscribed and handled by ``demo`` stream.

   ```shell
   -- In host
   # docker exec -it kuiper /bin/sh
   
   -- In docker instance
   # bin/cli create stream demo '(temperature float, humidity bigint) WITH (FORMAT="JSON", DATASOURCE="devices/+/messages")'
   Connecting to 127.0.0.1:20498...
   Stream demo is created.
   
   # bin/cli query
   Connecting to 127.0.0.1:20498...
   kuiper > select * from demo where temperature > 30;
   Query was submit successfully.
   
   ```

3. Publish sensor data to topic ``devices/device_001/messages`` of server ``tcp://broker.emqx.io:1883`` with any [MQTT client tools](https://medium.com/@emqtt/mqtt-client-tools-215ff7a17ad). Below sample uses ``mosquitto_pub``. 

   ```shell
   # mosquitto_pub -h broker.emqx.io -m '{"temperature": 40, "humidity" : 20}' -t devices/device_001/messages
   ```

4. If everything goes well,  you can see the message is print on docker ``bin/cli query`` window. Please try to publish another message with ``temperature`` less than 30, and it will be filtered by WHERE condition of the SQL. 

   ```
   kuiper > select * from demo WHERE temperature > 30;
   [{"temperature": 40, "humidity" : 20}]
   ```

   If having any problems, please take a look at ``log/stream.log``.

5. To stop the test, just press ``ctrl + c `` in ``bin/cli query`` command console.

6. Next for exploring more powerful features of EMQ X  Kuiper? Refer to below for how to apply EMQ X Kuiper in edge and integrate with AWS / Azure IoT cloud.

   - [Lightweight edge computing EMQ X Kuiper and Azure IoT Hub integration solution](https://www.emqx.io/blog/85)   [简体中文](https://www.emqx.io/cn/blog/87)
   - [Lightweight edge computing EMQ X Kuiper and AWS IoT Hub integration solution](https://www.emqx.io/blog/88)     [简体中文](https://www.emqx.io/cn/blog/94)

### Configuration

Kuiper supports using environment variables to modify configuration files in containers

When modifying configuration files through environment variables, the environment variables need to be set according to the prescribed format, for example:

```
KUIBER__BASIC__DEBUG => basic.debug in etc/kuiper.yaml

MQTT_SOURCES__DEMO_CONF__QOS => demo_conf.qos in etc/mqtt_source.yaml
```

The environment variables are separated by two "_", the content of the first part after the separation matches the file name of the configuration file, and the remaining content matches the different levels of the configuration item.

At the same time, Kuiper provides some fixed format environment variables to modify specific configurations, but this part of the content will be abolished in subsequent versions, please use with caution.

Use the environment variable to configure `etc/client.yaml`  on the Kuiper container.

| Options                         | Default               | Mapped                      |
| ------------------------------- | --------------------- | --------------------------- |
| CLIENT_HOST                     | 127.0.0.1             | client.basic.debug          |
| CLIENT_PORT                     | 20498                 | client.basic.port           |

Use the environment variable to configure `etc/kuiper.yaml`  on the Kuiper container.

| Options                         | Default               | Mapped                      |
| ------------------------------- | --------------------- | --------------------------- |
| KUIPER_DEBUG                    | false                 | kuiper.basic.debug          |
| KUIPER_CONSOLE_LOG              | false                 | kuiper.basic.consoleLog     |
| KUIPER_FILE_LOG                 | true                  | kuiper.basic.fileLog        |
| KUIPER_PORT                     | 20498                 | kuiper.basic.port           |
| KUIPER_REST_PORT                | 9081                  | kuiper.basic.restPort       |
| KUIPER_PROMETHEUS               | false                 | kuiper.basic.prometheus     |
| KUIPER_PROMETHEUS_PORT          | 20499                 | kuiper.basic.prometheusPort |

Use the environment variable to configure `etc/mqtt_sources.yaml`  on the Kuiper container.

| Options                         | Default               | Mapped                      |
| ------------------------------- | --------------------- | --------------------------- |
| MQTT_BROKER_ADDRESS             | tcp://127.0.0.1:1883  | default.servers             |
| MQTT_BROKER_SHARED_SUBSCRIPTION | true                  | default.sharedSubscription  |
| MQTT_BROKER_QOS                 | 1                     | default.qos                 |
| MQTT_BROKER_USERNAME            |                       | default.username            |
| MQTT_BROKER_PASSWORD            |                       | default.password            |
| MQTT_BROKER_CER_PATH            |                       | default.certificationPath   |
| MQTT_BROKER_KEY_PATH            |                       | default.privateKeyPath      |

Use the environment variable to configure `etc/sources/edgex.yaml`  on the Kuiper container.

| Options                    | Default                  | Mapped                    |
| ---------------------------| -------------------------| ------------------------- |
| EDGEX_PROTOCOL             | tcp                      | default.protocol          |
| EDGEX_SERVER               | localhost                | default.server            |
| EDGEX_PORT                 | 5563                     | default.port              |
| EDGEX_TOPIC                | events                   | default.topic             |
| EDGEX_SERVICE_SERVER       | http://localhost:48080   | default.serviceServer     |

All of the environment variable should be set with corresponding values that configured in file ``cmd/core-data/res/configuration.toml`` of EdgeX core-data service, as listed in below.

```
[MessageQueue]
Protocol = 'tcp'
Host = '*'
Port = 5563
Type = 'zero'
Topic = 'events'
```

```
[Service]
...
Host = 'localhost'
Port = 48080
...
```

If you want to configure more options, you can mount the configuration file into Kuiper container, like this:
```
$ docker run --name kuiper -v /path/to/mqtt_sources.yaml:/kuiper/etc/mqtt_sources.yaml -v /path/to/edgex.yaml:/kuiper/etc/sources/edgex.yaml emqx/kuiper:$tag
```

# More

If you'd like to know more about the project, please refer to [Github project](https://github.com/emqx/kuiper/blob/master/docs/en_US/README.md).

