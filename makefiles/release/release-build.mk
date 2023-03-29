# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2019-2023 Open Networking Foundation
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

## -----------------------------------------------------------------------
## Intent:
##   o Generate volctl binaries in a docker container container
##   o Copy container:/apps/release to localhost:{pwd}/release
## -----------------------------------------------------------------------
## [TODO] Replace ${GO_SH} $(single-quote) ..   with $(call quoted,cmd-text)
## -----------------------------------------------------------------------
release-build :

	@echo 
	@echo "** -----------------------------------------------------------------------"
	@echo "** $(MAKE): processing target [$@]"
	@echo "** Sandbox: $(shell /bin/pwd)"
	@echo "** -----------------------------------------------------------------------"

	@echo -e "\n** golang attributes"
	$(HIDE)${GO_SH} $(call quoted,which$(space)-a$(space)go)
	$(HIDE)${GO_SH} $(call quoted,go$(space)version)

	@echo -e "\n** Create filesystem target for docker volume: $(RELEASE_DIR)"
	$(RM) -r "./$(RELEASE_DIR)"
	mkdir -vp "$(RELEASE_DIR)"

	@echo
	@echo '** Docker builds bins into mounted filesystem:'
	@echo '**   container:/app/relase'
	@echo '**   localhost:{pwd}/release'
	@${GO_SH} $(single-quote) \
	  set -e -o pipefail; \
	  for x in ${RELEASE_OS_ARCH}; do \
	    OUT_PATH="$(RELEASE_DIR)/$(RELEASE_NAME)-$(subst -dev,_dev,$(VERSION))-$$x"; \
	    echo "$$OUT_PATH"; \
	    GOOS=$${x%-*} GOARCH=$${x#*-} go build -mod=vendor -v $(LDFLAGS) -o "$$OUT_PATH" cmd/voltctl/voltctl.go; \
	done \
$(single-quote)

	@echo
	@echo '** Post-build, files to release'
	$(HIDE)find "$(RELEASE_DIR)" ! -type d -print

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
clean ::
	$(RM) -r "./$(RELEASE_DIR)"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
help ::
	@echo '  release-build       Cross-compile binaries into a docker mounted filesystem'

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
