package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func run(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux %s: %s (%w)", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func CurrentSession() (string, error) {
	return run("display-message", "-p", "#{session_name}")
}

func CurrentPaneDir() (string, error) {
	return run("display-message", "-p", "#{pane_current_path}")
}

func SessionExists(name string) bool {
	_, err := run("has-session", "-t", "="+name)
	return err == nil
}

func NewSession(name, startDir string) error {
	_, err := run("new-session", "-d", "-s", name, "-c", startDir)
	return err
}
