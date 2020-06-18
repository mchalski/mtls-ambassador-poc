FROM golang:1.14-alpine3.12 as builder
RUN mkdir -p /go/src/github.com/mendersoftware/mtls-ping
WORKDIR /go/src/github.com/mendersoftware/mtls-ping

ADD ./ .
RUN CGO_ENABLED=0 GOARCH=amd64 go build 

FROM alpine:3.12
COPY --from=builder /go/src/github.com/mendersoftware/mtls-ping/mtls-ping /usr/bin/

RUN mkdir -p /etc/mtls-ping/certs
COPY --from=builder /go/src/github.com/mendersoftware/mtls-ping/certs /etc/mtls-ping/certs

RUN apk add --update ca-certificates && update-ca-certificates
ENTRYPOINT ["/usr/bin/mtls-ping"]
