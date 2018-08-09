#!/usr/bin/env bash
base=$(dirname $0)
main="oracle_exporter.go"
exe="oracle_exporter"

cd $bash
git pull

go get -v 
go build -v -o $exe $main