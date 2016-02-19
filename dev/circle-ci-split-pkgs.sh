#!/bin/bash

i=0
files=()
for pkg in $(go list ./... | grep -v /vendor/ | sort); do
    if [ $(($i % $CIRCLE_NODE_TOTAL)) -eq $CIRCLE_NODE_INDEX ]
    then
        files+=" $pkg"
    fi
    ((i=i+1))
done

echo ${files[@]}
