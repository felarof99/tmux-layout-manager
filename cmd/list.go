package cmd

import (
	"fmt"
	"sort"

	"layouts/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var (
	nameColor   = color.New(color.FgHiGreen, color.Bold)
	dimColor    = color.New(color.Faint)
	defaultMark = color.New(color.FgYellow)
)

var listCmd = &cobra.Command{
	Use:         "list",
	Aliases:     []string{"ls", "l"},
	Annotations: map[string]string{"group": "Layouts:"},
	Short:       "List available layouts",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}

		names := cfg.LayoutNames()
		if len(names) == 0 {
			fmt.Println("No layouts defined. Run `layouts init` to get started.")
			return nil
		}

		sort.Strings(names)

		for _, name := range names {
			layout := cfg.Layouts[name]
			windowCount := len(layout.Windows)
			paneCount := 0
			for _, w := range layout.Windows {
				paneCount += len(w.Panes)
			}

			line := nameColor.Sprint(name)
			if name == cfg.Default {
				line += " " + defaultMark.Sprint("(default)")
			}
			line += " " + dimColor.Sprintf("— %d windows, %d panes", windowCount, paneCount)
			fmt.Println(line)
		}

		return nil
	},
}
