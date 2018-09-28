#!/bin/sh

set -ex

find dist -type f -name 'generations*' -exec upx -9 {} \;