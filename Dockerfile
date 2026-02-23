# syntax=docker/dockerfile:1

FROM alpine AS builder

RUN addgroup -S -g 1000 appgroup && \
    adduser  -S -u 1000 -D -H -G appgroup appuser && \
    mkdir -p /app /config /data

FROM scratch

COPY --from=builder /etc/passwd /etc/group /etc/
COPY --from=builder --chmod=700 --chown=1000:1000 /app/ /app/
COPY --from=builder --chmod=700 --chown=1000:1000 /config/ /config/
COPY --from=builder --chmod=700 --chown=1000:1000 /data/ /data/

ENV XDG_CONFIG_HOME=/config \
    XDG_DATA_HOME=/data

VOLUME ["/config", "/data"]

WORKDIR /app
ARG TARGETARCH
USER appuser:appgroup
COPY --chmod=700 --chown=1000:1000 dist/thrippy_linux_${TARGETARCH}*/thrippy .

ENTRYPOINT ["./thrippy", "server"]
EXPOSE 14460/tcp 14470/tcp
HEALTHCHECK --interval=30s --timeout=3s --retries=2 CMD ["./thrippy", "health-check"]
