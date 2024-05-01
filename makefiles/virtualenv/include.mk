# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2017-2024 Open Networking Foundation Contributors
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
# SPDX-FileCopyrightText: 2017-2024 Open Networking Foundation Contributors
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------
# -----------------------------------------------------------------------
# Intent:
#   This makefile defines dependencies that will install a python virtualenv
#   beneath $(sandbox)/.venv/.  The $(activate) macro is used to source
#   .venv/bin/activate allowing command python and pip to be used.
# -----------------------------------------------------------------------
# Usage:
#   include makefiles/virtualenv/include.mk
#
# Makefile Target Dependencies:
#     tgt : $(venv-activate-patched)      # python 3.10+ local use
#     tgt : $(venv-activate-script)       # python < v3.8
#
# Make definitions (convenience macros used for command access)
#   PIP    := $(activate) && pip          # invoke pip from virtualenv
#   PYTHON := $(activate) && python       # invoke python from virtualenv
#
# Target declaration and command invocation:
#   my-target : $(venv-activate-script)   # dependency installs virtualenv
#   <tab>$(PYTHON) --version              # invoke python with arguments
#	<tab>$(PYTHON) my-command.py
#	<tab>$(activate) && pip install foobar
#
#   % make my-target                      # Invoke make target from shell
# -----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

##-------------------##
##---]  GLOBALS  [---##
##-------------------##
.PHONY: venv venv-patched

##------------------##
##---]  LOCALS  [---##
##------------------##
venv-name            ?= .venv#                            # default install directory
venv-abs-path        := $(sandbox-root)/$(venv-name)#     # Install directory
venv-activate-bin    := $(venv-name)/bin#                 # no whitespace
venv-activate-script := $(venv-activate-bin)/activate#    # dependency

##--------------------##
##---]  INCLUDES  [---##
##--------------------##
include $(onf-mk-dir)/virtualenv/requirements-txt.mk
include $(onf-mk-dir)/virtualenv/version.mk

# ------------------------------------------------------------------------
# Intent: Define macro activate= to access virtualenv activation script.
## -----------------------------------------------------------------------
#  Usage:
#    - $(activate) && python             # Syntax inlined within a target
#    - PYTHON := $(activate) && python   # Define a named command macro
# ------------------------------------------------------------------------
activate             ?= set +u && source $(venv-activate-script) && set -u

## -----------------------------------------------------------------------
## Intent: Explicit named installer target w/o dependencies.
##         Makefile targets should depend on venv-activate-script.
## -----------------------------------------------------------------------
venv := $(null)
venv += $(venv-activate-script)#        # virtualenv -p python3
venv += $(venv-requirements-txt)#       # pip install -r requirements.txt

venv: $(venv)

venv-patched : $(venv-activate-patched)

## -----------------------------------------------------------------------
## Intent: Activate script path dependency
## Usage:
##    o place on the right side of colon as a target dependency
##    o When the script does not exist install the virtual env and display.
## -----------------------------------------------------------------------
$(venv-activate-script):

	$(call banner-enter,(virtualenv -p python))

	virtualenv -p python3 $(venv-name)
	$(activate) && python -m pip install --upgrade pip
	$(activate) && pip install --upgrade setuptools

	$(HIDE)$(MAKE) --no-print-directory venv-requirements venv-version

	$(call banner-leave,(virtualenv -t python))

## -----------------------------------------------------------------------
## Intent: Explicit named installer target w/o dependencies.
##         Makefile targets should depend on venv-activate-script.
## -----------------------------------------------------------------------
venv-activate-patched := $(venv-activate-script).patched
venv-activate-patched : $(venv-activate-patched)
$(venv-activate-patched) : $(venv-activate-script)
	$(call banner-enter,Target $@)
	$(onf-mk-dir)/virtualenv/python_310_migration.sh
	touch $@
	$(call banner-leave,Target $@)

## -----------------------------------------------------------------------
## Intent: Explicit named installer target w/o dependencies.
##         Makefile targets should depend on venv-activate-script.
## -----------------------------------------------------------------------
# venv         : $(venv-activate-script)

## -----------------------------------------------------------------------
## Intent: Revert installation to a clean checkout
## -----------------------------------------------------------------------
sterile :: clean
	$(RM) -r "$(venv-abs-path)"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
help ::
	@printf '  %-33.33s %s\n' 'venv'         \
	  'Create a python virtual environment'
	@printf '  %-33.33s %s\n' 'venv-help'    \
	  'Extended target help'

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: venv-help
venv-help ::
	@printf '  %-33.33s %s\n' 'venv-patched' \
	  'venv patched for v3.10.6+ use'

	@printf '  %-33.33s %s\n' 'venv' \
	  'Create a python virtual environment'
	@printf '  %-33.33s %s\n' '  venv-name' \
	  'virtualenv installation directory name'

# include $(onf-mk-dir)/virtualenv/todo.mk

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
