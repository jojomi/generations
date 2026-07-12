#!/bin/sh
set -ex

pushd cmd > /dev/null
# build generations binary
go build -o generations

# build single PDF files
#GENERATIONS_DATABASE_BASE_DIR=/run/storage/encrypted/data/dokumente/Stammbaum/generations-data
GENERATIONS_DATABASE_BASE_DIR=~/.generations/databases
time ./generations genealogytree \
-d "$GENERATIONS_DATABASE_BASE_DIR" -o=false -v=true -n=1 \
config/stahl-enkel.yml

# cat output output/brigitte.pdf

# open result
ls -lahr output/stahl-enkel.pdf
xdg-open output/stahl-enkel.pdf

# remove binary
rm -f generations
popd > /dev/null
