package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

type RunningApp struct {
	Name string
	Cmd  *exec.Cmd
	PGID int
}

var runningApps []RunningApp

func main() {
	config, err := LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	if err := os.MkdirAll("apps", 0755); err != nil {
		fmt.Printf("Failed to create apps dir: %v\n", err)
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🔻 Shutting down all apps...")
		for _, app := range runningApps {
			if app.Cmd.Process != nil {
				_ = app.Cmd.Process.Kill()
			}
			_ = killProcessGroup(app.PGID)
			fmt.Printf("Terminated %s\n", app.Name)
		}
		os.Exit(0)
	}()

	for _, app := range config.Apps {
		appDir := filepath.Join("apps", app.Name)
		indexPath := filepath.Join(appDir, "index.js")

		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			fmt.Printf("Cloning %s into %s...\n", app.Repo, appDir)
			cmd := exec.Command("git", "clone", app.Repo, appDir)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error cloning %s: %v\n", app.Name, err)
				continue
			}
		}

		env := os.Environ()
		for k, v := range app.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}

		cmd := exec.Command("node", "index.js")
		cmd.Dir = appDir
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		setupCommand(cmd)

		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to start %s: %v\n", app.Name, err)
			continue
		}

		fmt.Printf("%s is running (PID %d)\n", app.Name, cmd.Process.Pid)

		appInstance := RunningApp{
			Name: app.Name,
			Cmd:  cmd,
		}
		if runtime.GOOS != "windows" {
			if pgid, err := getPGID(cmd.Process.Pid); err == nil {
				appInstance.PGID = pgid
			}
		}
		runningApps = append(runningApps, appInstance)
	}

	select {}
}
