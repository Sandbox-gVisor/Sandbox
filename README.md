![gVisor](g3doc/logo.png)

# gVisor Sandbox

### Original repository https://github.com/google/gvisor

## JavaScript Engine for System Call Handlers

## Introduction

In this project, we have undertaken the task of patching gVisor, an open-source container runtime sandbox, to integrate a JavaScript (JS) engine. 
The JS engine allows us to execute custom system call handlers written in JavaScript. 
These handlers provide us with valuable information about the running processes, system calls, and their arguments. 
Additionally, we can use our custom functions called "hooks" to modify specific values within the system call handling process.

## Motivation

The motivation behind this project was to extend the capabilities of gVisor and enable more flexible and dynamic handling of system calls. 
With the JS engine integration, we sought to gain insights into the inner workings of processes, manipulate system call arguments, and control system call behavior, all using JavaScript code.

## Documentation and examples

May be found in `examples/gWisord/` or just [click here](./examples/gWisord/README.md)

## Quick launch
Run:
```shell
./init_script.sh your_config.json // this will build and run gVisor
```
more about configuration file may be found [here](./examples/gWisord/configuration/README.md)

If you have already built the gvisor, you may run:
```shell
./run_script your_config.json
```

## Conclusion

The successful integration of a JavaScript engine into gVisor has significantly enhanced its capabilities 
by enabling the use of custom JavaScript-based system call handlers. 
These handlers empower us to extract vital information about processes, manipulate system call arguments, 
and control system call behavior. 
The flexibility offered by the hooks further allows for dynamic customization, 
making gVisor an even more powerful and versatile container runtime sandbox.

The potential applications of this patch range from debugging and monitoring to security analysis and testing, 
making it a valuable addition to gVisor's feature set. 
Further development and testing will continue to refine the system and explore additional use cases.
