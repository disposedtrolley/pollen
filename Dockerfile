FROM golang:alpine as build

ENV CGO_ENABLED=1

RUN apk add --no-cache \
    # Important: required for go-sqlite3
    gcc \
    # Required for Alpine
    musl-dev

WORKDIR /app

COPY . .

RUN go build

FROM alpine:latest

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=build /app/pollen pollen

CMD ["./pollen", "tick"]
