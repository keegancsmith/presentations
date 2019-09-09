#!/bin/bash

set -e

cmd='
BEGIN {print "# Presentations\n"}
{ print "- https://talks.godoc.org/github.com/keegancsmith/presentations/" $(NF-1) "/" $(NF) }
'

find . -type f -maxdepth 2 -name '*.slide' \
    | sort -n \
    | awk -F '/' "$cmd" > README.md

git diff README.md

