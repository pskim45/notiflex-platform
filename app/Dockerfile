FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY . .
RUN go test ./... && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /notiflex-api .

FROM scratch
COPY --from=builder /notiflex-api /notiflex-api
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/notiflex-api"]
