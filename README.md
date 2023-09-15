### Usage
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

### dashboard
![demo.jpg](./images/demo.jpg)