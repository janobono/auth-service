# auth-service

Simple Auth Service created in Golang as an gRPC service.

- [gRPC contract](./api/authservice.proto)

## build

```shell
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

If you have make and golang installed, you can use prepared targets.

- clean - to delete all generated sources
- generate - to generate sources (sqlx and gRPC)
- build - default target will call generate and build everything
- test - run all tests

After successful build there is just a simple client just to test if everything is OK.

```shell
./bin/auth-service-client -addr localhost:50052 -email simple@auth.org -password simple
JWT token: eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiYWRtaW4iLCJtYW5hZ2VyIl0sImV4cCI6MTc0Nzk0NTc0MSwiaWF0IjoxNzQ3OTM4NTQxLCJpc3MiOiJzaW1wbGUiLCJzdWIiOiIxIn0.BDU6MLsUfqHuL0N4zDD47sw4_61SAqrUOwD2OQmQQ9IhCGmhN39puF_HrY7FR8WHySUtQsfrfuZPe7S_x2KtkKhbTjxBwrJyHbvgLge6ia0xqthI1qMT1_UfxO-lvVW3J6jlTqhy33hzDlV3aTBCuXpTp9N3UYmeKNxx0myXaF_UdUFpzd7dhP_Wst_GksQnrbipZpQL0OfldGYX0RMXMpVS4QUxFEp8N8QvfJ1s6Qaa9S0Fjits3uf9DkL5rc31UjFKdFg-zQLLEClplf3FcdWtmOR5c1qXE7_t3ODjVzOIYGhWOMUy-DSlRAoHbf32piI-TwwNbMhdZ0_kpEe5oA
```

## environment variables

| Name             | Required | Description                                 |
|:-----------------|:--------:|:--------------------------------------------|
| ADDRESS          |    X     | Service address                             |
|                  |          |                                             |
| DB_URL           |    X     | Database url                                |
| DB_USER          |    X     | Database user                               |
| DB_PASSWORD      |    X     | Database password                           |
| DB_MAX_CONNS     |    X     | Database connection pooling max connections |
| DB_MIN_CONNS     |    X     | Database connection pooling min connections |
|                  |          |                                             |
| TOKEN_ISSUER     |    X     | Jwt token issuer                            |
| TOKEN_EXPIRES_IN |    X     | Jwt token expiration in minutes             |
