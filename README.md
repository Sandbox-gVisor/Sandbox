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

May be found in `examples/gWisord/` or just [click](./examples/gWisord/README.md)

## Features and Hooks

Our patched gVisor now includes the following key features and hooks:

1. **getPidInfo:**
    - Description: Provides the PID (Process ID), GID (Group ID), UID (User ID), and session information of the task.
    - Arguments: None.
    - Return Values: PidDto JSON object.

2. **getFdsInfo:**
    - Description: Provides information about all file descriptors (fds) of the task.
    - Arguments: None.
    - Return Values: DTO (Data Transfer Object) as an ArrayBuffer containing a marshalled array of JSON objects, each representing an fd.

3. **readBytes:**
    - Description: Reads bytes from the provided address in the task's address space.
    - Arguments: `addr` (number) - Address from which to read the data, `count` (number) - Number of bytes to read.
    - Return Values: ArrayBuffer containing the read data.

4. **writeBytes:**
    - Description: Writes bytes to the provided address in the task's address space.
    - Arguments: `addr` (number) - Address to which data will be written, `buffer` (ArrayBuffer) - Buffer containing the data to be written.
    - Return Values: `counter` (number) - Number of bytes actually written.

5. **writeString:**
    - Description: Writes the provided string to the provided address in the task's address space.
    - Arguments: `addr` (number) - Address from which to write the string, `str` (string) - The string to be written.
    - Return Values: `count` (number) - Number of bytes actually written.

6. **getArgv:**
    - Description: Provides the argv (argument vector) of the task.
    - Arguments: None.
    - Return Values: Array of strings representing the command-line arguments.

7. **getSignalInfo:**
    - Description: Provides the signal masks and sigactions of the task.
    - Arguments: None.
    - Return Values: SignalMaskDto JSON object.

8. **readString:**
    - Description: Reads a string from the provided address in the task's address space.
    - Arguments: `addr` (number) - Address from which to read the string, `count` (number) - Number of bytes to read.
    - Return Values: Read string.

9. **getEnvs:**
    - Description: Provides the environment variables of the task.
    - Arguments: None.
    - Return Values: Array of strings, each in the format `ENV_NAME=env_val`.

10. **getMmaps:**
    - Description: Provides mapping information similar to that found in procfs.
    - Arguments: None.
    - Return Values: String containing mappings similar to procfs.

11. **getFdInfo:**
    - Description: Provides information about a specific file descriptor of the task.
    - Arguments: `fd` (number) - File descriptor to get information about.
    - Return Values: DTO as an ArrayBuffer containing a marshalled JSON object representing the fd.

12. **print:**
    - Description: Prints all passed arguments.
    - Arguments: `msgs` (...any) - Values to be printed.
    - Return Values: `null`.

## Conclusion

The successful integration of a JavaScript engine into gVisor has significantly enhanced its capabilities by enabling the use of custom JavaScript-based system call handlers. These handlers empower us to extract vital information about processes, manipulate system call arguments, and control system call behavior. The flexibility offered by the hooks further allows for dynamic customization, making gVisor an even more powerful and versatile container runtime sandbox.

The potential applications of this patch range from debugging and monitoring to security analysis and testing, making it a valuable addition to gVisor's feature set. Further development and testing will continue to refine the system and explore additional use cases.
