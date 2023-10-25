# Description

This directory provides some examples of using gWisord and some info for it's configuration. 
Also, here you may found info about writing js scripts to interact with gVisor  

# Configuration info

May be found [here](configuration/README.md)

# Base info

With proper configuration gVisor may use js callbacks, which has ability to modify syscall arguments, return values
to allow or prohibit execution of a system calls and to do some other features. 
Callbacks should be written in some files.

Callback is registered for special syscall, and will be executed only if syscall is used.

Note that each callback is stored as string, so goja interprets the callback each time it should be executed.

For each syscall user can specify 2 callbacks:
- callback, which will be executed **before** syscall
- callback, which will be executed **after** syscall

Both callbacks can use:
- API provided by gVisor (full list of available functions you may see in [TODO]())
```js
hooks.print("my message") // "hooks" is reseved key word for our API
```
- local and global storage **// TODO**

## Callback before
Has the following abilities:
- get syscall arguments
- set:
  - new values for syscall arguments
  - new syscall return value (if syscall **new return value** is specified the **syscall** will **NOT be executed**)

## Callback after
Has the following abilities:
- get syscall arguments
- set:
    - new values for syscall arguments
    - new syscall return value 

# Examples
- [Substitution of GET request](./netSender)