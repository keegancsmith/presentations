#!/bin/bash

set -e
convert $1 -resize ${2:-1000x550}\> new_$1
mv $1 bak_$1
mv new_$1 $1
