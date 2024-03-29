### dashboard
![demo.jpg](./images/demo.jpg)

### Usage
docker-compose.yml
```yaml
version: "3"

services:
  multidatabase-query:
    image: shinhwagk/multidatabase:v0.2.11
    restart: always
    deploy:
      replicas: 2
    environment:
      EXECUTE_TIMEOUT: 60000
      ORACLE_USERPASS: oracle_exporter:oracle_exporter
  oracle_exporter_11g:
    image: docker.io/shinhwagk/oracle_exporter:v1.3.9-mdb
    restart: always
    ports:
      - 9521:9521
    command:
      - "--file.metrics=https://raw.githubusercontent.com/shinhwagk/oracle_exporter/mdb/yaml/metrics-11g.yml"
      - "--mdb.addr=multidatabase-query:5000"
    depends_on:
      - multidatabase-query
```

prometheus.yml
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
alerting:
rule_files:

scrape_configs:
  - job_name: oracle-exporter-common
    scrape_interval: 20s
    metrics_path: /metrics
    scheme: http
    params:
      collect[]:
        - session-timemodel
        - session-statistic
        - session-event
        - ash
        - session
        - system-timemodel
        - system-event
        - system-statistic
    relabel_configs:
      - source_labels: [db_dsn]
        target_label: __param_dsn
        replacement: $1
    static_configs:
      - targets:
          - oracle_exporter_11g:9521
        labels:
          db_server: db-1
          db_dsn: "1.1.1.1:1521/servicename"
  - job_name: oracle-exporter-sql
    scrape_interval: 1m
    metrics_path: /metrics
    scheme: http
    params:
      collect[]:
        - sql
    relabel_configs:
      - source_labels: [db_dsn]
        target_label: __param_dsn
        replacement: $1
    static_configs:
      - targets:
          - oracle_exporter_11g:9521
        labels:
          db_server: db-1
          db_dsn: "1.1.1.1:1521/servicename"

```