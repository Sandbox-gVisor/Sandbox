![gVisor](g3doc/logo.png)

# Project gVisor sandbox

## Introduction
In this report, we outline the progress made in patching gvisor for our project. The main objective of the patching process was to enhance gvisor's functionality and logging capabilities to suit our specific requirements.

## Tasks Completed
1. **Reading at Address**:
    - Implemented the ability to read data from a specified memory address within the container.

2. **Writing at Address**:
    - Implemented the ability to write data to a specified memory address within the container.

3. **Retrieving Syscall Arguments**:
    - Added functionality to retrieve and log syscall arguments for better analysis and debugging.

4. **Printing Mapping**:
    - Implemented the ability to print memory mappings of the container for detailed inspection.

5. **Printing Syscall Arguments**:
    - Extended the logging capabilities to include printing syscall arguments for each executed syscall.

6. **Modifying Syscall Arguments**:
    - Implemented the capability to intercept and modify syscall arguments before execution.

7. **Obtaining pid, gid, uid**:
    - Enhanced logging to include process identifiers (pid), group identifiers (gid), and user identifiers (uid) for better process tracking and auditing.

8. **Additional File Descriptor (fd) Information**:
    - Included more detailed information about file descriptors in the logs for better understanding of the container's I/O operations.

9. **Enhanced Logging**:
    - Added more extensive and informative logs to aid in debugging and monitoring container behavior.

10. **Obtaining Executable Information**:
    - Implemented functionality to retrieve and log information about the executable file associated with each process in the container.

11. **Environment Variables and argv**:
    - Extended logging to include environment variables and command-line arguments (argv) for each process.

12. **Web Interface for Logs**:
    - Developed a web-based interface to visualize and access the generated logs conveniently.

13. **Signal Mask and Signal Handlers**:
    - Included logging of signal masks and signal handlers for each process, enabling better signal-related analysis.

14. **Getting and Setting Niceness**:
    - ability to retrieve and modify the "niceness" of processes.

15. **Process Termination**:
    - enable terminating individual processes or groups of processes.

16. **Umask**:
    - incorporating umask functionality for setting default permissions for newly created files.

17. **Session Information**:
    - gather and log session-related information for the container.

18. **Dynamic Configuration During Runtime**:
    - allowing the container's configuration to be modified during runtime.

19. **Additional Process Information from Handler Scripts**:
    - information about processes from external handler scripts to enrich container monitoring.

## etc.....

## Conclusion
The gvisor patching process has been productive so far, with significant improvements in logging, syscall handling, and process information retrieval. The ongoing tasks are expected to be completed in the near future, further enhancing the functionality and customization capabilities of gvisor for our project's specific needs.

We will continue to monitor progress and provide updates as we complete the remaining tasks.

