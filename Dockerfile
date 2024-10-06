FROM docker.io/library/golang:1.23.2-bookworm AS builder

WORKDIR /usr/src/app
COPY . .
RUN make clean build

FROM scratch
COPY --from=builder /usr/src/app/dist/switch-exporter /app
ENTRYPOINT ["/app"]
