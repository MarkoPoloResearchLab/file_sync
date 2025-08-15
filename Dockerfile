# syntax=docker/dockerfile:1

FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go build -o /filez-sync ./cmd/filez-sync

FROM debian:bookworm-slim
COPY --from=build /filez-sync /usr/local/bin/filez-sync
ENTRYPOINT ["filez-sync"]
