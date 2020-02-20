help:
	@echo "release      - build binaries using cross compliing for the support architectures"
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

internal/pkg/commands/%_pb.go: assets/protosets/%.pb
	@echo "/*" > $@
	@echo " * Copyright 2019-present Open Networking Foundation" >> $@
	@echo " *" >> $@
	@echo " * Licensed under the Apache License, Version 2.0 (the "License");" >> $@
	@echo " * you may not use this file except in compliance with the License." >> $@
	@echo " * You may obtain a copy of the License at" >> $@
	@echo " *" >> $@
	@echo " * http://www.apache.org/licenses/LICENSE-2.0" >> $@
	@echo " *" >> $@
	@echo " * Unless required by applicable law or agreed to in writing, software" >> $@
	@echo " * distributed under the License is distributed on an "AS IS" BASIS," >> $@
	@echo " * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied." >> $@
	@echo " * See the License for the specific language governing permissions and" >> $@
	@echo " * limitations under the License." >> $@
	@echo " */" >> $@
	@echo "package commands" >> $@
	@echo "" >> $@
	@echo "var $(shell echo $(subst .pb,,$(subst assets/protosets/voltha_,,$<)) |tr '[:lower:]' '[:upper:]')Descriptor = []byte{" >> $@
	hexdump -ve '1/1 "0x%02x,"' $< | fold -w 60 -s >> $@
	@echo "}" >> $@
	@go fmt $@

encode-protosets: internal/pkg/commands/voltha_v1_pb.go internal/pkg/commands/voltha_v2_pb.go internal/pkg/commands/voltha_v3_pb.go

SHELL=bash -e -o pipefail

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
VOLTHA_TOOLS_VERSION ?= 1.0.3

GO                = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang go
GO_SH             = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang sh -c
GO_JUNIT_REPORT   = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/appecho  -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-go-junit-report go-junit-report
GOCOVER_COBERTURA = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-gocover-cobertura gocover-cobertura
GOFMT             = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang gofmt
GOLANGCI_LINT     = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golangci-lint golangci-lint

release:
	@mkdir -p $(RELEASE_DIR)
	@${GO_SH} ' \
	  set -e -o pipefail; \
	  for x in ${RELEASE_OS_ARCH}; do \
	    OUT_PATH="$(RELEASE_DIR)/$(RELEASE_NAME)-$(subst -dev,_dev,$(VERSION))-$$x"; \
	    echo "$$OUT_PATH"; \
	    GOOS=$${x%-*} GOARCH=$${x#*-} go build -mod=vendor -v $(LDFLAGS) -o "$$OUT_PATH" cmd/voltctl/voltctl.go; \
	  done'

build:
	go build -mod=vendor $(LDFLAGS) cmd/voltctl/voltctl.go

install:
	go install -mod=vendor $(LDFLAGS) cmd/voltctl/voltctl.go

run:
	go run -mod=vendor $(LDFLAGS) cmd/voltctl/voltctl.go $(CMD)

lint-style:
	@echo "Running style check..."
	@gofmt_out="$$(${GOFMT} -l $$(find . -name '*.go' -not -path './vendor/*'))" ;\
	if [ ! -z "$$gofmt_out" ]; then \
	  echo "$$gofmt_out" ;\
	  echo "Style check failed on one or more files ^, run 'go fmt' to fix." ;\
	  exit 1 ;\
	fi
	@echo "Style check OK"

lint-sanity:
	@echo "Running sanity check..."
	@${GO} vet -mod=vendor ./...
	@echo "Sanity check OK"

lint-mod:
	@echo "Running dependency check..."
	@${GO} mod verify
	@echo "Dependency check OK. Running vendor check..."
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Staged or modified files must be committed before running this test" && echo "`git status`" && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files must be cleaned up before running this test" && echo "`git status`" && exit 1)
	${GO} mod tidy
	${GO} mod vendor
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Modified files detected after running go mod tidy / go mod vendor" && echo "`git status`" && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files detected after running go mod tidy / go mod vendor" && echo "`git status`" && exit 1)
	@echo "Vendor check OK."

lint: lint-style lint-sanity lint-mod

sca:
	@rm -rf ./sca-report
	@mkdir -p ./sca-report
	@echo "Running static code analysis..."
	@${GOLANGCI_LINT} run --deadline=20m -E golint --out-format junit-xml ./... | tee ./sca-report/sca-report.xml
	@echo ""
	@echo "Static code analysis OK"

test:
	@mkdir -p ./tests/results
	@${GO} test -mod=vendor -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	${GO_JUNIT_REPORT} < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	${GOCOVER_COBERTURA} < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml ;\
	exit $$RETURN

view-coverage:
	go tool cover -html ./tests/results/go-test-coverage.out

check: lint sca test

clean:
	rm -rf voltctl voltctl.cp release sca-report

mod-update:
	${GO} mod tidy
	${GO} mod vendor
