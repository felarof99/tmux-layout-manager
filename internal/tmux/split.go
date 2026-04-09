package tmux

import (
	"fmt"
	"strconv"
	"strings"
)

type SplitPreset struct {
	Name        string
	Description string
	Aliases     []string
}

var splitPresets = []SplitPreset{
	{
		Name:        "2cols",
		Description: "1 row, 2 columns",
		Aliases:     []string{"12", "c12"},
	},
	{
		Name:        "c",
		Description: "Additional 2 rows, 2 columns at the bottom",
		Aliases:     []string{"current", "ca22"},
	},
	{
		Name:        "a22",
		Description: "2 rows, 2 columns",
		Aliases:     []string{"22", "c22"},
	},
}

func SplitPresets() []SplitPreset {
	out := make([]SplitPreset, len(splitPresets))
	copy(out, splitPresets)
	return out
}

func FindSplitPreset(name string) (SplitPreset, bool) {
	needle := strings.ToLower(strings.TrimSpace(name))
	for _, preset := range splitPresets {
		if needle == preset.Name {
			return preset, true
		}
		for _, alias := range preset.Aliases {
			if needle == alias {
				return preset, true
			}
		}
	}
	return SplitPreset{}, false
}

func ApplySplitPreset(paneTarget, startDir, name string) error {
	preset, ok := FindSplitPreset(name)
	if !ok {
		return fmt.Errorf("unknown split preset %q", name)
	}

	switch preset.Name {
	case "2cols":
		return splitPaneTwoColumns(paneTarget, startDir)
	case "c":
		return splitPaneKeepTopAddBottomGrid(paneTarget, startDir)
	case "a22":
		return splitPaneGrid(paneTarget, startDir, 2, 2)
	default:
		return fmt.Errorf("unsupported split preset %q", preset.Name)
	}
}

func splitPaneTwoColumns(paneTarget, startDir string) error {
	if _, err := splitPane(paneTarget, "-h", 50, startDir); err != nil {
		return err
	}
	run("select-pane", "-t", paneTarget)
	return nil
}

// splitPaneKeepTopAddBottomGrid keeps the current pane as the top row and
// turns the bottom two-thirds into a 2x2 grid. This matches the "teamwork"
// workflow: split a right-side pane without disturbing the left pane.
func splitPaneKeepTopAddBottomGrid(paneTarget, startDir string) error {
	bottomPane, err := splitPane(paneTarget, "-v", 67, startDir)
	if err != nil {
		return err
	}

	bottomRight, err := splitPane(bottomPane, "-h", 50, startDir)
	if err != nil {
		return err
	}

	if _, err := splitPane(bottomPane, "-v", 50, startDir); err != nil {
		return err
	}
	if _, err := splitPane(bottomRight, "-v", 50, startDir); err != nil {
		return err
	}

	run("select-pane", "-t", paneTarget)
	return nil
}

func splitPaneGrid(paneTarget, startDir string, rows, cols int) error {
	if rows < 1 || cols < 1 {
		return fmt.Errorf("grid must have at least 1 row and 1 column")
	}
	if rows == 1 && cols == 1 {
		return nil
	}

	columnPanes := []string{paneTarget}
	for c := 1; c < cols; c++ {
		target := columnPanes[len(columnPanes)-1]
		remaining := cols - c
		p := remaining * 100 / (remaining + 1)
		newPane, err := splitPane(target, "-h", p, startDir)
		if err != nil {
			return fmt.Errorf("creating column %d: %w", c+1, err)
		}
		columnPanes = append(columnPanes, newPane)
	}

	for _, columnPane := range columnPanes {
		if err := splitPaneRows(columnPane, startDir, rows); err != nil {
			return err
		}
	}

	run("select-pane", "-t", paneTarget)
	return nil
}

func splitPaneRows(paneTarget, startDir string, rows int) error {
	for r := 1; r < rows; r++ {
		remaining := rows - r
		p := remaining * 100 / (remaining + 1)
		if _, err := splitPane(paneTarget, "-v", p, startDir); err != nil {
			return fmt.Errorf("creating row %d: %w", r+1, err)
		}
	}
	return nil
}

func splitPane(target, splitFlag string, percent int, startDir string) (string, error) {
	return run(
		"split-window",
		splitFlag,
		"-t", target,
		"-p", strconv.Itoa(percent),
		"-c", startDir,
		"-P",
		"-F", "#{pane_id}",
	)
}
