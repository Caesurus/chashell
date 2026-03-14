
SERVER_SOURCE=./cmd/server
CLIENT_SOURCE=./cmd/shell
LDFLAGS="-X main.targetDomain=$(DOMAIN_NAME) -X main.encryptionKey=$(ENCRYPTION_KEY) -s -w"
GCFLAGS="all=-trimpath=$$GOPATH"

CLIENT_BINARY=chashell
SERVER_BINARY=chaserv
TAGS=release

OSARCH ?= linux/amd64 linux/386 linux/arm windows/amd64 windows/386 darwin/amd64 darwin/arm64

.DEFAULT: help

help: ## Show Help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

check-env: ## Check if necessary environment variables are set.
ifndef DOMAIN_NAME
	$(error DOMAIN_NAME is undefined)
endif
ifndef ENCRYPTION_KEY
	$(error ENCRYPTION_KEY is undefined)
endif

build: check-env ## Build for the current architecture.
	mkdir -p release && \
	go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) -tags $(TAGS) -o release/$(CLIENT_BINARY) $(CLIENT_SOURCE) && \
	go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) -tags $(TAGS) -o release/$(SERVER_BINARY) $(SERVER_SOURCE)

deps: ## Download Go module dependencies
	go mod download

dep: deps ## Backwards-compatible alias

build-client: check-env ## Build the chashell client.
	@echo "Building shell"
	mkdir -p release && \
	for osarch in $(OSARCH); do \
		os=$${osarch%/*}; arch=$${osarch#*/}; \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		echo "  -> $$os/$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) -tags $(TAGS) -o "release/chashell_$${os}_$${arch}$${ext}" $(CLIENT_SOURCE) || exit $$?; \
	done

build-server: check-env ## Build the chashell server.
	@echo "Building server"
	mkdir -p release && \
	for osarch in $(OSARCH); do \
		os=$${osarch%/*}; arch=$${osarch#*/}; \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		echo "  -> $$os/$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) -tags $(TAGS) -o "release/chaserv_$${os}_$${arch}$${ext}" $(SERVER_SOURCE) || exit $$?; \
	done


build-all: check-env build-client build-server ## Build everything.

proto: ## Build the protocol buffer file
	protoc -I=proto/ --go_out=lib/protocol chacomm.proto

clean: ## Remove all the generated binaries
	rm -f release/chaserv*
	rm -f release/chashell*
