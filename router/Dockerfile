#### - BUILDER - ####
FROM golang:1.23.8 AS builder

WORKDIR /app

COPY ./go.mod go.mod
COPY ./go.sum go.sum
RUN go mod download

COPY ./ ./

RUN go build -o bin/bck ./bck && \
    go build -o bin/fwd ./fwd


#### - SERVER - ####
FROM alpine:3.21.3 AS server

RUN apk add --no-cache gcompat=1.1.0-r4 libstdc++=13.2.1_git20231014-r0

WORKDIR /app

COPY --from=builder /bin/bck ./bck
COPY --from=builder /bin/fwd ./fwd

RUN adduser --system --no-create-home nonroot
USER nonroot

# CMD ["./main"]