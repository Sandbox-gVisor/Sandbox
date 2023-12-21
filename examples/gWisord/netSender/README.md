# Launch guide for `netSender`

1. Run in the netSender example directory:

    ```shell
    go build
    ```
    
    The result is executable file "netSender"

2. Run our gVisor. **Note** that you **need** config with `runtime-socket` option

    Our `init_script.sh` runs `/bin/bash` inside the gVisor

3. Get to the netSender example directory
4. Run

    ```shell
    ./netSender
    ```

    Every second you will see 
    ```json
    Response body: {"activity":"Learn a new programming language","type":"education","participants":1,"price":0.1,"link":"","key":"5881028","accessibility":0.25}
   
    ```
    
5. Switch to another terminal
6. Configure [sandbox-cli](https://github.com/Sandbox-gVisor/sandbox-cli) to use the same `host` and `port` as gVisor
7. To add the hooks into gVisor run

    ```shell
    ./sandbox-cli state -c Sandbox/examples/gWisord/netSender/hooks.js
    ```
   
    If the script is correct you will see the message:

    ```
    Type: ok
    gVisor says: Everything ok
    Payload: {}
    ```

8. After this you will see that netSender behaviour is changed. Now it will be printing:

   ```json
   Response body: {"error":"No activity found with the specified parameters"}
   ```