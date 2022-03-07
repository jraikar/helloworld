# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

.PHONY: build
build: ## Build capi-api binary
	go build -o $(BIN_DIR)/capi-api github.com/aerospike/aerostation/capi-api
