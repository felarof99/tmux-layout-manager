package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"layouts/internal/config"
	"layouts/internal/tmux"

	"github.com/spf13/cobra"
)

var applyDir string

func init() {
	applyCmd.Flags().StringVarP(&applyDir, "dir", "d", "", "Working directory for panes (defaults to current pane dir)")
	rootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:         "apply [name]",
	Aliases:     []string{"a"},
	Annotations: map[string]string{"group": "Layouts:"},
	Short:       "Apply a layout to the current tmux session",
	Long: `Apply a layout to the current tmux session.

  layouts apply          — pick layout via fzf
  layouts apply <name>   — apply named layout
  layouts apply -d .     — apply using current directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tmux.IsInsideTmux() {
			return fmt.Errorf("must be inside a tmux session")
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}

		var layoutName string
		if len(args) >= 1 {
			layoutName = args[0]
		} else if cfg.Default != "" {
			layoutName = cfg.Default
		} else {
			picked, err := pickLayoutFzf(cfg)
			if err != nil {
				return err
			}
			layoutName = picked
		}

		layout := cfg.FindLayout(layoutName)
		if layout == nil {
			return fmt.Errorf("layout %q not found", layoutName)
		}

		sessionName, err := tmux.CurrentSession()
		if err != nil {
			return fmt.Errorf("getting current session: %w", err)
		}

		dir := applyDir
		if dir == "" {
			dir, err = tmux.CurrentPaneDir()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
		}

		if err := tmux.ApplyLayout(sessionName, dir, layout); err != nil {
			return fmt.Errorf("applying layout: %w", err)
		}

		fmt.Printf("Applied layout %q (%d windows)\n", layoutName, len(layout.Windows))
		return nil
	},
}

func pickLayoutFzf(cfg *config.Config) (string, error) {
	names := cfg.LayoutNames()
	if len(names) == 0 {
		return "", fmt.Errorf("no layouts defined in config")
	}

	fzfCmd := exec.Command("fzf",
		"--prompt", "layout > ",
		"--header", "Pick a layout",
		"--height", "~40%",
		"--reverse",
	)
	fzfCmd.Stdin = strings.NewReader(strings.Join(names, "\n"))
	fzfCmd.Stderr = os.Stderr

	out, err := fzfCmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && (exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130) {
			return "", ErrCancelled
		}
		return "", fmt.Errorf("fzf: %w", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return "", ErrCancelled
	}
	return result, nil
}
