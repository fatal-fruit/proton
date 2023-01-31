#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build
LEDGER_ENABLED ?= true

# ********** Golang configs **********

export GO111MODULE = on

GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)

# ********** process build tags **********

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(PROTON_BUILD_OPTIONS)))
  build_tags += gcc cleveldb
else ifeq (rocksdb,$(findstring rocksdb,$(PROTON_BUILD_OPTIONS)))
  build_tags += gcc rocksdb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace := $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# ********** process linker flags **********

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=proton \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=protond \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq (cleveldb,$(findstring cleveldb,$(PROTON_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
else ifeq (rocksdb,$(findstring rocksdb,$(PROTONS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb
endif
ifeq (,$(findstring nostrip,$(PROTON_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags '$(build_tags)' -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(PROTON_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

all: proto-gen lint test install


###############################################################################
###                                  Build                                  ###
###############################################################################

install: enforce-go-version
	@echo "Installing protond..."
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/protond

build: enforce-go-version
	@echo "Building protond..."
	go build $(BUILD_FLAGS) -o $(BUILDDIR)/ ./cmd/protond

enforce-go-version:
	@echo "Go version: $(GO_MAJOR_VERSION).$(GO_MINOR_VERSION)"
ifneq ($(GO_MINOR_VERSION),18)
	@echo "Go version 1.18 is required"
	@exit 1
endif

clean:
	rm -rf $(CURDIR)/artifacts/

distclean: clean
	rm -rf vendor/


###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-gen:
	@echo "Generating Protobuf files"
	@sh ./proto/scripts/protocgen.sh

proto-doc:
	@echo "Generating Protoc docs"
	@sh ./proto/scripts/protoc-doc-gen.sh

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@sh ./proto/scripts/protoc-swagger-gen.sh


###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_lint_cmd=github.com/golangci/golangci-lint/cmd/golangci-lint

lint:
	@echo "--> Running linter"
	@go run $(golangci_lint_cmd) run --timeout=10m

format:
	@go run $(golangci_lint_cmd) run ./... --fix
	@go run mvdan.cc/gofumpt -l -w x/ app/ tests/


###############################################################################
###                                Localnet                                 ###
###############################################################################

start-localnet: build
	rm -rf ~/.protond-liveness
	./build/protond init liveness --chain-id liveness --staking-bond-denom uproton --home ~/.protond-liveness
	./build/protond config chain-id proton-1 --home ~/.protond-liveness
	./build/protond config keyring-backend test --home ~/.protond-liveness
	./build/protond keys add val --home ~/.protond-liveness
	./build/protond add-genesis-account val 10000000000000000000000000uproton --home ~/.protond-liveness --keyring-backend test
	./build/protond gentx val 1000000000uproton --home ~/.protond-liveness --chain-id proton-1
	./build/protond collect-gentxs --home ~/.protond-liveness
	sed -i.bak'' 's/minimum-gas-prices = ""/minimum-gas-prices = "0uproton"/' ~/.protond-liveness/config/app.toml
	./build/protond start --home ~/.protond-liveness --x-crisis-skip-assert-invariants


###############################################################################
###                           Tests & Simulation                            ###
###############################################################################
PACKAGES_UNIT=$(shell go list ./... | grep -v -e '/tests/e2e')
PACKAGES_E2E=$(shell cd tests/e2e && go list ./... | grep '/e2e')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-e2e

test-unit: ARGS=-timeout=5m -tags='norace'
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)
test-e2e: ARGS=-timeout=25m -v
test-e2e: TEST_PACKAGES=$(PACKAGES_E2E)
$(TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	@echo "--> Running tests"
	@go test -mod=readonly -json $(ARGS) $(TEST_PACKAGES) | tparse
else
	@echo "--> Running tests"
	@go test -mod=readonly $(ARGS) $(TEST_PACKAGES)
endif
