from golang:1.21-bookworm as builder

workdir /src
copy go.mod go.sum ./
run go mod download
copy . .
run CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /sniproxy .

from debian:12

copy --from=builder /sniproxy /sniproxy
copy config.yaml /sniproxy.conf
copy domains.csv /domains.csv
entrypoint ["/sniproxy", "-config", "/sniproxy.conf"]
