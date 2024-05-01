# Howto create a python 3.10+ patch

1) Checkout voltha-docs
2) cd voltha-docs
3) Create a virtual environment:
   - make venv (default python version)
   - make venv-activate-patched (for python v3.10+)
4) make patch-init
5) modify the file to be patched beneath staging/${relative_path_to_patch}
6) make patch-create PATCH_PATH=${relative_path_to_patch}
    o This will create patches/${relative_path_to_patch}/patch
    o make patch-create PATCH_PATH=lib/python3.10/site-packages/sphinx/util/typing.py
      lib/python3.10/site-packages/sphinx/util/typing/patch
7) Verify
    o make sterile
    o make venv
8) Validate
    o make lint
    o make test

<!---

# -----------------------------------------------------------------------
# Copyright 2023-2024 Open Networking Foundation Contributors
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
# SPDX-FileCopyrightText: 2023-2024 Open Networking Foundation Contributors
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------
# Intent:
# -----------------------------------------------------------------------

--->
