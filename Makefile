BINARIES = auth-service auth-service-client

.PHONY: test

default: build

clean:
	@echo "  >  Cleaning build cache"
	@-rm -rf bin && go clean ./...

build:
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