FROM alpine
RUN apk update && apk add --no-cache libaio

FROM golang:1.10.3

RUN apt update
RUN apt install -y libaio1 unzip

RUN curl -OL https://github.com/shinhwagk/oracle_exporter/releases/download/v2.6.2/instantclient-basic-linux.x64-12.2.0.1.0.zip
RUN unzip instantclient-basic-linux.x64-12.2.0.1.0.zip
RUN mkdir -p /opt/oracle
RUN mv instantclient_12_2 /opt/oracle/instantclient_12_2
ENV INSTANT_CLIENT /opt/oracle/instantclient_12_2
RUN rm instantclient-basic-linux.x64-12.2.0.1.0.zip

RUN echo $INSTANT_CLIENT > /etc/ld.so.conf.d/oracle-instantclient.conf

ENV LD_LIBRARY_PATH $INSTANT_CLIENT:$LD_LIBRARY_PATH

ENV GOBIN /go/bin

RUN go get -v github.com/shinhwagk/oracle_exporter/collector
WORKDIR /go/src/github.com/shinhwagk/oracle_exporter

RUN go get -v

RUN go build -o oracle_exporter

ENTRYPOINT [ "./oracle_exporter" ]
