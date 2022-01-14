#!/usr/bin/env bash

set -e

VERSION="1.0.0"

BIN_DIR=$1
BASE_NAME=$2

echo "Preparing to verifying..." 1>&2

OS=linux
ARCH=amd64
CHECKSUMS_FILE=${BIN_DIR}/${BASE_NAME}_SHA256SUMS

CHECKSUM=$(cut -f 1 -d ' ' ${CHECKSUMS_FILE})
RELEASE_SHA=$(shasum -a 256 ${BIN_DIR}/${BASE_NAME} | cut -f 1 -d ' ')

echo "Verifying... ${BIN_DIR}/${BASE_NAME}" 1>&2

if [[ "$CHECKSUM" == "$RELEASE_SHA" ]]; then
  echo "Checksum and release hash are equal"
else
  echo "Checksum and release hash are unequal"
  exit 1
fi
