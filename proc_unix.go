//go:build darwin || linux

package main

import (
	"os/exec"
	"syscall"
)

func setupCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func getPGID(pid int) (int, error) {
	return syscall.Getpgid(pid)
}

func killProcessGroup(pgid int) error {
	return syscall.Kill(-pgid, syscall.SIGKILL)
}
