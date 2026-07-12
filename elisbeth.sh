#!/bin/sh
set -ex

(cd /home/johannes/go/src/gitlab.com/jojomi/herkur && ./data-update.sh && ./export-generations.sh)
./get-live-deps.sh

# familybook
NAME=elisabeth ./familybook.sh -o -v=false