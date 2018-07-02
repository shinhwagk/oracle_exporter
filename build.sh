#!/bin/bash
unzip -o instantclient-basic-linux.x64-12.2.0.1.0.zip
docker build -t wex/oracle_exporter .


tag=0.9.25 && ./build.sh ${tag}
# docker rm -f oracle_func3
# docker run -d \
# --name oracle_func3 \
# -p 9174:9100 \
# -v /etc/localtime:/etc/localtime \
# -e DATA_SOURCE_NAME=system/oracle1171@10.65.193.38:1521/func3 \
# wex/oracle_exporter:${tag}