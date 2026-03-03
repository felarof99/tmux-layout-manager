package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"layouts/internal/config"

	"github.com/spf13/cobra"
)

var configPath bool

func init() {
	configCmd.Flags().BoolVar(&configPath, "path", false, "Print config file path")
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:         "config",
	Aliases:     []string{"c", "cfg"},
	Annotations: map[string]string{"group": "Config:"},
	Short:       "Open or show config",
	Long: `Config management.

  layouts config          — open config in editor
  layouts config --path   — print config file path`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := config.ConfigPath()

		if configPath {
			fmt.Println(p)
			return nil
		}

		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("no config found — run `layouts init` first")
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}

		editor := cfg.Editor
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor == "" {
			editor = "nvim"
		}

		c := exec.Command(editor, p)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
