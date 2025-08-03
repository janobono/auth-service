# auth-service

Auth Service written in Golang.

- [openapi contract](contract/openapi/auth-service.yaml)
- [gRPC contract](contract/proto/auth-service.proto)
- [sql schema](db/schema.sql)

## build

```bash
./build.sh
```

or

```shell
docker build -t auth-service:latest .
```

## run

```shell
docker compose up
```

## stop

```shell
docker compose down
```

## make

If you have `make`, `Go`, `node.js` and `protoc` ([How to install the protocol buffer compiler](https://protobuf.dev/installation/)) installed, you can use these prepared targets:

- `tools` - to install all tools and modules
- `clean` - to delete all generated sources
- `generate-openapi` - generate source files from openapi (http) into `generated/openapi`
- `generate-proto` - generate source files from proto (gRPC) into `generated/proto`
- `generate` - to generate all sources
- `build` - default target will call generate and build everything
- `fmt` - format code
- `test` - run tests
- `vet` - check code with vet

## local run

To start the service locally, follow these steps:

1. **Set Environment Variables**  
   Create a `.env.local` file with the required environment variables, or set them manually in your system.
2. **Run necessary services**  
   Run necessary services with docker command `docker compose -f infra.yaml up`.

## environment variables

| Name                                  | Example                                                      | Description                                                                                                   |
|:--------------------------------------|:-------------------------------------------------------------|:--------------------------------------------------------------------------------------------------------------|
| PROD                                  | false                                                        | Production mode flag - log level is switched from debug to info                                               |
| GRPC_ADDRESS                          | :50052                                                       | Service gRPC port                                                                                             |
| HTTP_ADDRESS                          | :8080                                                        | Service http port                                                                                             |
| CONTEXT_PATH                          | /api                                                         | Rest api context path                                                                                         |
|                                       |                                                              |                                                                                                               |
| DB_URL                                | localhost:5432/app                                           | Database url                                                                                                  |
| DB_USER                               | app                                                          | Database user                                                                                                 |
| DB_PASSWORD                           | app                                                          | Database password                                                                                             |
| DB_MAX_CONNECTIONS                    | 5                                                            | Database connection pooling max connections                                                                   |
| DB_MIN_CONNECTIONS                    | 2                                                            | Database connection pooling min connections                                                                   |
| DB_MIGRATIONS_URL                     | file://./migrations                                          | Database migrations directory url                                                                             |
|                                       |                                                              |                                                                                                               |
| MAIL_HOST                             | localhost                                                    | SMTP service host                                                                                             |
| MAIL_PORT                             | 1025                                                         | SMTP service port                                                                                             |
| MAIL_USER                             | app@auth.org                                                 | Default application mail account                                                                              |
| MAIL_PASSWORD                         |                                                              | Default application mail account password                                                                     |
| MAIL_AUTH_ENABLED                     | false                                                        | Enabled/Disable mail authentication                                                                           |
| MAIL_TLS_ENABLED                      | false                                                        | Enabled/Disable mail TLS                                                                                      |
| MAIL_TEMPLATE_URL                     |                                                              | Path or URL to the HTML mail template file. If empty, a built-in default template will be used.               |
| MAIL_TEMPLATE_RELOAD_INTERVAL         | 0                                                            | Interval in minutes for automatically reloading the mail template. A value of 0 disables automatic reloading. |
|                                       |                                                              |                                                                                                               |
| SECURITY_READ_AUTHORITIES             | manager,employee                                             | Default read authorities                                                                                      |
| SECURITY_WRITE_AUTHORITIES            | admin                                                        | Default write authorities                                                                                     |
| SECURITY_DEFAULT_USERNAME             | simple@auth.org                                              | Default user created at first start - remove after your admin account is created                              |
| SECURITY_DEFAULT_PASSWORD             | $2a$10$gRKMsjTON2A4b5PDIgjej.EZPvzVaKRj52Mug/9bfQBzAYmVF0Cae | Default user password created at first start                                                                  |
| SECURITY_TOKEN_ISSUER                 | simple                                                       | token issuer                                                                                                  |
| SECURITY_ACCESS_TOKEN_EXPIRES_IN      | 30                                                           | access token expiration in minutes                                                                            |
| SECURITY_ACCESS_TOKEN_JWK_EXPIRES_IN  | 720                                                          | access token jwt key expiration in minutes                                                                    |
| SECURITY_REFRESH_TOKEN_EXPIRES_IN     | 10080                                                        | refresh token expiration in minutes                                                                           |
| SECURITY_REFRESH_TOKEN_JWK_EXPIRES_IN | 20160                                                        | refresh token jwt key expiration in minutes                                                                   |
| SECURITY_CONTENT_TOKEN_EXPIRES_IN     | 10080                                                        | content token expiration in minutes                                                                           |
| SECURITY_CONTENT_TOKEN_JWK_EXPIRES_IN | 20160                                                        | content token jwt key expiration in minutes                                                                   |
|                                       |                                                              |                                                                                                               |
| CORS_ALLOWED_ORIGINS                  | http://localhost:3000                                        | Allowed origins for CORS                                                                                      |
| CORS_ALLOWED_METHODS                  | GET,POST,PUT,PATCH,DELETE                                    | Allowed HTTP methods                                                                                          |
| CORS_ALLOWED_HEADERS                  | Origin,Content-Type,Accept,Authorization                     | Allowed HTTP headers                                                                                          |
| CORS_EXPOSED_HEADERS                  | Content-length                                               | Exposed headers in CORS                                                                                       |
| CORS_ALLOW_CREDENTIALS                | true                                                         | Whether credentials are allowed in CORS                                                                       |
| CORS_MAX_AGE                          | 12                                                           | Max age (in hours) for CORS preflight response caching                                                        |
|                                       |                                                              |                                                                                                               |
| APP_MAIL_CONFIRMATION                 | true                                                         | Enable/Disable sending confirmation token as a part of the signUp process                                     |
| APP_CONFIRMATION_URL                  | http://localhost:3000/confirm                                | If confirmation is enabled this url with token si part of the signUp information mail                         |

### Security Warning

> ⚠️ **Default Credentials**
>
> The default admin user (`simple@auth.org`) is created automatically on the first run.
> **Make sure to remove or replace it after your own admin account is created**.
> Leaving default credentials active in production is a serious security risk.
