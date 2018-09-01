#!/bin/sh
set -ex

cd ../cmd
go build -o examplebinary
./examplebinary --config-file ../example/config.yml
rm examplebinary

lualatex test.tex
lualatex test.tex

mv test.pdf ../example/example.pdf

cd ../example
convert example.pdf example.png
xdg-open example.pdf