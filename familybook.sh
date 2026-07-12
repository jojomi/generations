#!/bin/sh
set -ex

BINARY_NAME=_local_bin_testing_generations
COMMAND_DIR=cmd/generations

function askYesNo {
        QUESTION=$1
        DEFAULT=$2
        if [ "$DEFAULT" = true ]; then
                OPTIONS="[Y/n]"
                DEFAULT="y"
            else
                OPTIONS="[y/N]"
                DEFAULT="n"
        fi
        read -p "$QUESTION $OPTIONS " -n 1 -s -r INPUT
        INPUT=${INPUT:-${DEFAULT}}
        echo ${INPUT}
        if [[ "$INPUT" =~ ^[yY]$ ]]; then
            ANSWER=true
        else
            ANSWER=false
        fi
}


# frische Daten
pushd /home/johannes/go/src/gitlab.com/jojomi/herkur > /dev/null
# docker-compose up -d
# ./sync-db.sh
# ./export-generations.sh --include-fiancees
popd > /dev/null


# build generations binary
rm -f "${BINARY_NAME}"
go build -v -o "${COMMAND_DIR}/${BINARY_NAME}" "./${COMMAND_DIR}/"


pushd ${COMMAND_DIR} > /dev/null

# build single PDF file
#GENERATIONS_DATABASE_BASE_DIR=/run/storage/encrypted/data/dokumente/Stammbaum/generations-data
GENERATIONS_DATABASE_BASE_DIR=~/.generations/databases
CONFIG_NAME=${NAME:-brigitte}-familybook
time "./${BINARY_NAME}" familybook \
-d "$GENERATIONS_DATABASE_BASE_DIR" -o=false -v=true -n=1 "$@" \
"config/${CONFIG_NAME}.yml"

# Optional: optimize PDF file size with pdfoptimize (go get github.com/jojomi/pdfoptimize)
# pdfoptimize "output/${CONFIG_NAME}.pdf" "output/${CONFIG_NAME}.opt.pdf"

# cleanup
rm -f "${BINARY_NAME}"


# upload
if [ "${NAME:-brigitte}" = "brigitte" -o "${NAME:-brigitte}" = "cousins" -o "${NAME:-brigitte}" = "cousins-partner" -o "${NAME:-brigitte}" = "cousins-gäste" -o "${NAME:-brigitte}" = "sippe-stahl" ]; then
    askYesNo "Upload PDF für Brigitte via data.slzm.de?" false
    DOIT=$ANSWER

    if [ "$DOIT" = true ]; then
        cp "output/${CONFIG_NAME}.pdf" "/home/johannes/Seafile/Stammbaum Brigitte/"
    fi
fi

if [ "${NAME:-brigitte}" = "brigitte" -o "${NAME:-brigitte}" = "cousins" -o "${NAME:-brigitte}" = "cousins-partner" -o "${NAME:-brigitte}" = "cousins-gäste" -o "${NAME:-brigitte}" = "sippe-stahl" ]; then
    askYesNo "Upload databases Sippe Stahl für Brigitte via data.slzm.de?" false
    DOIT=$ANSWER

    if [ "$DOIT" = true ]; then
        cp "${GENERATIONS_DATABASE_BASE_DIR}"/Sippe\ Stahl.yml ~/"Seafile/Stammbaum Brigitte/databases/"
        cp "${GENERATIONS_DATABASE_BASE_DIR}"/Elisabeth\ Schmid.yml ~/"Seafile/Stammbaum Brigitte/databases/"
    fi
fi


popd > /dev/null
