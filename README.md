### Usage
```sh
export DATA_SOURCE_NAME=${ip}:${port}/${service_name}
./oracle_exporter-2.4.6.linux-amd64
```
### metrics 11g,12c
- ash-10g
- ash-11g
- logHistory-10g
- logHistory-11g
- session-10g
- session-11g
- sessionEvent-10g
- sessionEvent-11g
- sessionStats-10g
- sessionStats-11g
- sessionTimeModel-10g
- sessionTimeModel-11g
- sql-10g
- sql-11g
- systemEvent-10g
- systemEvent-11g
- systemStats-10g
- systemStats-11g
- systemTimeModel-10g
- systemTimeModel-11g
- tablespace-10g
- tablespace-11g
- segment-10g
- segment-11g

```
http://127.0.0.1:9100/metrics?collect[]=sql-11g

export DATA_SOURCE_NAME=system/oracle@10.65.193.26/orayali3

go run main.go --file.metrics=http://gitlab.wexfin.com/oradba/prom-oracle/raw/master/metric.yml
```

```sh
go run main.go --database.datasource system/oracle1171@10.65.193.14:1521/func1 --file.metrics http://gitlab.wexfin.com/oradba/prom-oracle/raw/master/metrics-11g.yaml
```