# syntax=docker/dockerfile:1
FROM golang:1.20-buster AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .
RUN make build
RUN ./build/stockticker -h

FROM ubuntu:latest
LABEL org.opencontainers.image.source = "https://github.com/tcuthbert/stockticker"
ARG UID=1000
ARG GID=1000

RUN apt-get update \
  && apt-get install ca-certificates -y && update-ca-certificates

RUN groupadd -g "${GID}" stockticker \
  && useradd --home-dir /app --create-home -u "${UID}" -g "${GID}" stockticker

WORKDIR /app
COPY --from=builder /app/build/stockticker .

EXPOSE 5000
USER stockticker
CMD ["/app/stockticker"]
