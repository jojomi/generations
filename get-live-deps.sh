#!/usr/bin/env sh
set -ex

#rsync -avz --delete --exclude=".git" --exclude="output" --exclude="pictures" "/home/johannes/go/src/github.com/jojomi/generations/" "/home/johannes/go/src/gitlab.com/jojomi/herkur/vendor/github.com/jojomi/generations/"
#rsync -avz --delete --exclude=".git" --exclude="output" --exclude="pictures" "/home/johannes/go/src/github.com/jojomi/generations/" "/home/johannes/go/pkg/mod/github.com/jojomi/generations@v0.0.0-20190817192325-72081379c89d/"

rsync -avz --delete --exclude=".git" --exclude="output" --exclude="pictures" "/home/johannes/go/src/github.com/jojomi/timecalc/" "/home/johannes/go/src/github.com/jojomi/generations/vendor/github.com/jojomi/timecalc/"
rsync -avz --delete --exclude=".git" --exclude="output" --exclude="pictures" "/home/johannes/go/src/github.com/jojomi/timecalc/" "/home/johannes/go/pkg/mod/github.com/jojomi/timecalc@v0.0.0-20190818131329-cd942dd0b67d/"