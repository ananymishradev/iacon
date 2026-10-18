FROM golang:1.23-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY testbed/loadgen ./testbed/loadgen
RUN CGO_ENABLED=0 go build -o /out/loadgen ./testbed/loadgen

FROM alpine:3.20
COPY --from=builder /out/loadgen /loadgen
ENTRYPOINT ["/loadgen"]
