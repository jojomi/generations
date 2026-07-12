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
config/elisabeth-ahnen.yml

# open result
ls -lahr output/elisabeth-ahnen.pdf
xdg-open output/elisabeth-ahnen.pdf

# remove binary
rm -f generations
popd > /dev/null
