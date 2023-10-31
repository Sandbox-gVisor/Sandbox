#!/bin/bash

# start gVisor
# shellcheck disable=SC1010
./bin/runsc -strace --rootless --network=host --debug --debug-log=/tmp/runsc-bruh/log --syscall-init-config="$1" do bash