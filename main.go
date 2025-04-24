package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

type RunningApp struct {
	Name      string
	Cmd       *exec.Cmd
	PGID      int
	StartedAt time.Time
	StoppedAt *time.Time
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
			if app.Cmd != nil && app.Cmd.Process != nil {
				_ = app.Cmd.Process.Kill()
				_ = killProcessGroup(app.PGID)
				fmt.Printf("Terminated %s\n", app.Name)
			}
		}
		os.Exit(0)
	}()

	for _, app := range config.Apps {
		appDir := filepath.Join("apps", app.Name)
		if _, err := os.Stat(appDir); os.IsNotExist(err) {
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

		steps := []struct {
			name string
			cmd  string
		}{
			{"build", app.Commands["build"]},
			{"test", app.Commands["test"]},
			{"start", app.Commands["start"]},
		}

		var startCmd *exec.Cmd
		for _, step := range steps {
			if step.cmd == "" {
				continue
			}
			fmt.Printf("Running %s command for %s...\n", step.name, app.Name)

			cmd := getCommand(step.cmd)
			cmd.Dir = appDir
			cmd.Env = env
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			setupCommand(cmd)

			if err := cmd.Start(); err != nil {
				fmt.Printf("Failed to start %s: %v\n", app.Name, err)
				break
			}

			if step.name == "start" {
				startCmd = cmd
			} else {
				_ = cmd.Wait()
			}
		}

		if startCmd == nil {
			fmt.Printf("No start command for %s, skipping\n", app.Name)
			continue
		}

		fmt.Printf("%s is running (PID %d)\n", app.Name, startCmd.Process.Pid)

		appInstance := RunningApp{
			Name:      app.Name,
			Cmd:       startCmd,
			StartedAt: time.Now(),
		}
		if runtime.GOOS != "windows" {
			if pgid, err := getPGID(startCmd.Process.Pid); err == nil {
				appInstance.PGID = pgid
			}
		}
		runningApps = append(runningApps, appInstance)
	}

	go startWebService()
	select {}
}

func startWebService() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		type AppStatus struct {
			Name          string     `json:"name"`
			Running       bool       `json:"running"`
			StartedAt     time.Time  `json:"started_at"`
			StoppedAt     *time.Time `json:"stopped_at,omitempty"`
			UptimeSeconds int64      `json:"uptime_seconds"`
		}

		var status []AppStatus
		now := time.Now()

		for _, app := range runningApps {
			running := app.Cmd.ProcessState == nil || !app.Cmd.ProcessState.Exited()
			uptime := now.Sub(app.StartedAt).Seconds()
			if app.StoppedAt != nil {
				uptime = app.StoppedAt.Sub(app.StartedAt).Seconds()
			}
			status = append(status, AppStatus{
				Name:          app.Name,
				Running:       running,
				StartedAt:     app.StartedAt,
				StoppedAt:     app.StoppedAt,
				UptimeSeconds: int64(uptime),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	fmt.Println("Ready-Set: live on 8080")
	http.ListenAndServe(":8080", nil)
}
