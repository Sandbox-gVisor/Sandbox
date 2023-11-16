# Launch guide for `allAddressesAlreadyInUse`

1. Run in the allAddressesAlreadyInUse example directory:

    ```shell
    make
    ```

   The result is executable file "main"

2. Run our gVisor. **Note** that you **need** config with `runtime-socket` option

   Our `init_script.sh` builds gVisor and runs `/bin/bash` inside the gVisor

3. Get to the allAddressesAlreadyInUse example directory
4. Run

    ```shell
    ./main {port}
    ```
   where port is the port to bind.
   
   If `localhost:{port}` is not used the program will print
   ```
   Successfully bind to port {port} :)
   ```

5. Switch to another terminal
6. Configure [sandbox-cli](https://github.com/Sandbox-gVisor/sandbox-cli) to use the same `host` and `port` as gVisor
7. To add the hooks into gVisor run

   ```shell
   ./sandbox-cli state -c Sandbox/examples/gWisord/allAddressesAlreadyInUse/hooks.js
   ```

   If the script is correct you will see the message:

   ```
   Type: ok
   gVisor says: Everything ok
   ```

8. Now every execution of 

   ```shell
   ./main {port}
   ```
   will fail because of bind error
