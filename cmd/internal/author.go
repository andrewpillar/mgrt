package internal

import (
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

func git(subcmd string, args ...string) (string, string, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := exec.Command("git", append([]string{subcmd}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// mgrtAuthor will attempt to get author information from git using the
// config.name and config.email properties. If this fails, then it falls back
// to getting the current user's username.
func mgrtAuthor() (string, error) {
	stdout, _, err := git("config", "user.name")

	if err != nil {
		u, err := user.Current()

		if err != nil {
			return "", err
		}
		return u.Username, nil
	}

	name := strings.TrimSpace(stdout)

	stdout, _, err = git("config", "user.email")

	if err != nil {
		return name, nil
	}
	return name + " <" + strings.TrimSpace(stdout) + ">", nil
}
