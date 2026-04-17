package cmd

import (
	"fmt"

	"layouts/internal/tmux"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(maximizeCmd)
}

var maximizeCmd = &cobra.Command{
	Use:         "maximize",
	Aliases:     []string{"max", "m", "zoom", "z"},
	Annotations: map[string]string{"group": "Layouts:"},
	Short:       "Toggle maximize for the current tmux pane",
	Long: `Toggle maximize for the current tmux pane.

This uses tmux's native zoom behavior, so running it once maximizes the
current pane and running it again restores the previous split layout.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tmux.IsInsideTmux() {
			return fmt.Errorf("must be inside a tmux session")
		}

		paneTarget, err := tmux.CurrentPaneTarget()
		if err != nil {
			return fmt.Errorf("getting current pane: %w", err)
		}

		zoomed, err := tmux.TogglePaneZoom(paneTarget)
		if err != nil {
			return fmt.Errorf("toggling pane maximize: %w", err)
		}

		if zoomed {
			fmt.Println("Maximized current pane")
			return nil
		}

		fmt.Println("Restored split layout")
		return nil
	},
}
