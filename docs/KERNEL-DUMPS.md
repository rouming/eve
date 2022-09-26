# Intruduction to kdump and kexec

`kdump` is a service which provides a crash dumping mechanism. The service enables you to save the contents of the system memory for analysis.

`kdump` uses the kexec system call to boot into the second kernel (a capture kernel) without rebooting; and then captures the contents of the crashed kernel’s memory (a crash dump or a vmcore) and saves it into a file. The second kernel resides in a reserved part of the system memory.