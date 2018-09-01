#!/bin/sh
set -ex

go generate ./...

COVERAGE_FILENAME=/tmp/go-test-coverage.out
go test -coverprofile="${COVERAGE_FILENAME}" -failfast "$@"
go tool cover -func="${COVERAGE_FILENAME}"
