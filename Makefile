
LOCAL_BIN := $(shell pwd)/bin
PROTO_DIR := api/telegram_service
GOOGLEAPIS_DIR := $(LOCAL_BIN)/googleapis
GOLANGCI_LINT := $(LOCAL_BIN)/golangci-lint
PROTOC_VERSION := 25.1
PROTOC := $(LOCAL_BIN)/protoc
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(OS)-$(ARCH).zip
GRPC_GATEWAY_DIR := $(LOCAL_BIN)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)


ifeq ($(ARCH),x86_64)
    ARCH := x86_64
else ifeq ($(ARCH),aarch64)
    ARCH := aarch_64
endif
ifeq ($(OS),darwin)
    OS := osx
    ifeq ($(ARCH),x86_64)
        ARCH := x86_64
    else
        ARCH := aarch_64
    endif
endif


.PHONY: all
all: deps generate

.PHONY: deps
deps: clean install-protoc install-go-plugins install-googleapis

.PHONY: install-protoc
install-protoc:
	@echo "📥 Downloading protoc $(PROTOC_VERSION) for $(OS)/$(ARCH)..."
	mkdir -p $(LOCAL_BIN) /tmp/protoc_install
	curl -L $(PROTOC_URL) -o /tmp/protoc_install/protoc.zip
	cd /tmp/protoc_install && unzip -o protoc.zip
	cp /tmp/protoc_install/bin/protoc $(LOCAL_BIN)/
	cp -r /tmp/protoc_install/include $(LOCAL_BIN)/
	chmod +x $(LOCAL_BIN)/protoc
	rm -rf /tmp/protoc_install
	@echo "✅ protoc installed to $(LOCAL_BIN)"

.PHONY: install-go-plugins
install-go-plugins:
	@echo "🔧 Installing protoc Go plugins..."
	mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	GOBIN=$(LOCAL_BIN) go install go.uber.org/mock/mockgen@latest
	@echo "✅ Go plugins installed to $(LOCAL_BIN)"

.PHONY: install-googleapis
install-googleapis:
	@if [ ! -d $(GOOGLEAPIS_DIR) ]; then \
		echo "📥 Downloading googleapis..."; \
		git clone --depth 1 https://github.com/googleapis/googleapis.git $(GOOGLEAPIS_DIR); \
	else \
		echo "✅ googleapis already exists at $(GOOGLEAPIS_DIR)"; \
	fi

.PHONY: generate
generate: export PATH := $(LOCAL_BIN):$(PATH)
generate:
	@echo "⚙️ Generating Go code from proto files..."
	mkdir -p pb/go swagger
	    $(PROTOC) -I $(PROTO_DIR) -I $(GOOGLEAPIS_DIR) -I $(GRPC_GATEWAY_DIR) \
		--go_out=pb/go --go_opt=paths=source_relative \
		--go-grpc_out=pb/go --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb/go --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=swagger \
		$(PROTO_DIR)/telegram_service.proto
	@echo "✅ Code generation complete"

.PHONY: build
build:
	@echo "Building telegram service..."
	mkdir -p $(LOCAL_BIN)
	go build -o $(LOCAL_BIN)/telegram ./cmd/telegram_service/main.go
	@echo "Build complete: $(LOCAL_BIN)/telegram"

.PHONY: run
run: build
	@echo "Starting Telegram Service..."
	$(LOCAL_BIN)/telegram

.PHONY: lint install-lint

install-lint:
	@mkdir -p $(LOCAL_BIN)
	@if [ ! -f "$(GOLANGCI_LINT)" ]; then \
		curl -sSfL https://golangci-lint.run/install.sh | \
		sh -s -- -b $(LOCAL_BIN) v2.11.3; \
	fi

lint: install-lint
	$(GOLANGCI_LINT) run ./...

lint-fix: install-lint
	$(GOLANGCI_LINT) run --fix ./...


.PHONY: test test-cover test-verbose

test:
	go test ./...

test-cover:
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-race:
	go test -race ./...

.PHONY:  mock
mock:
	PATH="$(LOCAL_BIN):$(PATH)" go generate ./...

.PHONY: clean-mocks
clean-mocks: ## Очистка сгенерированных моков
	@echo "Cleaning mocks..."
	@find . -name "*_mock.go" -type f -delete
	@find . -name "mock_*.go" -type f -delete

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(LOCAL_BIN)
	rm -rf pb/go swagger
	@echo "✅ Clean complete"
