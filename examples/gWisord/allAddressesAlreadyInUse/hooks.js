hooks.AddCbBefore(49, failBind)

function failBind() {
    return {        // specifying both return value and errno to replace
        "ret": -1,  // return value
        "errno": 98 // errno
    }               // because of specified return value and errno the syscall won't be executed
}
