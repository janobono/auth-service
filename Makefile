BINARIES = auth-service auth-service-client

.PHONY: test

default: build

clean:
	@echo "  >  Cleaning build cache"
	@go clean ./...
	@rm -rf bin
	@rm -f ./internal/repository/*.go
	@rm -f ./api/*.go

generate:
	@echo "  >  Generate source files"
	@sqlc generate
	@protoc --go_out=. --go_opt=paths=source_relative \--go-grpc_out=. --go-grpc_opt=paths=source_relative api/authservice.proto

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