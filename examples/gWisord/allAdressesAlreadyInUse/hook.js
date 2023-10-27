
// The name of this function means that this callback
// will be executed before syscall with sysno == 49 (bind)
function syscall_before_49() {
    return {        // specifying both return value and errno to replace
        "ret": -1,  // return value
        "errno": 98 // errno
    }               // because of specified return value and errno the syscall won't be executed
}