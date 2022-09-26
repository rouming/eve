#!/bin/sh
#
# SPDX-License-Identifier: Apache-2.0

if test -f /proc/vmcore; then
    #
    # We are in dump capture kernel, nice.
    #
    MAX=5
    DIR=/persist/kcrashes

    # Create a folder
    mkdir $DIR > /dev/null 2<&1

    # Keep $MAX-1 fresh dumps
    ls -t $DIR | tail -n +$MAX | xargs --no-run-if-empty -I '{}' rm $DIR/{}

    # Show panic log from the dmesg of a crashed kernel
    makedumpfile --dump-dmesg /proc/vmcore /tmp/dmesg > /dev/null
    echo ">>>>>>>>>> Crashed kernel dmesg BEGIN <<<<<<<<<<" > /dev/kmsg
    sed -n -e '/Kernel panic - not syncing/,$p' /tmp/dmesg | \
        while read line; do echo $line > /dev/kmsg; done
    echo ">>>>>>>>>> Crashed kernel dmesg END <<<<<<<<<<" > /dev/kmsg

    # Collect a minimal kernel dump for security reasons
    KDUMP_PATH=$DIR/kcrash-$(date +%Y-%m-%d-%H-%M-%S).dump
    makedumpfile -d 31 /proc/vmcore $KDUMP_PATH > /dev/null 2>&1
    echo "kdump collected: $KDUMP_PATH" > /dev/kmsg

    # Simulate the default reboot after panic kernel behaviour
    TIMEOUT=`cat /proc/sys/kernel/panic`
    if [ $TIMEOUT -gt 0 ]; then
        echo "Rebooting in $TIMEOUT seconds..." > /dev/kmsg
        sleep $TIMEOUT
    elif [ $TIMEOUT -eq 0 ]; then
        # Wait forever
        while [ 1 ]; do sleep 1; done
    fi

    # Reboot immediately
    umount /persist
    echo b > /proc/sysrq-trigger

    # Unreachable line
fi
