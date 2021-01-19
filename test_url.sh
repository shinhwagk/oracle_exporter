http://127.0.0.1:9100/metrics?collect[]=sql-11g

export DATA_SOURCE_NAME=system/oracle@10.65.193.26/orayali3

go run main.go --file.metrics=http://gitlab.wexfin.com/oradba/prom-oracle/raw/master/metric.yml