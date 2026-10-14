FROM golang:1.23-bookworm AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/iacon ./cmd/iacon

FROM alpine:3.20

RUN addgroup -S iacon && adduser -S -G iacon iacon \
    && apk add --no-cache wget
WORKDIR /app
COPY --from=builder /out/iacon /app/iacon
USER iacon
EXPOSE 8090
ENTRYPOINT ["/app/iacon"]
