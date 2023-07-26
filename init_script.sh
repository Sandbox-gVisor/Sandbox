#!/bin/bash

# we have no bin folder in our repository
mkdir bin

# building gVisor and put executable to bin
sudo make copy TARGETS=runsc DESTINATION=bin/

# here we should start frontend, backend, redis and broker
sudo docker compose up -d

# start gVisor
./bin/runsc -strace --rootless --network=host --debug --debug-log=/tmp/runsc-bruh/log --syscall-init-config=$1 do bash

# stop containers after exiting from gVisor (works ap planned only if we execute bash)
sudo docker compose down -v
