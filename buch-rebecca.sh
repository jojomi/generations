#!/bin/bash
set -ex


# generate PDF
time NAME=rebecca ./familybook.sh -o -n=2

# cut to pieces for epubli.de
cd cmd/output
pdftk rebecca-familybook.pdf cat 1 output rebecca-titel.pdf
pdftk rebecca-familybook.pdf cat 3-end output rebecca-innen.pdf