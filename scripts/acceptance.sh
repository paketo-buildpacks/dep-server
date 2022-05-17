#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly DEPSERVERDIR="$(cd "${PROGDIR}/.." && pwd)"

go test -v "${DEPSERVERDIR}"/pkg/dependency/acceptance
