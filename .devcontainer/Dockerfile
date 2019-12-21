FROM golang:1.13.5

RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go env -w GO111MODULE="on"

RUN go get -v golang.org/x/tools/gopls
RUN go get -v github.com/mdempsky/gocode
RUN go get -v github.com/uudashr/gopkgs/cmd/gopkgs
RUN go get -v github.com/ramya-rao-a/go-outline
RUN go get -v github.com/stamblerre/gocode
RUN go get -v github.com/rogpeppe/godef
RUN go get -v golang.org/x/lint/golint