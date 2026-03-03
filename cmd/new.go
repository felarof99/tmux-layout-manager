package cmd

import (
	"fmt"

	"layouts/internal/config"
	"layouts/internal/tmux"

	"github.com/spf13/cobra"
)

var newDir string

func init() {
	newCmd.Flags().StringVarP(&newDir, "dir", "d", ".", "Working directory for the session")
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:         "new <session-name> [layout]",
	Aliases:     []string{"n"},
	Annotations: map[string]string{"group": "Layouts:"},
	Short:       "Create a new tmux session with a layout",
	Long: `Create a new tmux session and apply a layout to it.

  layouts new mysession dev        — create "mysession" with dev layout
  layouts new mysession            — create "mysession" with default layout
  layouts new mysession -d ~/code  — use ~/code as working directory`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}

		sessionName := args[0]

		if tmux.SessionExists(sessionName) {
			return fmt.Errorf("session %q already exists", sessionName)
		}

		var layoutName string
		if len(args) >= 2 {
			layoutName = args[1]
		} else if cfg.Default != "" {
			layoutName = cfg.Default
		}

		var layout *config.LayoutConfig
		if layoutName != "" {
			layout = cfg.FindLayout(layoutName)
			if layout == nil {
				return fmt.Errorf("layout %q not found", layoutName)
			}
		}

		if err := tmux.CreateSessionWithLayout(sessionName, newDir, layout); err != nil {
			return fmt.Errorf("creating session: %w", err)
		}

		msg := fmt.Sprintf("Created session %q", sessionName)
		if layoutName != "" {
			msg += fmt.Sprintf(" with layout %q", layoutName)
		}
		fmt.Println(msg)

		if tmux.IsInsideTmux() {
			fmt.Printf("Switch with: tmux switch-client -t %s\n", sessionName)
		} else {
			fmt.Printf("Attach with: tmux attach -t %s\n", sessionName)
		}
		return nil
	},
}
