FROM golang:1.10.1

RUN apt update
RUN apt install -y libaio1

ENV INSTANT_CLIENT /opt/oracle/instantclient_12_2

ADD instantclient_12_2 $INSTANT_CLIENT

RUN echo $INSTANT_CLIENT > /etc/ld.so.conf.d/oracle-instantclient.conf

ENV LD_LIBRARY_PATH $INSTANT_CLIENT:$LD_LIBRARY_PATH

ADD src/golang.org /go/src/golang.org 
ADD src/gopkg.in /go/src/gopkg.in

ADD oracle_exporter.go /opt/oracle_exporter.go

ENV GOBIN /go/bin

WORKDIR /opt

CMD go get && go build -o oracle_exporter && ./oracle_exporter