#!/bin/sh

TRACE_SYSCALLS=mmap

apk add strace

for pid in $(pgrep qemu-system); do
	strace --stack-traces --trace="$TRACE_SYSCALLS" -p $pid >> /persist/newlog/collect/strace-qemu-$pid.log 2>&1 &
done
