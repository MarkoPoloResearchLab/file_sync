# syntax=docker/dockerfile:1

FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go build -o /zync ./cmd/zync

FROM debian:bookworm-slim
COPY --from=build /zync /usr/local/bin/zync
ENTRYPOINT ["zync"]
