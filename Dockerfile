FROM golang:1.14-alpine3.12 as builder
RUN mkdir -p /go/src/github.com/mendersoftware/mtls-ambassador
WORKDIR /go/src/github.com/mendersoftware/mtls-ambassador

ADD ./ .
RUN CGO_ENABLED=0 GOARCH=amd64 go build 

FROM alpine:3.12
COPY --from=builder /go/src/github.com/mendersoftware/mtls-ambassador/mtls-ambassador /usr/bin/

RUN mkdir -p /etc/mtls/certs
COPY --from=builder /go/src/github.com/mendersoftware/mtls-ambassador/certs /etc/mtls/certs

COPY --from=builder /go/src/github.com/mendersoftware/mtls-ambassador/config.yaml /etc/mtls/config.yaml

RUN apk add --update ca-certificates && update-ca-certificates
ENTRYPOINT ["/usr/bin/mtls-ambassador", "--config", "/etc/mtls/config.yaml"]
