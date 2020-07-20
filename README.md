# mTLS Ambassador Service

Proof of concept mTLS Ambassador which:

- stands in front of the Mender API Gateway
- protects the device API with mTLS by requiring valid mTLS certs on every call
- checks client certs against a single supplied tenant's CA cert
- in general - proxies the API calls 1:1 over plain HTTPS
- with one exception: on auth requests, injects an extra preauthorization call just before forwarding

As a result, mTLS aware devices are automatically accepted into Mender, just on the basis
of their valid certs.

The authorization flow is unchanged, auth requests still play a central role,
and the client will still obtain a JWT for further communication.

No user action is required though for successful auth (i.e. manual admission, manual preauth).

## Build
A standard builder dockerfile is included and two deployment methods:
- docker-compose
- k8s

For docker-compose, simply issue `docker-compose build` and the service is ready to go.

For k8s deployments, there's strictly no need to build anything; it pulls prebuilt docker images, e.g.:
`registry.mender.io/mendersoftware/mtls-ambassador:1.0.2`

Building, tagging and publishing images was done manually, as in:
```
docker build -t registry.mender.io/mendersoftware/mtls-ambassador:1.0.2 .
docker push registry.mender.io/mendersoftware/mtls-ambassador:1.0.2
```

## Run
The Ambassador works against the hardcoded `staging.hosted.mender.io:443` backend (parametrize this as an improvement).

