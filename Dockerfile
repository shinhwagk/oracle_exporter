FROM oraclelinux:7-slim
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
RUN GOOS=linux go build -ldflags "-X main.Version=0.3.0 -s -w" -o ./oracle_exporter
