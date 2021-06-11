package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Git Clone Repo helper to speed up git clone")
		fmt.Println("need repo url")
		os.Exit(1)
	}
	dest_dir := ""
	repo_url := os.Args[1]
	repo_url = strings.ReplaceAll(repo_url, "https://github.com/", "https://github.com.cnpmjs.org/")
	if len(os.Args) > 2 {
		dest_dir = os.Args[2]

	}
	wd, _ := os.Getwd()
	args := []string{
		// "git",
		"clone",
		repo_url,
	}
	if dest_dir != "" {
		args = append(args, dest_dir)
	} else {
		dest_dir = filepath.Base(repo_url)
		dest_dir = strings.ReplaceAll(dest_dir, ".git", "")
	}
	if _, err := os.Stat(dest_dir); !os.IsNotExist(err) {
		ExecuteCommand(dest_dir, "git", []string{
			// "git",
			"remote",
			"-v",
		}...)
		ExecuteCommand(dest_dir, "git", []string{
			// "git",
			"status",
		}...)
		ExecuteCommand(dest_dir, "git", []string{
			// "git",
			"pull",
			"--rebase",
		}...)
		os.Exit(0)
	}
	ExecuteCommand(wd, "git", args...)
	os.Exit(0)
}
func ExecuteCommand(wd, name string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if e := cmd.Start(); e != nil {
		fmt.Printf("fail to start cmd for %s\n", e)
		os.Exit(1)
	}
	if e := cmd.Wait(); e != nil {
		if ee, ok := e.(*exec.ExitError); ok {
			code := ee.ExitCode()
			if code != -1 {
				fmt.Printf("error is not -1 %s %s\n", e.Error(), string(ee.Stderr))
				os.Exit(1)
			} else {
				fmt.Printf("error code -1\n")
				os.Exit(1)
			}
		}
	}
}
