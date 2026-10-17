#!/bin/bash

XEPHYR=$(whereis -b Xephyr | sed -E 's/^.*: ?//')
if [ -z "$XEPHYR" ]; then
    echo "Xephyr not found, exiting"
    exit 1
fi

xinit ./xinitrc -- \
    "$XEPHYR" \
        :2 \
        -br \
        -screen 1024x768 \
        -softCursor 
