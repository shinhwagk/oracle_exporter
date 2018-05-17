#!/bin/bash
go get github.com/shinhwagk/oracle_exporter/collector
go build -o oracle_exporter
./oracle_exporter