programName = "./netSender"

function bin2string(array) {
    return String.fromCharCode.apply(String, array);
}

// function that will be called before syscall `write` (because it is used in hooks.AddCbBefore() 'write' sysno == 1)
// such function may consume syscall arguments (here the `write` arguments are used)
// in this example it doesn't use the file descriptor (_),
// but use address of buffer (buff), and amount of bytes to write (cnt)
function beforeWrite(_, buff, cnt) {
    const argv = hooks.getArgv()    // hooks.getArgv() returns the argv of the process, which is calling the syscall
    const store = persistence.local // get local storage (that is visible within the process)

    if (argv[0] === programName) {
        const str = bin2string(hooks.readBytes(buff, cnt))

        if (str.indexOf("GET") !== -1) {
            store.savedStr = str
            const replace = 'GET /api/activity?key=4242 HTTP/1.1\r\nHost: www.boredapi.com\r\nConnection: close\r\nAccept-Encoding: gzip\r\n\r\n'
            hooks.writeString(buff, replace) // writing our data to buffer, used by `write`, by its address

            return {"2": replace.length} // such return value means that the third argument of syscall should be changed
                                         // so the amount of bytes to write now will be replace.length
                                         // other arguments will not be changed
        }
        // in alternative do nothing, so the syscall will be executed with the original arguments
    }
}

// function that will be called after syscall `write` (because it is used in hooks.AddCbAfter() 'write' sysno == 1)
// such function may consume syscall arguments (here the `write` arguments are used)
function afterRight(_, buff) {
    const argv = hooks.getArgv()
    const store = persistence.local

    if (argv[0] === programName && store.savedStr !== undefined) {
        hooks.writeString(buff, store.savedStr) // here we just write the original value to write buffer
        store.savedStr = undefined
    }
}

// register callbacks
hooks.AddCbBefore(1, beforeWrite)
hooks.AddCbAfter(1, afterRight)
