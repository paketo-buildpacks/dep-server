#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly DEPSERVERDIR="$(cd "${PROGDIR}/.." && pwd)"

# Make build dir
mkdir -p "${DEPSERVERDIR}/build/module"

# Remove previous artifact
rm -rf "${DEPSERVERDIR}"/build/*.tgz

pushd "${DEPSERVERDIR}/build/module" > /dev/null || return
  # Get the asset URL of the tar.gz and the bom.json files
  bom_json_url="$(curl -s https://api.github.com/repos/CycloneDX/cyclonedx-gomod/releases/latest | jq -r ".assets[] | select(.name | contains(\"linux_amd64.bom.json\")) | .browser_download_url")"
  bom_tar_url="$(curl -s https://api.github.com/repos/CycloneDX/cyclonedx-gomod/releases/latest | jq -r ".assets[] | select(.name | contains(\"linux_amd64.tar.gz\")) | .browser_download_url")"

  # Download files
  printf "%s\n" "Downloading assets from github"
  curl -s -L -o cyclonedx-gomod.bom.json "$bom_json_url"
  curl -s -L -o cyclonedx-gomod.tar.gz "$bom_tar_url"

  # Untar provided tar file
  mkdir -p cyclonedx-gomod
  tar -xzf cyclonedx-gomod.tar.gz -C cyclonedx-gomod/

  # Modify file structure
  pushd "${DEPSERVERDIR}/build/module/cyclonedx-gomod/" > /dev/null || return
    printf "%s\n" "Modifying file structure"
    mkdir -p bin
    mv cyclonedx-gomod bin/cyclonedx-gomod

    # Tar up module
    printf "%s\n" "Tarring up module"
    tar -czf "${DEPSERVERDIR}/build/artifact.tgz" ./*
  popd > /dev/null || return

  # Rename tar file
  printf "%s\n" "Renaming tar file"
  version="$(jq -r ".metadata.tools[] | select(.name=\"cyclonedx-gomod\") | .version" < cyclonedx-gomod.bom.json)"
  sha=$(sha256sum  "${DEPSERVERDIR}/build/artifact.tgz" | cut -c1-8)

  mv "${DEPSERVERDIR}/build/artifact.tgz" "${DEPSERVERDIR}/build/cyclonedx-gomod_${version}_linux_x64_bionic_${sha}.tgz"
popd > /dev/null || return

# Dump everything else
printf "%s\n" "Removing build artifacts"
rm -rf "${DEPSERVERDIR}/build/module"