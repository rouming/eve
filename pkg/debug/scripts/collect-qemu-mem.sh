#!/bin/sh

apk add procps gawk >/dev/null 2>&1

DATE=$(date -Is)
echo $DATE
mkdir -p /persist/qemu-mem/$DATE

proc_maps() {
    cat /proc/$1/maps | gawk -n -F '[- ]' '{
        a = "0x" $1
        b = "0x" $2
        v = b - a
        gb = v/1024/1024/1024
        mb = v/1024/1024
        kb = v/1024
        unit = "b"

        if (gb >= 1.0) { v = gb; unit = "Gb"; }
        else if (mb >= 1.0) { v = mb; unit = "Mb"; }
        else if (kb >= 1.0) { v = kb; unit = "Kb"; }
        printf("%4.0f%s %s\n", v, unit, $0);}'
}

for pid in $(pgrep qemu-system); do
    ps -wu -p $pid | tee -a /persist/qemu-mem/$DATE/ps-awux-grep-qemu
    cat /proc/$pid/maps >  /persist/qemu-mem/$DATE/$pid-maps
    proc_maps $pid > /persist/qemu-mem/$DATE/$pid-proc-maps
    uuid=$(ps --no-headers -wu -p $pid | awk '{print $24}')
    find /sys/fs/cgroup/memory/eve-user-apps/$uuid -type f -print -exec cat {} \; > /persist/qemu-mem/$DATE/$uuid-cgroup-memory \
		 2>/dev/null
done
