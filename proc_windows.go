//go:build windows

package main

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func setupCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP,
	}
}

func getPGID(pid int) (int, error) {
	return pid, nil
}

func killProcessGroup(pid int) error {
	return exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F").Run()
}

func getCommand(command string) *exec.Cmd {
	parts := strings.Fields(command)
	return exec.Command(parts[0], parts[1:]...)
}
