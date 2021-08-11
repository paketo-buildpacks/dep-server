#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly DEPSERVERDIR="$(cd "${PROGDIR}/.." && pwd)"

# Make build dir
mkdir -p "${DEPSERVERDIR}/build/module"
# Remove previous artifact
if [[ -f "${DEPSERVERDIR}/build/cyclonedx.tgz" ]]; then
  rm "${DEPSERVERDIR}/build/*.tgz"
fi
# NPM install in build dir
pushd "${DEPSERVERDIR}/build/module" > /dev/null || return
  printf "%s\n" "Installing @cyclonedx/bom"

  npm install --quiet --no-progress --global-style @cyclonedx/bom  &> /dev/null
popd > /dev/null || return
# Tar up module
pushd "${DEPSERVERDIR}/build/module/node_modules/@cyclonedx/bom/" > /dev/null || return
  printf "%s\n" "Tarring up module"

  tar -czf "${DEPSERVERDIR}/build/artifact.tgz" ./*

  printf "%s\n" "Renaming tar file"
  version=$(jq -r '.version' < package.json)
  sha=$(sha256sum  "${DEPSERVERDIR}/build/artifact.tgz" | cut -c1-8)

  mv "${DEPSERVERDIR}/build/artifact.tgz" "${DEPSERVERDIR}/build/cyclonedx-node-module_${version}_linux_x64_bionic_${sha}.tgz"
popd > /dev/null || return
# Dump everything else
printf "%s\n" "Removing build artifacts"
rm -rf "${DEPSERVERDIR}/build/module"
