#!/bin/sh
set -ex

cd ../cmd
go build -o examplebinary
./examplebinary "$@" ../example/config.yml
rm examplebinary

lualatex test.tex
lualatex test.tex

mv test.pdf ../example/example.pdf

cd ../example
# if you see this: attempt to perform an operation not allowed by the security policy `PDF' @ error/constitute.c/IsCoderAuthorized/408.
# https://stackoverflow.com/a/53180170/4021739
convert example.pdf example.png
xdg-open example.pdf
