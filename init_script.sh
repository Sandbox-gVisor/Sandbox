#!/bin/bash

# we have no bin folder in our repository
mkdir bin

# building gVisor and put executable to bin
make copy TARGETS=runsc DESTINATION=bin/

./run_script.sh "$1"

