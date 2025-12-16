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

FROM golang:1.22-alpine AS dev

WORKDIR /app

RUN apk add --no-cache git make bash

COPY go.mod ./
RUN go mod download

COPY . .
RUN go mod tidy

CMD ["make", "test"]

