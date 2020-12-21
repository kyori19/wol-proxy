FROM golang:1.15.6-alpine3.12 as build

RUN mkdir -p /opt/wol-proxy/
WORKDIR /opt/wol-proxy/
COPY . /opt/wol-proxy/

RUN go build

FROM alpine:3.12.3

RUN mkdir -p /opt/wol-proxy/
WORKDIR /opt/wol-proxy/
COPY --from=build /opt/wol-proxy/wol-proxy \
  /opt/wol-proxy/favicon.ico \
  /opt/wol-proxy/wol.gohtml /opt/wol-proxy/

ENTRYPOINT ["/opt/wol-proxy/wol-proxy"]
