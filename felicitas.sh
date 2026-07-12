#!/bin/sh
set -ex


# frische Daten
pushd /home/johannes/go/src/gitlab.com/jojomi/herkur > /dev/null
# docker-compose up -d
# ./sync-db.sh
# ./export-generations.sh --include-fiancees
popd > /dev/null


pushd cmd > /dev/null

# build generations binary
go build -o generations

# build single PDF file
#GENERATIONS_DATABASE_BASE_DIR=/run/storage/encrypted/data/dokumente/Stammbaum/generations-data
GENERATIONS_DATABASE_BASE_DIR=~/.generations/databases
##COMMAND=familybook
##CONFIG_NAME=felicitas-familybook
COMMAND=genealogytree
CONFIG_NAME=felicitas
#CONFIG_NAME=mechthild-vorfahren
time ./generations "${COMMAND}" \
-d "$GENERATIONS_DATABASE_BASE_DIR" -o=false -v=true -n=1 "$@" \
"config/${CONFIG_NAME}.yml"

# open result
ls -lahr "output/${CONFIG_NAME}.pdf"
xdg-open "output/${CONFIG_NAME}.pdf"

# remove binary
rm -f generations
popd > /dev/null
