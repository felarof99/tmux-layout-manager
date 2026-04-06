package cmd

import (
	"fmt"

	"layouts/internal/config"
	"layouts/internal/tmux"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(splitCmd)
}

var splitCmd = &cobra.Command{
	Use:         "split <name>",
	Aliases:     []string{"sp"},
	Annotations: map[string]string{"group": "Layouts:"},
	Short:       "Split the current window into panes using a layout",
	Args:        cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tmux.IsInsideTmux() {
			return fmt.Errorf("must be inside a tmux session")
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}

		layout := cfg.FindLayout(args[0])
		if layout == nil {
			return fmt.Errorf("layout %q not found", args[0])
		}

		winTarget, err := tmux.CurrentWindowTarget()
		if err != nil {
			return fmt.Errorf("getting current window: %w", err)
		}

		dir, err := tmux.CurrentPaneDir()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		if err := tmux.SplitCurrentWindow(winTarget, dir, layout); err != nil {
			return fmt.Errorf("splitting window: %w", err)
		}

		fmt.Printf("Split current window using %q (%d panes)\n", args[0], len(layout.Windows[0].Panes))
		return nil
	},
}
