FROM golang:1 as builder

WORKDIR /build
ADD main.go .
ADD mdb.go .
ADD mp.go .

# RUN go env -w GOPROXY=https://goproxy.cn,direct
# RUN go env -w GO111MODULE="on"

ADD go.mod .
ADD go.sum .
RUN GOOS=linux go build -o oracle_exporter main.go mdb.go mp.go

FROM frolvlad/alpine-glibc:alpine-3.14_glibc-2.33
WORKDIR /app
COPY --from=builder /build/oracle_exporter oracle_exporter
ENTRYPOINT ["/app/oracle_exporter"]