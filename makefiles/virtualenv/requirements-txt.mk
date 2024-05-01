# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2024 Open Networking Foundation Contributors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# -----------------------------------------------------------------------
# SPDX-FileCopyrightText: 2024 Open Networking Foundation Contributors
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------
## Intent: pip install with dependencies
## ----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

##-------------------##
##---]  GLOBALS  [---##
##-------------------##
venv-requirements := $(venv-abs-path)/makedep/requirements.txt.ts

## -----------------------------------------------------------------------
## Intent: Define a makefile target able to install venv python modules
##   when changes are made within the requirements.txt file.
## -----------------------------------------------------------------------
## [MAKEFILE TARGETS]
## venv-requirements
##   Named target used to abstract underlying dependency filename.
## $(venv-requirements-txt)
##   Make macro used to detect changes in requirements.txt file.
##   Timestamp filename is also the primary target for invoking pip install.
## -----------------------------------------------------------------------
.PHONY: venv-requirements
venv-requirements    : $(venv-requirements)
$(venv-requirements) : requirements.txt

	$(call banner-enter,venv-requirements)
	@mkdir -p $(dir $@)
	$(activate) && python -m pip install -r 'requirements.txt'
	@touch $@

	$(call banner-leave,venv-requirements)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
venv-help ::
	@printf '  %-33.33s %s\n' 'venv-requirements' \
	  'pip install -r requirements.txt (dependency driven)'

# [EOF]
