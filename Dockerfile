FROM golang:1.10.1

RUN apt update
RUN apt install -y libaio1

ENV INSTANT_CLIENT /opt/oracle/instantclient_12_2

ADD instantclient_12_2 $INSTANT_CLIENT

RUN echo $INSTANT_CLIENT > /etc/ld.so.conf.d/oracle-instantclient.conf

ENV LD_LIBRARY_PATH $INSTANT_CLIENT:$LD_LIBRARY_PATH

ADD src/golang.org /go/src/golang.org 
ADD src/gopkg.in /go/src/gopkg.in

ENV GOBIN /go/bin

RUN go get github.com/prometheus/client_golang/prometheus
RUN go get github.com/prometheus/client_golang/prometheus/promhttp
RUN go get github.com/prometheus/common/log
RUN go get github.com/prometheus/common/version
RUN go get gopkg.in/alecthomas/kingpin.v2
RUN go get gopkg.in/goracle.v2

WORKDIR /opt
ENTRYPOINT [ "./oracle_exporter" ]

ADD oracle_exporter.go /opt/oracle_exporter.go
RUN go get github.com/shinhwagk/oracle_exporter/collector
RUN go build -o oracle_exporter