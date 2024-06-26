# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2017-2024 Open Networking Foundation (ONF) and the ONF Contributors
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
#
# SPDX-FileCopyrightText: 2024 Open Networking Foundation (ONF) and the ONF Contributors
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

MAKEDIR     ?= $(error MAKEDIR= is required)
ONF_MAKEDIR ?= $(MAKEDIR)

# Helpers -- eventually defined in lf/transition.mk
onf-mk-dir   := $(MAKEDIR)
sandbox-root ?= $(TOP)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
help::
	@echo "USAGE: $(MAKE) [options] [target] ..."
        # @echo "  test                          Sanity check chart versions"

include $(MAKEDIR)/consts.mk
include $(ONF_MAKEDIR)/utils/include.mk      # dependency-less helper macros
include $(ONF_MAKEDIR)/etc/include.mk        # banner macros
include $(ONF_MAKEDIR)/virtualenv/include.mk#  # python, lint, JJB dependency
include $(MAKEDIR)/todo.mk
include $(MAKEDIR)/lint/include.mk

include $(MAKEDIR)/commands/pre-commit/include.mk

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
