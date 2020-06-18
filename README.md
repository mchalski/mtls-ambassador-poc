mTLS 'ping' service
======

For authenticated clients responds to:

```
GET /ping

{"message":"pong"}
```

- tenant CA cert is baked in (certs/tenant-foo.ca.crt, bundled in tenant.ca.pem; self signed)
- 2 signed client certs are generated for testing (= tenant's devices)
- server cert/key must be substituted (by default - self signed for 'localhost', only for local testing)
- default port is 8080, substitute if needed

## Run

```
docker-compose build

MTLS_PING_CERT=<path> MTLS_PING_KEY=<path> MTLS_PING_PORT=<port> docker-compose up
```

## Test
```
curl --cert  certs/tenant-foo.client.1.crt --key certs/tenant-foo.client.1.key -ivk https://<server:port>/ping
```


