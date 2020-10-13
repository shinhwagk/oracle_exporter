FROM golang:1.15

RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go env -w GO111MODULE="on"

ENV INSTANT_CLIENT /opt/oracle/instantclient_12_2
ENV LD_LIBRARY_PATH $INSTANT_CLIENT:$LD_LIBRARY_PATH

RUN sed -i 's/deb\.debian\.org/mirrors\.aliyun\.com/' /etc/apt/sources.list && \
    sed -i 's/security\.debian\.org/mirrors\.aliyun\.com/' /etc/apt/sources.list && \
    apt update && apt install -y libaio1 curl unzip && \
    curl -OL https://github.com/shinhwagk/oracle_exporter/releases/download/v2.6.2/instantclient-basic-linux.x64-12.2.0.1.0.zip
RUN unzip instantclient-basic-linux.x64-12.2.0.1.0.zip && \
    mkdir -p /opt/oracle && mv instantclient_12_2 /opt/oracle/instantclient_12_2 && \
    echo $INSTANT_CLIENT > /etc/ld.so.conf.d/oracle-instantclient.conf && \
    rm instantclient-basic-linux.x64-12.2.0.1.0.zip
# RUN go get -v golang.org/x/tools/gopls
# RUN go get -v github.com/mdempsky/gocode
# RUN go get -v github.com/uudashr/gopkgs/cmd/gopkgs
# RUN go get -v github.com/ramya-rao-a/go-outline
# RUN go get -v github.com/stamblerre/gocode
# RUN go get -v github.com/rogpeppe/godef
# RUN go get -v golang.org/x/lint/golint