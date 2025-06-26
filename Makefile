.PHONY: clean generate generate-proto generate-openapi generate-sqlc build fmt test vet

default: build

clean:
	@echo "  >  Cleaning build cache"
	@go clean ./...
	@rm -rf bin
	@rm -rf generated

generate-proto:
	@echo "  > Generate proto source files"
	mkdir -p generated/proto && \
	for file in contract/proto/*.proto; do \
		protoc -I=contract/proto \
			--go_out=generated/proto --go_opt=paths=source_relative \
			--go-grpc_out=generated/proto --go-grpc_opt=paths=source_relative \
			"$$file"; \
	done

generate-openapi:
	@echo "  > Generate openapi source files"
	openapi-generator-cli generate \
	--generator-name go-gin-server \
	--input-spec contract/openapi/auth-service.yaml \
	--output generated/openapi-gen \
	--additional-properties=interfaceOnly=true,packageName=openapi,generateMetadata=false,generateGoMod=false &&\
	mkdir -p generated/openapi && cp -r generated/openapi-gen/go/* generated/openapi/ && rm -rf generated/openapi-gen

generate-sqlc:
	@echo "  >  Generate sqlc files"
	sqlc generate -f db/sqlc.yaml

generate: generate-proto generate-openapi generate-sqlc

build: generate
	go build -o bin/auth-service ./cmd/auth-service/main.go

fmt:
	@echo "  >  Formatting code"
	@go fmt ./...

test:
	@echo "  >  Executing unit tests"
	@go test -v ./...

vet:
	@echo "  >  Checking code with vet"
	@go vet ./...