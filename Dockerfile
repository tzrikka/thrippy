# syntax=docker/dockerfile:1

FROM alpine:latest AS builder
RUN mkdir -p /home/nonroot/.config/thrippy /home/nonroot/.local/share/thrippy && \
    chown -R 65532:65532 /home/nonroot && \
    chmod -R 700 /home/nonroot

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /home/nonroot
USER nonroot:nonroot

COPY --from=builder /home/nonroot/ .
VOLUME ["/home/nonroot/.config/thrippy", "/home/nonroot/.local/share/thrippy"]

ARG TARGETARCH
COPY --chmod=700 --chown=nonroot:nonroot dist/thrippy_linux_${TARGETARCH}*/thrippy .
ENTRYPOINT ["./thrippy", "server"]
EXPOSE 14460/tcp 14470/tcp

HEALTHCHECK --interval=10s --timeout=3s --retries=3 CMD ["./thrippy", "health-check"]
