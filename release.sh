#!/bin/sh
set -ex

cd cmd

# ~/.config/goreleaser/github_token must contain your github token
goreleaser --rm-dist
# for testing
##goreleaser --snapshot --rm-dist --skip-publish --skip-validate