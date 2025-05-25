PROTO_DIR := proto
PROTO_OUT := gen/authgrpc
PROTO_FILES := auth-service

BINARIES := auth-service auth-grpc-client auth-http-client

.PHONY: clean generate generate-proto generate-openapi generate-sqlx build fmt test vet

default: build

clean:
	@echo "  >  Cleaning build cache"
	@go clean ./...
	@rm -rf bin
	@rm -rf gen

generate-proto:
	@echo "  > Generate proto source files"
	@cd $(PROTO_DIR) && \
	for file in $(PROTO_FILES); do \
  		mkdir -p ../$(PROTO_OUT) && \
		protoc \
			-I=. \
			--go_out=../$(PROTO_OUT) --go_opt=paths=source_relative \
			--go-grpc_out=../$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
			$$file.proto ; \
	done

generate-openapi:
	@echo "  > Generate openapi source files"
	openapi-generator-cli generate \
	--generator-name go-gin-server \
	--input-spec openapi/auth-service.yaml \
	--output gen/openapi \
	--additional-properties=interfaceOnly=true,packageName=authrest,generateMetadata=false,generateGoMod=false &&\
	mkdir -p gen/authrest && cp -r gen/openapi/go/* gen/authrest/ && rm -rf gen/openapi

generate-sqlx:
	@echo "  >  Generate sqlx files"
	@mkdir -p gen/db/repository && sqlc generate

generate: generate-proto generate-openapi generate-sqlx

build: generate
	@for b in $(BINARIES); do \
  		echo "  >  Building binary" $$b ;\
		go build -o bin/$$b ./cmd/$$b/main.go ;\
	done

fmt:
	@echo "  >  Formatting code"
	@go fmt ./...

test:
	@echo "  >  Executing unit tests"
	@go test -v ./...

vet:
	@echo "  >  Checking code with vet"
	@go vet ./...