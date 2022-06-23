//go:build !windows

package main

import "syscall"

func ForkExec(argv0 string, argv []string, attr *syscall.ProcAttr) (pid int, err error) {
	return syscall.ForkExec(argv0, argv, attr)
}
