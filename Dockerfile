FROM golang:1.18-bullseye as build

WORKDIR /build

COPY go.mod go.sum main.go mdb.go mp.go ./
# RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -w' -o oracle_exporter main.go mdb.go mp.go

FROM gcr.io/distroless/base-debian10
COPY --from=builder /build/oracle_exporter /usr/bin/oracle_exporter
ENTRYPOINT ["oracle_exporter"]