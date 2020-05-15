ARG BUILD_DATE=0
ARG COMMIT=0
ARG VERSION=unknown
ARG BINARY=hciscan

FROM golang:stretch as builder
ARG BINARY
RUN mkdir -p $GOPATH/pkg/mod $GOPATH/bin $GOPATH/src /${BINARY}
COPY . /${BINARY}
WORKDIR /${BINARY}

RUN CGO_ENABLED=0 make ${BINARY}

FROM debian:buster-slim
ARG BUILD_DATE
ARG VERSION
ARG BINARY

RUN apt-get update -q \
  && apt-get install -yq bluez-tools bluez \
  && apt-get clean \
  && rm -rf /var/lib/apt/*

ENV ENDPOINT=""

COPY --from=builder /${BINARY}/${BINARY} /image

ENTRYPOINT [ "/image" ]