You need several pieces of configuration:
- an mTLS CA certificate in PEM format (=tenant's certificate), e.g.:
    - `certs/tenant-ca/tenant-foo.ca.pem`,
- a regular HTTPS server certificate + private key, e.g.:
    - `certs/server/server.crt`,
    - `certs/server/server.key`
- at least one client mTLS certificate signed by the CA + private key for testing, e.g.:
    - `certs/tenant-foo.client.1.crt`,
    - `certs/tenant-foo.client.1.key`
    - (these are signed by `certs/tenant-ca/tenant-foo.ca.key`)
- a created tenant and a user

### docker-compose
The compose setup has limited customizability and is best used
for quick test runs - it uses the default certs from `certs/` (tenant's CA and server cert/key).

To run:
1. create your user and tenant
2. set `MTLS_MENDER_USER` and `MTLS_MENDER_PASS` env vars in docker-compose.yml
3. run `docker-compose build`
4. run `docker-compose up`

The Ambassador is now running at `https://localhost:8080`.

You should see a successful startup sequence like this:
```
mtls-ambassador_1  | 2020/06/26 13:02:37 reading config
mtls-ambassador_1  | 2020/06/26 13:02:37 logging in to Mender to get mgmt token, user: mtls@mender.io
mtls-ambassador_1  | 2020/06/26 13:02:38 logging in to Mender: success
mtls-ambassador_1  | 2020/06/26 13:02:38 starting server
mtls-ambassador_1  | [GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.
mtls-ambassador_1  |
mtls-ambassador_1  | [GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
mtls-ambassador_1  |  - using env:      export GIN_MODE=release
mtls-ambassador_1  |  - using code:     gin.SetMode(gin.ReleaseMode)
mtls-ambassador_1  |
mtls-ambassador_1  | [GIN-debug] GET    /ping                     --> main.handlePing (3 handlers)
mtls-ambassador_1  | [GIN-debug] GET    /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
mtls-ambassador_1  | [GIN-debug] POST   /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
mtls-ambassador_1  | [GIN-debug] PUT    /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
mtls-ambassador_1  | [GIN-debug] PATCH  /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
mtls-ambassador_1  | [GIN-debug] HEAD   /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
mtls-ambassador_1  | [GIN-debug] OPTIONS /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
mtls-ambassador_1  | [GIN-debug] DELETE /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
mtls-ambassador_1  | [GIN-debug] CONNECT /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
mtls-ambassador_1  | [GIN-debug] TRACE  /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
```

Use the provided client certs in `certs/` to test it out.

### k8s on AWS
Deployment and Service manifests for an AWS deploment are available in `/k8s`. These support full customizability of your credentials and certificates.

It's assumed that you've correctly configured access to your cluster and are in the right kubernetes context.

Start with creating several secrets which will map to sensitive pod env vars and mounted cert files (values are just default examples - adjust accordingly):

1. `mtls-server-cert`
    - corresponds to server's HTTPS cert + key
    - it's a 2 value secret mounted under `/etc/mtls/certs/server`
    - `kubectl create secret generic mtls-server-cert --from-file=certs/server/server.crt --from-file=certs/server/server.key`
2. `mtls-tenant-pem`
    - corresponds to the tenant's CA cert in PEM format
    - mounted under `/etc/mtls/certs/tenant-ca`
    - `kubectl create secret generic mtls-tenant-ca-pem --from-file certs/tenant-ca/tenant.ca.pem`
3. `mender-creds`
    - 2 value secret, corresponds to env vars `MTLS_MENDER_USER` and `MTLS_MENDER_PASS` (your Ambassador user)
    - `kubectl create secret generic mender-creds --from-literal=username='...' --from-literal=password='...'`
4. `kubectl apply -f k8s/deployment.yaml`
5. `kubectl apply -f k8s/service.yaml`

Run `kubectl get services` to obtain the DNS name of your Ambassador instance:

```
NAME                      TYPE           CLUSTER-IP      EXTERNAL-IP                                                                     PORT(S)          AGE
mtls-ambassador-service   LoadBalancer   10.100.187.64   aa12d2cf0573e481cabd0b84b3e3448a-f4e2ba095b69e99a.elb.us-east-1.amazonaws.com   8080:32075/TCP   1m
```

Find your pod via `kubectl get pods`:

```
NAME                                         READY   STATUS    RESTARTS   AGE
mtls-ambassador-deployment-c9c4b64fc-x742t   1/1     Running   0          1m
```

Check logs to verify a successful startup:

```
kubectl logs -f mtls-ambassador-deployment-c9c4b64fc-x742t


2020/06/25 17:54:16 reading config
2020/06/25 17:54:16 logging in to Mender to get mgmt token, user: mtls@mender.io
2020/06/25 17:54:18 logging in to Mender: success
2020/06/25 17:54:18 starting server
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> main.handlePing (3 handlers)
[GIN-debug] GET    /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
[GIN-debug] POST   /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
[GIN-debug] PUT    /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
[GIN-debug] PATCH  /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
[GIN-debug] HEAD   /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
[GIN-debug] OPTIONS /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
[GIN-debug] DELETE /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
[GIN-debug] CONNECT /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
[GIN-debug] TRACE  /api/devices/*path        --> main.setupMenderApiHandler.func2 (3 handlers)
```

(Note: the above instructions also hold for minikube deployments, just replace `service.yaml` with `service.minikube.yaml`)

## Test

For the simplest possible test of your configuration, try the Ambassador's `/ping` probe with curl:

```
curl --cert  certs/tenant-foo.client.1.crt --key certs/tenant-foo.client.1.key -ivk https://aa12d2cf0573e481cabd0b84b3e3448a-f4e2ba095b69e99a.elb.us-east-1.amazonaws.com:8080/ping

* Connected to aa12d2cf0573e481cabd0b84b3e3448a-f4e2ba095b69e99a.elb.us-east-1.amazonaws.com (52.202.100.58) port 8080 (#0)                                     [48/847]
* ALPN, offering h2
* ALPN, offering http/1.1
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/certs/ca-certificates.crt
  CApath: /etc/ssl/certs
* TLSv1.3 (OUT), TLS handshake, Client hello (1):
* TLSv1.3 (IN), TLS handshake, Server hello (2):
* TLSv1.3 (IN), TLS Unknown, Certificate Status (22):
* TLSv1.3 (IN), TLS Unknown, Certificate Status (22):
...

GET /ping HTTP/2
> Host: aa12d2cf0573e481cabd0b84b3e3448a-f4e2ba095b69e99a.elb.us-east-1.amazonaws.com:8080
> User-Agent: curl/7.58.0
> Accept: */
...

{"message":"pong"}
```

If you see the `pong` message, your client certs correctly validate against the tenant's CA cert. You'll see details of the TLS
handshake to confirm it.

To actually test out the proxying and automatic preauth it's best to use an actual device.

We'll use a slightly modified mender bash client (`extra/mender-client.sh`):


1. Setup `MENDER_CLIENT_CERT` and `MENDER_CLIENT_KEY` env vars, e.g.:
    - `export MENDER_CLIENT_CERT=../mtls-ambassador/certs/tenant-foo.client.1.crt`
    - `export MENDER_CLIENT_KEY=../mtls-ambassador/certs/tenant-foo.client.1.key`
2. Setup an `extra/keys` directory for the client:
    - copy `$MENDER_CLIENT_KEY` as `private.key`
    - copy a public key extracted from `$MENDER_CLIENT_CERT` as `public.key`
    - (`/extra` has pre-extracted keys for provided client certs)
3. Setup the correct Ambassador URL, e.g.:
    - `export MENDER_SERVER_URL=https://aa12d2cf0573e481cabd0b84b3e3448a-f4e2ba095b69e99a.elb.us-east-1.amazonaws.com:8080`
4. Run the client:
    - `./mender-client.sh -t <tenant_token> -d rpi4`

The client should pass through authentication, upload inventory and go into deployment polling loop - as usual.

A new accepted device should also appear in the UI.

Service log excerpt:

```
[GIN] 2020/06/26 - 15:17:09 | 200 |  306.205545ms |  178.200.237.18 | POST     "/api/devices/v1/authentication/auth_requests"
2020/06/26 15:17:10 intercepted POST /auth_requests
2020/06/26 15:17:10 client cert details:                                                                                                                                2020/06/26 15:17:10 subject CN=device 1,O=Tenant Foo,ST=Some-State,C=US
2020/06/26 15:17:10 issuer CN=Tenant Foo,O=Tenant Foo,ST=Some-State,C=US                                                                                                2020/06/26 15:17:10 verifying client key
2020/06/26 15:17:10 client key: -----BEGIN PUBLIC KEY-----                                                                                                              MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxex+WqphwqgkDfPWJZZt
nXHvvRVhG6j3+q45skFyC8Wa0s3Re8TJIOUKXwx6YWrl333zqq+KOyiYPcaosVl+
y3IzRkT7hNnyExMFEZi2eygg6SINW4QtNIvTASaQqn831QyfkaaQLGl1vqNe262l
3uabUAYkaDf9Kaz/RbB5jjCse9d28pvSIlPjZGYd7sdj2qVSiOMDWh1tiCdr63Xl
/oiGMUMU1qlX9Tv2nRxJfzTLplSJh1C5sT5cqbv7EfxokWHLFYysSOUQLlINphxw
P5Yk1M0n4WaU6a0FQNmqN+EtnSjyjvM8VINnBMLjOVP7N1siax3e/lKG0bxyWR+T
hQIDAQAB
-----END PUBLIC KEY-----

2020/06/26 15:17:10 client key matches auth req key
2020/06/26 15:17:10 verifying client key: success
2020/06/26 15:17:10 preauthorizing
2020/06/26 15:17:10 proxying auth request to Mender
[GIN] 2020/06/26 - 15:17:10 | 200 |   23.072598ms |  178.200.237.18 | POST     "/api/devices/v1/authentication/auth_requests"
[GIN] 2020/06/26 - 15:17:10 | 200 |    5.216645ms |  178.200.237.18 | PATCH    "/api/devices/v1/inventory/device/attributes"
[GIN] 2020/06/26 - 15:17:10 | 204 |    9.663014ms |  178.200.237.18 | GET      "/api/devices/v1/deployments/device/deployments/next?artifact_name=release-v1&device_typ$
=rpi4"
[GIN] 2020/06/26 - 15:17:16 | 204 |    7.888859ms |  178.200.237.18 | GET      "/api/devices/v1/deployments/device/deployments/next?artifact_name=release-v1&device_typ$
=rpi4"
[GIN] 2020/06/26 - 15:17:21 | 204 |    5.343698ms |  178.200.237.18 | GET      "/api/devices/v1/deployments/device/deployments/next?artifact_name=release-v1&device_typ$
=rpi4"
...
```

### Implementation notes

#### k8s AWS config
TODO

#### Ambassador code
- POC quality - basic separation of concerns, but not great for testability (no interfaces, etc)
- no unit tests, tested heavily by hand
- original idea: use `gin/gonic`, proxy by repacking request manually
    - instead: used `net/http` ReverseProxy which does that and more
    - e.g. deals with 'hop-by-hop' headers, possibly more conventions
    - double check, but consider using it for production
- missing functional bits:
    - no revocation support
    - no management API token refresh
