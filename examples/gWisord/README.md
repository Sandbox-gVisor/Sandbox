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
- API provided by gVisor (full list of available functions you may see in [below](#list-of-api-functions))
```js
hooks.print("my message") // "hooks" is reseved key word for our API
```
- local and global storage **// TODO**

## Callback registration
You have 2 ways to register your callback
- Call `hooks.AddCbBefore(...)` or `hooks.AddCbAfter(...)` ([see below](#list-of-api-functions))
- Set them in config (see in configuration info for more info)

## Callback before
Has the following abilities:
- get syscall arguments
- set:
  - new values for syscall arguments
  - both new syscall return value and errno (if syscall **new return value and errno** is specified the **syscall** will **NOT be executed**)

## Callback after
Has the following abilities:
- get syscall arguments
- set:
    - new values for syscall arguments
    - new syscall return value 

# Examples
- [Substitution of GET request](./netSender/README.md)
- [Failing the execution of syscall every time](allAddressesAlreadyInUse/README.md)

# List of API functions

Some API functions have object as return value. The structure of such objects you can see below the table

| func name         | arguments                               | return value             | description                                                                                                            |
|-------------------|-----------------------------------------|--------------------------|------------------------------------------------------------------------------------------------------------------------|
| AddCbBefore       | sysno `number`<br/>cb `function`        | `null`                   | Registers function (**cb**) which will be executed __before__ syscall with number == **sysno**                         |
| AddCbAfter        | sysno `number`<br/>cb `function`        | `null`                   | Registers function (**cb**) which will be executed __after__ syscall with number == **sysno**                          |
| anonMmap          | length `number`                         | `number`                 | Allocates **length** bytes in process memory. **Returns** the start address of memory region                           |
| getArgv           | -                                       | `[]string`               | **Returns** array of strings which is the command line arguments                                                       |
| getEnvs           | -                                       | `[]string`               | **Returns** the array of environment variables (string, which have format like ENVIRONMENT_NAME=environment_value)     |
| getFdInfo         | fd `number`                             | `object (FdInfoDto)`     | **Returns** the dto, which provides info about task's file description by given **fd**                                 |
| getFdsInfo        | -                                       | `[]object (FdInfoDto)`   | **Returns** the array of dto, each dto provides info for some task's file description                                  |
| getMmaps          | -                                       | `string`                 | **Returns** string, that represents mappings of the task (looks like mappings from procfs)                             |
| getPidInfo        | -                                       | `object (PidInfoDto)`    | **Returns** the dto, which provides info about task's PID, GID, UID, session                                           |
| getSignalInfo     | -                                       | `object (SignalInfoDto)` | **Returns** the dto, which provides info about task's signal masks and sigactions                                      |
| getThreadInfo     | - <br/> **or** <br/> tid `number`       | `object (ThreadInfoDto)` | **Returns** the dto, which provides TID, TGID (PID) and list of other TIDs in thread group.                            |
| logJson           | msg `any`                               | `null`                   | Sends the given **msg** to log socket                                                                                  |
| munmap            | addr `number`<br/> length `number`      | `null`                   | Delete the mappings from the specified address range by given **addr** and **length** of the region                    |
| nameToSignal      | name `string`                           | `number`                 | **Returns** the number of the signal by provided **name**                                                              |
| print             | msgs `...any`                           | `null`                   | Prints all the given **msgs**                                                                                          |
| readBytes         | addr `number`<br/> count `number`       | `ArrayBuffer`            | Reads **count** bytes from memory by given **addr**. **Returns** the bytes read                                        |
| readString        | addr `number`<br/> count `number`       | `string`                 | Reads the string (string.length <= **count**) by given **addr**. **Returns** the read string                           |
| resumeThreads     | -                                       | `null`                   | Resume threads stopped by `stopThreads`.                                                                               |
| sendSignal        | tid `number`<br/> signo `number`        | `null`                   | Sends to task with tid == **tid** the signal with number **signo**                                                     |
| signalMaskToNames | mask `number`                           | `[]string`               | Parses provided signal **mask** to signal names. **Returns** array of strings - names of signals specified in the mask |
| stopThreads       | -                                       | `null`                   | Stop all threads except the caller. May be useful for preventing TOCTOU attack.                                        |
| writeBytes        | addr `number`<br/> buffer `ArrayBuffer` | `number`                 | Writes to memory the given **buffer** by the given **addr**. **Returns** the amount of really written bytes            |
| writeString       | addr `number`<br/> str `string`         | `number`                 | Writes the given **str** by given **addr**. **Returns** the amount of bytes really written                             |

```
SignalInfoDto = {
  signalMask `number`       // signal mask of the task
  signalWaitMask `number`   // task will be blocked until one of signals in signalWaitMask is pending
  savedSignalMask `number`  // savedSignalMask is the signal mask that should be applied after the task has either delivered one signal to a user handler or is about to resume execution in the untrusted application
  sigactions [
    handler `string`
    flags `string`      
    restorer `number`
    signalsInSet `[]string` // array of strings, each string is a signal name      
  ]
}

PidInfoDto = {
  PID `number`
  GID `number`
  UID `number`
  session {
    sessionID `number`
    PGID `number`
    foregroundID `number`
    otherPGIDs `[]number`
  }
}

FdInfoDto = {
  fd `number`
  name `string`       // file path
  mode `string`       // mode like rwxr--r--
  flags `string`      // flags of the file
  nlinks `number`
  readble `boolean`
  writable `boolean`
}
```
