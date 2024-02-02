# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2019-2024 Open Networking Foundation (ONF) and the ONF Contributors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# -----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

.DEFAULT_GOAL := help

TOP         ?= .
MAKEDIR     ?= $(TOP)/makefiles

quoted = $(quote-single)$(1)$(quote-single)

$(if $(VERBOSE),$(eval export VERBOSE=$(VERBOSE))) # visible to include(s)

##--------------------##
##---]  INCLUDES  [---##
##--------------------##
include $(MAKEDIR)/include.mk
include $(MAKEDIR)/release/include.mk

ifdef LOCAL_LINT
  include $(MAKEDIR)/lint/golang/sca.mk
endif

## Are lint-style and lint-sanity targets defined in docker ?
help ::
	@echo
	@echo "build        - build the binary as a local executable"
	@echo "install      - build and install the binary into \$$GOPATH/bin"
	@echo "run          - runs voltctl using the command specified as \$$CMD"
	@echo "lint-style   - Verify code is properly gofmt-ed"
	@echo "lint-sanity  - Verify that 'go vet' doesn't report any issues"
	@echo "lint-mod     - Verify the integrity of the 'mod' files"
	@echo "lint         - run static code analysis"
	@echo "sca          - Runs various SCA through golangci-lint tool"
	@echo "test         - run unity tests"
	@echo "check        - runs targets that should be run before a commit"
	@echo "clean        - remove temporary and generated files"

# SHELL=bash -e -o pipefail

VERSION=$(shell cat ./VERSION)
GITCOMMIT=$(shell git rev-parse HEAD)
ifeq ($(shell git ls-files --others --modified --exclude-standard 2>/dev/null | wc -l | sed -e 's/ //g'),0)
GITDIRTY=false
else
GITDIRTY=true
endif
GOVERSION=$(shell go version 2>&1 | sed -E  's/.*(go[0-9]+\.[0-9]+\.[0-9]+).*/\1/g')
HOST_OS=$(shell uname -s | tr A-Z a-z)
ifeq ($(shell uname -m),x86_64)
	HOST_ARCH ?= amd64
else
	HOST_ARCH ?= $(shell uname -m)
endif
BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags \
	"-X \"github.com/opencord/voltctl/internal/pkg/cli/version.Version=$(VERSION)\"  \
	 -X \"github.com/opencord/voltctl/internal/pkg/cli/version.VcsRef=$(GITCOMMIT)\"  \
	 -X \"github.com/opencord/voltctl/internal/pkg/cli/version.VcsDirty=$(GITDIRTY)\"  \
	 -X \"github.com/opencord/voltctl/internal/pkg/cli/version.GoVersion=$(GOVERSION)\"  \
	 -X \"github.com/opencord/voltctl/internal/pkg/cli/version.Os=$(HOST_OS)\" \
	 -X \"github.com/opencord/voltctl/internal/pkg/cli/version.Arch=$(HOST_ARCH)\" \
	 -X \"github.com/opencord/voltctl/internal/pkg/cli/version.BuildTime=$(BUILDTIME)\""

# Release related items
# Generates binaries in $RELEASE_DIR with name $RELEASE_NAME-$RELEASE_OS_ARCH
# Inspired by: https://github.com/kubernetes/minikube/releases
RELEASE_DIR     ?= release
RELEASE_NAME    ?= voltctl
RELEASE_OS_ARCH ?= linux-amd64 linux-arm64 windows-amd64 darwin-amd64

# tool containers
VOLTHA_TOOLS_VERSION ?= 2.4.0

docker-run = docker run --rm --user $$(id -u):$$(id -g)#     # Docker command stem
docker-run-app = $(docker-run) -v ${CURDIR}:/app#            # w/filesystem mount

GO                = $(docker-run-app) $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang go
GO_SH             = $(docker-run-app) $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang sh -c
GO_JUNIT_REPORT   = $(docker-run) -v ${CURDIR}:/appecho  -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-go-junit-report go-junit-report
GOCOVER_COBERTURA = $(docker-run-app) -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-gocover-cobertura gocover-cobertura
GOLANGCI_LINT     = $(docker-run-app) $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg -e GOGC=10 voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golangci-lint golangci-lint

## -----------------------------------------------------------------------
## Why is docker an implicit dependency for "make lint" (?)
##   o A fixed version is required for jenkins build/release jobs.
##   o Devs should have the option of using whatever is available
##     including bleeding edge software and tool upgrades w/o overhead.
## -----------------------------------------------------------------------
## Usage:
##   % export LOCAL_DEV_MODE=1
##   % make lint
##   % make check
## -----------------------------------------------------------------------
ifdef LOCAL_DEV_MODE
  GO            := $(clean-env) go
  GOLANGCI_LINT := golangci-lint
