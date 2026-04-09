package cmd

import (
	"fmt"
	"strings"

	"layouts/internal/tmux"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(splitCmd)
}

var splitCmd = &cobra.Command{
	Use:         "split [preset]",
	Aliases:     []string{"sp"},
	Annotations: map[string]string{"group": "Layouts:"},
	Short:       "Split the current pane using a built-in preset",
	Long: `Split the current tmux pane using a built-in preset.

  layouts split        — list split presets
  layouts split 2cols  — split current pane into 1 row, 2 columns
  layouts split c      — keep current pane on top, add a 2x2 grid below
  layouts split a22    — split current pane into 2 rows, 2 columns`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			printSplitPresets()
			return nil
		}

		if !tmux.IsInsideTmux() {
			return fmt.Errorf("must be inside a tmux session")
		}

		preset, ok := tmux.FindSplitPreset(args[0])
		if !ok {
			return fmt.Errorf("unknown split preset %q\n\n%s", args[0], splitPresetHelp())
		}

		paneTarget, err := tmux.CurrentPaneTarget()
		if err != nil {
			return fmt.Errorf("getting current pane: %w", err)
		}

		dir, err := tmux.CurrentPaneDir()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		if err := tmux.ApplySplitPreset(paneTarget, dir, preset.Name); err != nil {
			return fmt.Errorf("splitting pane: %w", err)
		}

		fmt.Printf("Split current pane using %q (%s)\n", preset.Name, strings.ToLower(preset.Description))
		return nil
	},
}

func printSplitPresets() {
	fmt.Println(splitPresetHelp())
}

func splitPresetHelp() string {
	var b strings.Builder
	b.WriteString("Available split presets:\n")
	for _, preset := range tmux.SplitPresets() {
		b.WriteString(fmt.Sprintf("  %-6s %s\n", preset.Name, preset.Description))
	}
	b.WriteString("\n")
	b.WriteString("Legacy aliases: 12/c12 -> 2cols, 22/c22 -> a22\n")
	b.WriteString("Run `layouts split <preset>` to apply one to the current pane.")
	return b.String()
}
