package main

import (
	"bytes"
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println(err)
	}

}

var repo = flag.String("repo", "", "github repository")

func prUrl(branch string) string {
	baseUrl := "https://github.com"
	return baseUrl + "/" + *repo + "/compare/master..." + branch + "?template=release.md&expand=1"
}

func main() {
	flag.Parse()
	exec.Command("git", "fetch", "-all").Run()
	args := []string{"for-each-ref", "--sort=-committerdate", "refs/heads", "refs/remotes", "--format=%(color:yellow)%(refname:short)%(color:reset)"}
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	releaseFiles := strings.Split(out.String(), "\n")
	for _, branch := range releaseFiles {
		strings.Replace(branch, "origin/release/", "release/", 1)
		if strings.Contains(branch, "release/") {
			openbrowser(prUrl(branch))
			break
		}
	}
}
