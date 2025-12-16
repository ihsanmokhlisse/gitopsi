# =============================================================================
# GITOPSI Container Build
# =============================================================================
# Multi-stage build for minimal, secure container image
# - Uses scratch base for minimal attack surface
# - Only includes the gitopsi binary (no shell, no OS packages)
# - Runs as non-root user
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder - Compile the Go binary
# -----------------------------------------------------------------------------
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum* ./
RUN go mod download && go mod verify

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags="-s -w \
        -X github.com/ihsanmokhlisse/gitopsi/internal/cli.Version=${VERSION} \
        -X github.com/ihsanmokhlisse/gitopsi/internal/cli.Commit=${COMMIT} \
        -X github.com/ihsanmokhlisse/gitopsi/internal/cli.BuildDate=${BUILD_DATE}" \
    -o /gitopsi ./cmd/gitopsi

RUN /gitopsi version

# -----------------------------------------------------------------------------
# Stage 2: Runtime - Minimal production image (DISTROLESS)
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

LABEL org.opencontainers.image.title="gitopsi"
LABEL org.opencontainers.image.description="GitOps Repository Generator CLI"
LABEL org.opencontainers.image.source="https://github.com/ihsanmokhlisse/gitopsi"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.vendor="gitopsi"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder --chown=nonroot:nonroot /gitopsi /usr/local/bin/gitopsi

WORKDIR /workspace

USER nonroot:nonroot

ENTRYPOINT ["gitopsi"]
CMD ["--help"]

# -----------------------------------------------------------------------------
# Stage 3: Runtime with Shell - For users who need shell access
# -----------------------------------------------------------------------------
FROM alpine:3.20 AS runtime-shell

LABEL org.opencontainers.image.title="gitopsi"
LABEL org.opencontainers.image.description="GitOps Repository Generator CLI (with shell)"

RUN apk add --no-cache ca-certificates git \
    && adduser -D -u 1000 gitopsi

COPY --from=builder /gitopsi /usr/local/bin/gitopsi

WORKDIR /workspace
RUN chown gitopsi:gitopsi /workspace

USER gitopsi

ENTRYPOINT ["gitopsi"]
CMD ["--help"]

# -----------------------------------------------------------------------------
# Stage 4: Development - Full development environment
# -----------------------------------------------------------------------------
FROM golang:1.23 AS dev

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    git make bash curl \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0 \
    && go install golang.org/x/vuln/cmd/govulncheck@latest

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go mod tidy

CMD ["go", "test", "-v", "-race", "./..."]
