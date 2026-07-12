#!/usr/bin/env bash
set -e

NAME=sippe-stahl ./familybook.sh --verbose=false -n=2 --open=false "$@"
NAME=brigitte ./familybook.sh --verbose=false -n=2 --open=false "$@"
# NAME=cousins ./familybook.sh --verbose=false -n=2 --open=false "$@"
# NAME=cousins-partner ./familybook.sh --verbose=false -n=2 --open=false "$@"
# NAME=cousins-gäste ./familybook.sh --verbose=false -n=2 --open=false "$@"

# zip it
pushd cmd > /dev/null
ZIP_FILENAME=output/cousins-familybooks-$(date +"%Y-%d-%m").zip
zip "${ZIP_FILENAME}" --junk-paths output/cousins*familybook.pdf config/cousins*familybook.yml ~/.generations/databases/Familie\ *.yml
popd > /dev/null
