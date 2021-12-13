FROM golang:1 as builder

WORKDIR /build
ADD main.go .
ADD mdb.go .
ADD mp.go .
ADD healthcheck.go .

# RUN go env -w GOPROXY=https://goproxy.cn,direct
# RUN go env -w GO111MODULE="on"

ADD go.mod .
ADD go.sum .
RUN GOOS=linux go build -o oracle_exporter main.go mdb.go mp.go
RUN GOOS=linux go build -o healthcheck healthcheck.go

FROM frolvlad/alpine-glibc:alpine-3.14_glibc-2.33
WORKDIR /app
COPY --from=builder /build/oracle_exporter oracle_exporter
COPY --from=builder /build/healthcheck healthcheck
HEALTHCHECK --interval=5s --timeout=1m CMD /app/healthcheck
ENTRYPOINT ["/app/oracle_exporter"]