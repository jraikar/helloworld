
COVER_DIR=./.cover
COVER_FILE="${COVER_DIR}/coverall.out"
COVER_HTML="${COVER_DIR}/coverall.html"
PACKAGES=$(shell go list ./... | grep -v '/vendor/')

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: manifests generate fmt vet ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.2/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

##@ Build

build: generate fmt vet aeroctl build-capi-api build-api-server build-user-service ## Build manager binary.
	go build -o bin/manager main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

docker-build: test ## Build docker image with the manager.
	docker build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	($(KUSTOMIZE) build config/crd | kubectl replace -f -) || ($(KUSTOMIZE) build config/crd | kubectl create -f -)
	#$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > app/controller.yaml #| kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

BIN_DIR := bin

.PHONY: aeroctl
aeroctl: ## Build aeroctl binary
	go build -o $(BIN_DIR)/aeroctl github.com/aerospike/aerostation/cmd/aeroctl

build-api-server: ## Build capi-api binary
	go build -o $(BIN_DIR)/api-server github.com/aerospike/aerostation/api-server

build-capi-api: ## Build capi-api binary
	go build -o $(BIN_DIR)/capi-api github.com/aerospike/aerostation/capi-api

build-user-service: ## Build user-api binary
	go build -o $(BIN_DIR)/user github.com/aerospike/aerostation/user-service


.PHONY: web-server
web-server: ## build the web UI components
	go build -o $(BIN_DIR)/aeroapi github.com/aerospike/aerostation/dashboard

.PHONY: capi-api
capi-api:
	go build -o $(BIN_DIR)/aero-api github.com/aerospike/aerostation/capi-api/
.PHONY: build-proto
build-proto:
	protoc --proto_path=capi-api/messages --go_out=capi-api/messages --go-grpc_out=capi-api/messages --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative service.proto aerospike.proto kubernetes.proto


.PHONY: user-service
user-service:
	go build -o $(BIN_DIR)/user github.com/aerospike/aerostation/user-service/

clean:
	go clean -i -x -r

cover-html:
	rm -rf ${COVER_DIR}
	mkdir ${COVER_DIR}
	echo "mode: count" > ${COVER_FILE}
	$(foreach pkg, ${PACKAGES},\
		go test -coverprofile=${COVER_DIR}/cover-pkg.pkg.out -covermode=count ${pkg};\
		grep -h -v "^mode:" ${COVER_DIR}/*.pkg.out >> ${COVER_FILE} || true;\
	)
	go tool cover -html=${COVER_FILE} -o ${COVER_HTML}

swagger-generate:
	swagger generate spec -o api-server/swaggerui/swagger.json -m

swagger-serve:
	swagger serve dashboard/swaggerui/swagger.json
