# Stage1: OpenAPI code generation
FROM public.ecr.aws/docker/library/node:lts-alpine AS openapi

RUN apk add openjdk17-jre && npm install @openapitools/openapi-generator-cli -g

WORKDIR /src
COPY openapi ./openapi

RUN openapi-generator-cli generate \
    --generator-name go-gin-server \
    --input-spec openapi/auth-service.yaml \
    --output gen/openapi \
    --additional-properties=interfaceOnly=true,packageName=authrest,generateMetadata=false,generateGoMod=false &&\
    mkdir -p gen/authrest && cp -r gen/openapi/go/* gen/authrest/ && rm -rf gen/openapi

# Stage2: gRPC code generation
FROM public.ecr.aws/docker/library/golang:alpine AS proto

RUN apk update && apk add --no-cache make protobuf-dev

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /src
COPY proto ./proto

RUN mkdir -p gen/authgrpc && \
    for file in proto/*.proto; do \
        protoc -I=proto \
            --go_out=gen/authgrpc --go_opt=paths=source_relative \
            --go-grpc_out=gen/authgrpc --go-grpc_opt=paths=source_relative \
            "$file"; \
    done

# Stage3: sqlc code generation
FROM public.ecr.aws/docker/library/golang:alpine AS sqlc

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

WORKDIR /src
COPY sqlc.yaml .
COPY db db/

RUN mkdir -p gen/db/repository && sqlc generate

# Stage4: build
FROM public.ecr.aws/docker/library/golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=openapi /src/gen/authrest ./gen/authrest
COPY --from=proto /src/gen/authgrpc ./gen/authgrpc
COPY --from=sqlc /src/gen/db ./gen/db

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/auth-service ./cmd/auth-service

# Stage5: Final image
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/auth-service .
COPY db/migrations db/migrations

# Default command
ENTRYPOINT ["/app/auth-service"]