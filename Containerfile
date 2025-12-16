FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod ./
RUN go mod download

COPY . .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gitopsi ./cmd/gitopsi

FROM alpine:3.19 AS runtime

RUN apk add --no-cache ca-certificates git

COPY --from=builder /gitopsi /usr/local/bin/gitopsi

WORKDIR /workspace

ENTRYPOINT ["gitopsi"]

FROM golang:1.22 AS dev

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    git make bash \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go mod tidy

CMD ["go", "test", "-v", "-race", "./..."]

