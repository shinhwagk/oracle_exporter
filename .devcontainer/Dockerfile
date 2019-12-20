FROM golang:1.13.5

RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go env -w GO111MODULE="on"

RUN go get -v golang.org/x/tools/gopls