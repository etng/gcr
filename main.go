package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

var prefixes = []string{
	"git@",
	"https://",
	"http://",
}

func isRepoUrl(u string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(u, prefix) {
			return true
		}
	}
	return false
}

type ShallowConfig struct {
	Parent string   `json:"parent,omitempty" yaml:"parent" mapstructure:"parent"`
	Urls   []string `json:"urls,omitempty" yaml:"urls" mapstructure:"urls"`
}
type GcrConfig struct {
	Shallows []*ShallowConfig `json:"shallows,omitempty" yaml:"shallows" mapstructure:"shallows"`
}

func ExpandPath(fp string) string {
	if strings.HasPrefix(fp, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, fp[2:])
	}
	return fp
}
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Git Clone Repo helper to speed up git clone")
		fmt.Println("need repo url")
		os.Exit(1)
	}
	if os.Args[1] == "us" {
		log.Printf("doing shallow git pull batch update")
		configPath := ExpandPath("~/.gcr.yml")

		if body, e := ioutil.ReadFile(configPath); e == nil {
			var config GcrConfig
			if e := yaml.Unmarshal(body, &config); e != nil {
				print("fail to decode yaml")
				panic(e)
			}
			log.Printf("updating shallow copy repos")
			chCC := make(chan struct{}, 5)
			for _, shallowConfig := range config.Shallows {
				for _, _repoUrl := range shallowConfig.Urls {
					repoUrl := _repoUrl
					chCC <- struct{}{}
					go func() {
						log.Printf("checking shallow git pull %s", repoUrl)
						GitCloneRepo(nil, ExpandPath(shallowConfig.Parent), "", repoUrl, true, []string{"--depth", "1"}...)
						<-chCC
					}()
				}
			}
		} else {
			log.Printf("fail to read config file %s", configPath)
		}
		return
	}
	dest_dir := ""
	repo_url := ""

	args := os.Args[1:]
	idx := len(args) - 1
	if isRepoUrl(args[idx]) {
		repo_url = args[idx]
		idx -= 0
	} else {
		repo_url = args[idx-1]
		dest_dir = args[idx]
		idx -= 1
	}
	wd, _ := os.Getwd()
	GitCloneRepo(nil, wd, dest_dir, repo_url, false, args[:idx]...)
	os.Exit(0)
}
func isDir(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}
func GitCloneRepo(wg *sync.WaitGroup, wd, dest_dir, repo_url string, removeOld bool, gitArgs ...string) {
	if wg != nil {
		defer wg.Done()
	}
	repo_url = strings.ReplaceAll(repo_url, "https://github.com/", "https://github.com.cnpmjs.org/")
	guess_dest_dir := strings.ReplaceAll(filepath.Base(repo_url), ".git", "")
	if dest_dir == "" {
		dest_dir = filepath.Join(wd, guess_dest_dir)
	} else if isDir(dest_dir) && !isDir(filepath.Join(dest_dir, ".git")) {
		dest_dir = filepath.Join(dest_dir, guess_dest_dir)
	}
	log.Printf("checking out from %q to %q with args %v", repo_url, dest_dir, gitArgs)
	gitArgs = append(gitArgs, repo_url, dest_dir)

	if di, err := os.Stat(dest_dir); !os.IsNotExist(err) {
		if removeOld {
			if time.Since(di.ModTime()).Hours() < 24 {
				// stop delete and checkout cycle
				log.Printf("%q is modified recently, skip delete-and-checkout", dest_dir)
				return
			}
			os.RemoveAll(dest_dir)
		} else {
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
			return
		}
		// os.Exit(0)
	}
	ExecuteCommand(wd, "git", append([]string{"clone"}, gitArgs...)...)

}
func ExecuteCommand(wd, name string, args ...string) {
	log.Printf("git %s", strings.Join(args, " "))
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
