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
	Use:         "split [spec]",
	Aliases:     []string{"sp"},
	Annotations: map[string]string{"group": "Layouts:"},
	Short:       "Split the current pane or relayout the current window using a generated grid spec",
	Long: `Split the current tmux pane or relayout the current tmux window using a generated grid spec.

  layouts split       — show split syntax and examples
  layouts split 22    — split current pane into 2 rows, 2 columns
  layouts split 23    — split current pane into 2 rows, 3 columns
  layouts split c22   — relayout the current window into 2 rows, 2 columns
  layouts split c23   — relayout the current window into 2 rows, 3 columns
  layouts split 3x4   — split current pane into 3 rows, 4 columns
  layouts split c3x4  — relayout the current window into 3 rows, 4 columns`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			printSplitPresets()
			return nil
		}

		if !tmux.IsInsideTmux() {
			return fmt.Errorf("must be inside a tmux session")
		}

		spec, ok := tmux.ParseSplitSpec(args[0])
		if !ok {
			return fmt.Errorf("invalid split spec %q\n\n%s", args[0], splitPresetHelp())
		}

		paneTarget, err := tmux.CurrentPaneTarget()
		if err != nil {
			return fmt.Errorf("getting current pane: %w", err)
		}

		windowTarget, err := tmux.CurrentWindowTarget()
		if err != nil {
			return fmt.Errorf("getting current window: %w", err)
		}

		dir, err := tmux.CurrentPaneDir()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		if err := tmux.ApplySplitSpec(windowTarget, paneTarget, dir, spec); err != nil {
			return fmt.Errorf("applying split spec: %w", err)
		}

		if spec.UseCurrentWindow {
			fmt.Printf("Relaid out current window using %q (%s)\n", spec.Name, spec.Description)
			return nil
		}

		fmt.Printf("Split current pane using %q (%s)\n", spec.Name, spec.Description)
		return nil
	},
}

func printSplitPresets() {
	fmt.Println(splitPresetHelp())
}

func splitPresetHelp() string {
	var b strings.Builder
	b.WriteString("Split spec syntax:\n")
	b.WriteString("  <rows><cols>   split current pane into a grid (single-digit form)\n")
	b.WriteString("  <rows>x<cols>  split current pane into a grid (multi-digit form)\n")
	b.WriteString("  c<spec>        relayout the current window into that grid and keep existing panes inside it\n")
	b.WriteString("\nExamples:\n")
	for _, spec := range tmux.SplitExamples() {
		b.WriteString(fmt.Sprintf("  %-6s %s\n", spec.Name, spec.Description))
	}
	b.WriteString("\n")
	b.WriteString("Aliases: 2cols -> 12, a22 -> 22, c/current/ca22 -> c22\n")
	b.WriteString("Run `layouts split <spec>` to apply it to the current pane or current window.")
	return b.String()
}
