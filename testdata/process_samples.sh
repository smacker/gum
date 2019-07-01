#!/bin/sh

set -e

BIN=${JAVA_GUM_BIN:-/Users/smacker/Downloads/gumtree-2.1.2/bin/gumtree}
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"


(cd $DIR
    mkdir -p parsed
    for f in `find ./samples -type f -name "*_v[01].java"`; do
        echo "parsing $f"
        mkdir -p `dirname ./parsed/$f`
        $BIN parse $f > ./parsed/$f
    done

    for f in `find ./samples -type f -name "*_v0.java"`; do
        out="./parsed/${f/_v0\.java/_diff.java}"
        echo "generating diff $out"
        $BIN jsondiff $f "${f/_v0/_v1}" > $out
    done
)