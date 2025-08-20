# Stage1: OpenAPI code generation
FROM public.ecr.aws/docker/library/node:lts-alpine AS openapi

RUN apk add openjdk17-jre && npm install @openapitools/openapi-generator-cli -g

WORKDIR /src
COPY contract/openapi/auth-service.yaml ./auth-service.yaml

RUN openapi-generator-cli generate \
    --generator-name go-gin-server \
    --input-spec auth-service.yaml \
    --output generated/openapi-gen \
    --additional-properties=interfaceOnly=true,packageName=openapi,generateMetadata=false,generateGoMod=false &&\
    mkdir -p generated/openapi && cp -r generated/openapi-gen/go/* generated/openapi/ && rm -rf generated/openapi-gen

# Stage2: gRPC code generation
FROM public.ecr.aws/docker/library/golang:alpine AS proto

RUN apk update && apk add --no-cache make protobuf-dev

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /src
COPY contract/proto ./proto

RUN mkdir -p generated/proto && \
    for file in proto/*.proto; do \
        protoc -I=proto \
            --go_out=generated/proto --go_opt=paths=source_relative \
            --go-grpc_out=generated/proto --go-grpc_opt=paths=source_relative \
            "$file"; \
    done

# Stage3: sqlc code generation
FROM public.ecr.aws/docker/library/golang:alpine AS sqlc

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

WORKDIR /src
COPY db db/

RUN mkdir -p generated/sqlc && sqlc generate -f db/sqlc.yaml

# Stage4: build
FROM public.ecr.aws/docker/library/golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=openapi /src/generated/openapi ./generated/openapi
COPY --from=proto /src/generated/proto ./generated/proto
COPY --from=sqlc /src/generated/sqlc ./generated/sqlc

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/auth-service ./cmd/auth-service

# Stage5: Final image
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/auth-service .
COPY migrations /app/migrations
COPY templates /app/templates

# Default command
ENTRYPOINT ["/app/auth-service"]