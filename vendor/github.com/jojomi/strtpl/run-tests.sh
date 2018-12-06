#!/bin/sh

set -e

COVERAGEPATH=/tmp/coverage.out

go test -bench=. -benchmem -coverprofile="$COVERAGEPATH"
go tool cover -func="$COVERAGEPATH"
go tool cover -html="$COVERAGEPATH"
rm "$COVERAGEPATH"
