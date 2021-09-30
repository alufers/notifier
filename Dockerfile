FROM golang:alpine AS builder
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git bash && mkdir -p /build/notifier && go get -u github.com/swaggo/swag/cmd/swag

WORKDIR /build/notifier

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download -json

COPY . .

RUN mkdir -p /app && go generate . &&  CGO_ENABLED=0 go build -ldflags='-s -w -extldflags="-static"' -o /app/notifier

FROM scratch AS bin-unix
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/notifier /app/notifier

ENTRYPOINT ["/app/notifier"]
