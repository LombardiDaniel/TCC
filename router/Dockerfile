#### - BUILDER - ####
FROM golang:1.23.8 AS builder

WORKDIR /app

COPY ./go.mod go.mod
COPY ./go.sum go.sum
RUN go mod download

COPY ./ ./

RUN go build -o /bin/bck ./bck && \
    go build -o /bin/fwd ./fwd

#### - SERVER - ####
FROM alpine:3.21.3 AS server

RUN apk add --no-cache gcompat libstdc++

WORKDIR /app

COPY --from=builder /bin/bck ./cmd/bck
COPY --from=builder /bin/fwd ./cmd/fwd

RUN adduser --system --no-create-home nonroot
USER nonroot

# CMD ["./main"]