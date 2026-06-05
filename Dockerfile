FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOPRIVATE=git.tongyuan.cc
ENV GONOSUMDB=git.tongyuan.cc
ENV GONOSUMCHECK=1
ENV GOTOOLCHAIN=auto


WORKDIR /build

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN GOTOOLCHAIN=local go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOTOOLCHAIN=local go build -ldflags="-s -w" -o /build/server ./cmd/server/

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /build/server .
COPY --from=builder /build/web/templates ./web/templates
COPY --from=builder /build/config.yaml .

RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["./server"]
