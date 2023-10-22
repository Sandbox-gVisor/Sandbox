# Configuration info

Please see example of configuration file: [conf.json](conf.json)

# Options description
## `runtime-socket`

The value of this option is string like `"{host}:{port}"`.

If this option is specified gVisor will be listening on it for requests.
Users can interact with already running gVisor using such url.

gVisor uses custom protocol for requests and responses. 

Now the simplest way to communicate with gVisor is to use [sandbox-cli](https://github.com/Sandbox-gVisor/sandbox-cli)

## `log-socket`

The value of this option is also string like `"{host}:{port}"`.

If this option is specified gVisor will be sending strace logs to `log-socket` (don't forget to specify `-strace` option then running runsc).
Users can send custom logs to this socket by using hooks.log()

If nobody listens on `log-socket` and the option was specified gVisor will exit immediately.

## `callbacks`

// TODO