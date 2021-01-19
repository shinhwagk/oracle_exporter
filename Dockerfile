FROM oraclelinux:7-slim as builder
RUN yum install -y oracle-golang-release-el7
RUN yum install -y oracle-release-el7 && \
    yum install -y oracle-instantclient18.3-basic && \
    rm -rf /var/cache/yum

RUN echo /usr/lib/oracle/18.3/client64/lib > /etc/ld.so.conf.d/oracle-instantclient.conf && \
    ldconfig

ENV LD_LIBRARY_PATH=/usr/lib/oracle/18.3/client64/lib:$LD_LIBRARY_PATH
WORKDIR /build
ADD main.go .

RUN yum install -y golang

RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go env -w GO111MODULE="on"

ADD go.mod .
ADD go.sum .
RUN go get -v -u
RUN GOOS=linux go build main.go

FROM oraclelinux:7-slim
RUN yum install -y oracle-release-el7 && \
    yum install -y oracle-instantclient18.3-basic && \
    rm -rf /var/cache/yum
RUN du -sh /usr/lib/oracle
RUN echo /usr/lib/oracle/18.3/client64/lib > /etc/ld.so.conf.d/oracle-instantclient.conf && \
    ldconfig

ENV LD_LIBRARY_PATH=/usr/lib/oracle/18.3/client64/lib:$LD_LIBRARY_PATH
WORKDIR /app
COPY --from=builder /build/main .
ENTRYPOINT /app/main