#!/bin/bash

# we have no bin folder in our repository
mkdir bin

# building gVisor and put executable to bin
make copy TARGETS=runsc DESTINATION=bin/

# start gVisor
./bin/runsc -strace --rootless --network=host --debug --debug-log=/tmp/runsc-bruh/log --syscall-init-config=$1 do bash

