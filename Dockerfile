FROM golang:1.10.3

RUN apt update
RUN apt install -y libaio1

ENV INSTANT_CLIENT /opt/oracle/instantclient_12_2

ADD instantclient_12_2 $INSTANT_CLIENT

RUN echo $INSTANT_CLIENT > /etc/ld.so.conf.d/oracle-instantclient.conf

ENV LD_LIBRARY_PATH $INSTANT_CLIENT:$LD_LIBRARY_PATH

ENV GOBIN /go/bin

RUN git config --global http.proxy http://10.65.193.52:8118

RUN go get -v github.com/shinhwagk/oracle_exporter/collector
WORKDIR /go/src/github.com/shinhwagk/oracle_exporter

RUN go get -v

RUN go build -o oracle_exporter

ENTRYPOINT [ "./oracle_exporter" ]
