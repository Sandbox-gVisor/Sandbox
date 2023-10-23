programName = "./netSender"

function bin2string(array) {
    return String.fromCharCode.apply(String, array);
}

function beforeWrite(_, buff, cnt) {
    const argv = hooks.getArgv()
    const store = persistence.local

    if (argv[0] === programName) {
        const str = bin2string(hooks.readBytes(buff, cnt))

        if (str.indexOf("GET") !== -1) {
            store.savedStr = str
            const replace = 'GET /api/activity?key=4242 HTTP/1.1\r\nHost: www.boredapi.com\r\nConnection: close\r\nAccept-Encoding: gzip\r\n\r\n'
            hooks.writeString(buff, replace)

            return {"2": replace.length}
        }
    }
}

function afterRight(_, buff) {
    const argv = hooks.getArgv()
    const store = persistence.local

    if (argv[0] === programName && store.savedStr !== undefined) {
        hooks.writeString(buff, store.savedStr)
        store.savedStr = undefined
    }
}

hooks.AddCbBefore(1, beforeWrite)
hooks.AddCbAfter(1, afterRight)