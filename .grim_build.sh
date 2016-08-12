#!/bin/bash

set -eu

. /opt/golang/preferred/bin/go_env.sh

export GOPATH="$(pwd)/go"
export PATH="$GOPATH/bin:$PATH"

export SR_TEST_SCHEMA_REGISTRY=http://kafka-changes-qa.aws.infra.mediamath.com:8081
export SR_TEST_REQUIRED=true

export VERBOSE=yes

cd "./$CLONE_PATH"

make test
