# Copyright 2021 EMQ Technologies Co., Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.15.1 AS builder

COPY . /go/kuiper

WORKDIR /go/kuiper

RUN apt update && apt install -y pkg-config libczmq-dev

RUN make build_with_edgex \
    && ln -s /go/kuiper/_build/kuiper-$(git describe --tags --always)-$(go env GOOS)-$(go env GOARCH) /usr/local/kuiper

RUN go build  -o kuiper_conf_util deploy/docker/conf_util.go \
    && if [ "$(uname -m)" != "aarch64" ]; then apt install -y upx && upx ./kuiper_conf_util; fi \
    && ln -s /go/kuiper/kuiper_conf_util /usr/bin/kuiper_conf_util

RUN ln -s /go/kuiper/deploy/docker/docker-entrypoint.sh /usr/bin/docker-entrypoint.sh

EXPOSE 9081 20498

ENV MAINTAINER="emqx.io"
ENV KUIPER_HOME /usr/local/kuiper
ENV KUIPER__BASIC__CONSOLELOG true

WORKDIR /usr/local/kuiper
VOLUME ["/usr/local/kuiper/etc", "/usr/local/kuiper/data", "/usr/local/kuiper/plugins", "/usr/local/kuiper/log"]
ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

CMD ["./bin/kuiperd"]
