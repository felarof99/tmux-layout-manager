package cmd

import (
	"fmt"

	"layouts/internal/config"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:         "init",
	Annotations: map[string]string{"group": "Config:"},
	Short:       "Initialize config with example layouts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Init(); err != nil {
			return err
		}
		fmt.Printf("Created config at %s\n", config.ConfigPath())
		fmt.Println("Edit it to define your layouts, then run `layouts list`.")
		return nil
	},
}
