#!/usr/bin/env bash

set -e

BIN_DIR=$1
BASE_NAME=$2

echo "Preparing to sign..." 1>&2

HOSTNAME=downloads.${DOMAIN:-"spacelift.io"}
CHECKSUMS_FILE=${BIN_DIR}/${BASE_NAME}_SHA256SUMS
BINARY_NAME=${BIN_DIR}/${BASE_NAME}

SHASUM=$(openssl dgst -sha256 ${BINARY_NAME} | cut -d' ' -f2)

echo "${SHASUM}  ${BASE_NAME}" >> $CHECKSUMS_FILE

echo "Signing the checksums file..." 1>&2

gpg \
    --local-user $GPG_KEY_ID     \
    --output=$CHECKSUMS_FILE.sig \
    --passphrase=$GPG_PASSPHRASE \
    --pinentry-mode=loopback     \
    --detach-sig                 \
    $CHECKSUMS_FILE