endif

## -----------------------------------------------------------------------
## Intent: Cross-compile binaries for release
## -----------------------------------------------------------------------
release: release-build

## -----------------------------------------------------------------------
## Local Development Helpers
## -----------------------------------------------------------------------
local-lib-go:
ifdef LOCAL_LIB_GO
	$(RM) -r vendor/github.com/opencord/voltha-lib-go/v7/pkg
	mkdir -p vendor/github.com/opencord/voltha-lib-go/v7/pkg
	cp -r ${LOCAL_LIB_GO}/pkg/* vendor/github.com/opencord/voltha-lib-go/v7/pkg/
endif

## -----------------------------------------------------------------------
## Itent:
## -----------------------------------------------------------------------
build: local-lib-go
	go build -mod=vendor $(LDFLAGS) cmd/voltctl/voltctl.go

## -----------------------------------------------------------------------
## Itent:
## -----------------------------------------------------------------------
install:
	go install -mod=vendor $(LDFLAGS) cmd/voltctl/voltctl.go

## -----------------------------------------------------------------------
## Itent:
## -----------------------------------------------------------------------
run:
	go run -mod=vendor $(LDFLAGS) cmd/voltctl/voltctl.go $(CMD)

## -----------------------------------------------------------------------
## Itent:
## -----------------------------------------------------------------------
lint-mod:
	@echo "Running dependency check..."
	@$(GO) mod verify
	@echo "Dependency check OK. Running vendor check..."
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Staged or modified files must be committed before running this test" && git status -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files must be cleaned up before running this test" && git status -- go.mod go.sum vendor && exit 1)

	$(HIDE)$(MAKE) --no-print-directory mod-update

	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Modified files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@echo "Vendor check OK."

ifndef LOCAL_LINT
  lint : lint-mod
endif

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
# lint: lint-mod lint-dockerfile ## Run all lint targets

## -----------------------------------------------------------------------
##   Intent: Syntax check golang source
## See Also: makefilles/lint/golang/sca.mk
## -----------------------------------------------------------------------
sca:
	@$(RM) -r ./sca-report
	@mkdir -p ./sca-report
	@echo "Running static code analysis..."
	@${GOLANGCI_LINT} run --deadline=20m --out-format junit-xml ./... \
	    | tee ./sca-report/sca-report.xml
	@echo ""
	@echo "Static code analysis OK"

## -----------------------------------------------------------------------
## Intent: Evaluate test targets (docker required)
## -----------------------------------------------------------------------
test:
	@mkdir -p ./tests/results
	@$(GO) test -mod=vendor -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	${GO_JUNIT_REPORT} < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	${GOCOVER_COBERTURA} < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml ;\
	exit $$RETURN

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
view-coverage:
	go tool cover -html ./tests/results/go-test-coverage.out

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
check: lint sca test

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: mod-update
mod-update: go-version mod-tidy mod-vendor

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: go-version
go-version :
	$(call banner-enter,Target $@)
	${GO} version
	$(call banner-leave,Target $@)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: mod-tidy
mod-tidy :
	$(call banner-enter,Target $@)
	${GO} mod tidy
	$(call banner-leave,Target $@)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: mod-vendor
mod-vendor : mod-tidy
mod-vendor :
	$(call banner-enter,Target $@)
	$(if $(LOCAL_FIX_PERMS),chmod o+w $(CURDIR))
	${GO} mod vendor
	$(if $(LOCAL_FIX_PERMS),chmod o-w $(CURDIR))
	$(call banner-leave,Target $@)

## ---------------------------------------------------------
## ---------------------------------------------------------
clean ::
	$(RM) -r voltctl voltctl.cp sca-report

## ---------------------------------------------------------
## This belongs in a library makefile: makefiles/go/clean.mk
## ---------------------------------------------------------
go-clean-cache += -cache
# go-clean-cache += -fuzzcache
go-clean-cache += -modcache
go-clean-cache += -testcache

go-clean-args += -i # installed binaries
go-clean-args += -r # recursive
go-clean-args += -x # verbose removal

sterile :: clean
	$(GO) clean $(go-clean-cache)
	$(GO) clean $(go-clean-args)

## [SEE ALSO]
## -----------------------------------------------------------------------
##   o https://dave.cheney.net/tag/gogc
## -----------------------------------------------------------------------

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
