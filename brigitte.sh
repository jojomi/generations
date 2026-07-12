#!/bin/sh
set -ex

# source: https://stackoverflow.com/a/31939275/4021739
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


pushd cmd > /dev/null

# build generations binary
go build -o generations

# build single PDF files
#GENERATIONS_DATABASE_BASE_DIR=/run/storage/encrypted/data/dokumente/Stammbaum/generations-data
GENERATIONS_DATABASE_BASE_DIR=~/.generations/databases
time ./generations genealogytree \
-d "$GENERATIONS_DATABASE_BASE_DIR" -o=false -v=false -n=1 "$@" \
config/brigitte-familie-josef-stahl.yml \
config/brigitte-familie-karl-stahl.yml \
config/brigitte-familie-bock.yml \
config/brigitte-familie-frank.yml \
config/brigitte-familie-martin-stahl.yml \
config/brigitte-familie-anton-stahl.yml \
config/brigitte-familie-johann-stahl.yml \
config/brigitte-familie-dick.yml \
config/brigitte-familie-mark.yml \
config/brigitte-familie-alois-stahl.yml \
config/brigitte-familie-kilb.yml \
config/brigitte-familie-josef-franziskus-stahl.yml \
config/brigitte-familie-augustin-stahl.yml \
config/brigitte-familie-ohrnberger.yml \
config/brigitte-familie-symbol.yml

#config/brigitte-familie-augustin-stahl.yml




# concat PDF files
pdftk \
output/brigitte-familie-josef-stahl.pdf \
output/brigitte-familie-karl-stahl.pdf \
output/brigitte-familie-bock.pdf \
output/brigitte-familie-frank.pdf \
output/brigitte-familie-martin-stahl.pdf \
output/brigitte-familie-anton-stahl.pdf \
output/brigitte-familie-johann-stahl.pdf \
output/brigitte-familie-dick.pdf \
output/brigitte-familie-mark.pdf \
output/brigitte-familie-alois-stahl.pdf \
output/brigitte-familie-kilb.pdf \
output/brigitte-familie-josef-franziskus-stahl.pdf \
output/brigitte-familie-augustin-stahl.pdf \
output/brigitte-familie-ohrnberger.pdf \
output/brigitte-familie-symbol.pdf \
cat output output/brigitte.pdf

# output/brigitte-familie-kilb.pdf \
# cat output output/brigitte.pdf



# open result
ls -lahr output/brigitte.pdf
xdg-open output/brigitte.pdf


# Jahreskalender
popd > /dev/null
NAME=cousins-partner-jubi ./familybook.sh -v=false
pushd cmd > /dev/null

# zip it
ZIP_FILENAME=output/brigitte-stammbäume-$(date +"%Y-%m-%d").zip
zip "${ZIP_FILENAME}" --junk-paths output/brigitte.pdf \
    output/brigitte-familie-josef-stahl.pdf \
    output/brigitte-familie-karl-stahl.pdf \
    output/brigitte-familie-bock.pdf \
    output/brigitte-familie-frank.pdf \
    output/brigitte-familie-martin-stahl.pdf \
    output/brigitte-familie-anton-stahl.pdf \
    output/brigitte-familie-johann-stahl.pdf \
    output/brigitte-familie-dick.pdf \
    output/brigitte-familie-mark.pdf \
    output/brigitte-familie-alois-stahl.pdf \
    output/brigitte-familie-kilb.pdf \
    output/brigitte-familie-josef-franziskus-stahl.pdf \
    output/brigitte-familie-augustin-stahl.pdf \
    output/brigitte-familie-ohrnberger.pdf \
    output/brigitte-familie-symbol.pdf \
    config/brigitte-familie-josef-stahl.yml \
    config/brigitte-familie-karl-stahl.yml \
    config/brigitte-familie-bock.yml \
    config/brigitte-familie-frank.yml \
    config/brigitte-familie-martin-stahl.yml \
    config/brigitte-familie-anton-stahl.yml \
    config/brigitte-familie-johann-stahl.yml \
    config/brigitte-familie-dick.yml \
    config/brigitte-familie-mark.yml \
    config/brigitte-familie-alois-stahl.yml \
    config/brigitte-familie-kilb.yml \
    config/brigitte-familie-josef-franziskus-stahl.yml \
    config/brigitte-familie-augustin-stahl.yml \
    config/brigitte-familie-ohrnberger.yml \
    config/brigitte-familie-symbol.yml \
    config/brigitte-familie-josef-stahl.yml \
    config/brigitte-familie-karl-stahl.yml \
    config/brigitte-familie-bock.yml \
    config/brigitte-familie-frank.yml \
    config/brigitte-familie-martin-stahl.yml \
    config/brigitte-familie-anton-stahl.yml \
    config/brigitte-familie-johann-stahl.yml \
    config/brigitte-familie-dick.yml \
    config/brigitte-familie-mark.yml \
    config/brigitte-familie-alois-stahl.yml \
    config/brigitte-familie-kilb.yml \
    config/brigitte-familie-josef-franziskus-stahl.yml \
    config/brigitte-familie-augustin-stahl.yml \
    config/brigitte-familie-ohrnberger.yml \
    config/brigitte-familie-symbol.yml \
    config/cousins-partner-jubi-familybook.pdf ~/.generations/databases/Familie\ *.yml
## small: zip "${ZIP_FILENAME}" --junk-paths output/brigitte.pdf output/brigitte-familie-*.small.pdf ~/.generations/databases/Familie\ *.yml config/brigitte-familie-*.yml
#zip --delete herkur.yml "${ZIP_FILENAME}"

# upload
askYesNo "Upload für Brigitte via data.slzm.de?" true
DOIT=$ANSWER

if [ "$DOIT" = true ]; then
    cp "${ZIP_FILENAME}" "/home/johannes/Seafile/Stammbaum Brigitte/"
fi

# remove binary
rm -f generations
popd > /dev/null
