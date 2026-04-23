#!/usr/bin/env bash
# Copyright (c) 2026 Jodoo CLI Authors
# SPDX-License-Identifier: MIT

set -euo pipefail

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
DATE="$(date +%Y-%m-%d)"
LDFLAGS="-s -w -X jodoo-cli/internal/build.Version=${VERSION} -X jodoo-cli/internal/build.Date=${DATE}"

go build -trimpath -ldflags "${LDFLAGS}" -o jodoo-cli .
echo "OK: ./jodoo-cli (${VERSION})"
