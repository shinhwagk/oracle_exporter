FROM oraclelinux:7-slim
RUN yum install -y oracle-release-el7 && \
    yum install -y oracle-instantclient18.3-basic && \
    rm -rf /var/cache/yum

RUN echo /usr/lib/oracle/18.3/client64/lib > /etc/ld.so.conf.d/oracle-instantclient.conf && \
    ldconfig

ENV LD_LIBRARY_PATH=/usr/lib/oracle/18.3/client64/lib:$LD_LIBRARY_PATH
WORKDIR /app
ADD https://github.com/shinhwagk/oracle_exporter/releases/download/v2.6.2/oracle_exporter-v2.6.2.linux-amd64.tar.gz .
ENTRYPOINT [ "/app/oracle_exporter-v2.6.2.linux-amd64" ]
