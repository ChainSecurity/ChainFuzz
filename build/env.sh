#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

if [ ! -L "$GOPATH/src/fuzzer/vendor" ]; then
    ln -s "$GOPATH/src/github.com/ethereum/go-ethereum/vendor" $GOPATH/src/fuzzer/vendor
fi

# Launch the arguments with the configured environment.
exec "$@"
# /bin/bash
