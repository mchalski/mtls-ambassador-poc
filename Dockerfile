FROM golang:1.14-alpine3.12 as builder
RUN mkdir -p /go/src/github.com/mendersoftware/mtls-ambassador-poc
WORKDIR /go/src/github.com/mendersoftware/mtls-ambassador-poc

ADD ./ .
RUN CGO_ENABLED=0 GOARCH=amd64 go build 

FROM alpine:3.12
COPY --from=builder /go/src/github.com/mendersoftware/mtls-ambassador-poc/mtls-ambassador-poc /usr/bin/

RUN mkdir -p /etc/mtls/certs
COPY --from=builder /go/src/github.com/mendersoftware/mtls-ambassador-poc/certs /etc/mtls/certs

RUN apk add --update ca-certificates && update-ca-certificates
ENTRYPOINT ["/usr/bin/mtls-ambassador-poc"]
