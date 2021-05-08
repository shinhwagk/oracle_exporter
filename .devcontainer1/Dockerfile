FROM go

RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go env -w GO111MODULE="on"

RUN go get -v -u golang.org/x/tools/gopls

# ENV INSTANT_CLIENT=/usr/lib/oracle/18.5/client64/lib
# ENV LD_LIBRARY_PATH $INSTANT_CLIENT:$LD_LIBRARY_PATH
# RUN echo ${INSTANT_CLIENT} > /etc/ld.so.conf.d/oracle-instantclient.conf
# RUN ldconfig