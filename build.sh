#!/usr/bin/env bash
base=$(dirname $0)
main="oracle_exporter.go"
exe="oracle_exporter"

cd $bash
git pull

mkdir -p /go/src/golang.org/x && cd /go/src/golang.org/x/ && git clone --depth=1 https://github.com/golang/crypto
mkdir -p /go/src/golang.org/x && cd /go/src/golang.org/x/ && git clone --depth=1 https://github.com/golang/sys
go get -v 
go build -v -o $exe $main