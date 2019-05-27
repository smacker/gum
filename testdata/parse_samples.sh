#!/bin/sh

set -e

BIN=/Users/smacker/Downloads/gumtree-2.1.2/bin/gumtree

mkdir -p parsed
for f in `find ./samples -type f -name "*.java" -o -name "*.py" -o -name "*.rb"`; do
    mkdir -p `dirname ./parsed/$f`
    $BIN parse $f > ./parsed/$f
done
