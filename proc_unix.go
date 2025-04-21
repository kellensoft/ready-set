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
