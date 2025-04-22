//go:build windows

package main

import (
	"os/exec"

	"syscall"

	"golang.org/x/sys/windows"
)

func setupCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP,
	}
}

func getPGID(_ int) (int, error) {
	return 0, nil // PGID isn't used on Windows
}

func killProcessGroup(_ int) error {
	return nil // No-op
}
