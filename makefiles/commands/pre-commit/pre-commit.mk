# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2022-2024 Open Networking Foundation (ONF) Contributors
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
# SPDX-FileCopyrightText: 2022-2024 Open Networking Foundation Contributors
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

## -----------------------------------------------------------------------
## Intent: Invoke the pre-commit command
## -----------------------------------------------------------------------
.PHONY: pre-commit
pre-commit : venv
	$(activate) && pre-commit

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
help ::
	@printf '  %-33.33s %s\n' 'pre-commit' \
	  'Invoke command pre-commit'
	@printf '  %-33.33s %s\n' 'pre-commit-help' \
	  'Display extended target help (pre-commit-*)'

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
tox-help ::
	@printf '  %-33.33s %s\n' 'tox-run' \
	  'Self documenting alias for command tox'

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
