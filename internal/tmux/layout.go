package tmux

import (
	"fmt"
	"strconv"
	"strings"

	"layouts/internal/config"
)

func ApplyLayout(sessionName, startDir string, layout *config.LayoutConfig) error {
	if layout == nil || len(layout.Windows) == 0 {
		return fmt.Errorf("layout has no windows")
	}
	return addWindows(sessionName, startDir, layout.Windows)
}

func CreateSessionWithLayout(name, startDir string, layout *config.LayoutConfig) error {
	if err := NewSession(name, startDir); err != nil {
		return err
	}
	if layout == nil || len(layout.Windows) == 0 {
		return nil
	}
	return applyToNewSession(name, startDir, layout.Windows)
}

func baseIndex() int {
	out, err := run("show-option", "-gv", "base-index")
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(out))
	return n
}

func addWindows(sessionName, startDir string, windows []config.WindowConfig) error {
	for _, win := range windows {
		out, err := run("new-window", "-t", "="+sessionName, "-n", win.Name, "-c", startDir, "-P", "-F", "#{window_index}")
		if err != nil {
			return fmt.Errorf("creating window %s: %w", win.Name, err)
		}
		winIdx, _ := strconv.Atoi(strings.TrimSpace(out))

		if err := applyPanes(sessionName, winIdx, startDir, win); err != nil {
			return err
		}
	}
	return nil
}

func applyToNewSession(sessionName, startDir string, windows []config.WindowConfig) error {
	base := baseIndex()
	var firstWinIdx int

	for i, win := range windows {
		var winIdx int
		if i == 0 {
			winIdx = base
			if _, err := run("rename-window", "-t", fmt.Sprintf("=%s:%d", sessionName, winIdx), win.Name); err != nil {
				return fmt.Errorf("renaming window: %w", err)
			}
			firstWinIdx = winIdx
		} else {
			out, err := run("new-window", "-t", "="+sessionName, "-n", win.Name, "-c", startDir, "-P", "-F", "#{window_index}")
			if err != nil {
				return fmt.Errorf("creating window %s: %w", win.Name, err)
			}
			winIdx, _ = strconv.Atoi(strings.TrimSpace(out))
		}

		if err := applyPanes(sessionName, winIdx, startDir, win); err != nil {
			return err
		}
	}

	firstWin := fmt.Sprintf("=%s:%d", sessionName, firstWinIdx)
	run("select-window", "-t", firstWin)
	run("select-pane", "-t", firstWin+".0")

	return nil
}

func applyPanes(sessionName string, windowIdx int, startDir string, win config.WindowConfig) error {
	if len(win.Panes) <= 1 {
		if len(win.Panes) == 1 && win.Panes[0].Cmd != "" {
			sendCommand(sessionName, windowIdx, 0, win.Panes[0].Cmd)
		}
		return nil
	}

	if win.Rows > 1 {
		return applyGridPanes(sessionName, windowIdx, startDir, win)
	}

	sizes := computeSizes(win.Panes)

	splitFlag := "-h"
	if win.Split == "vertical" {
		splitFlag = "-v"
	}

	for i := 1; i < len(win.Panes); i++ {
		remainingSum := 0
		for j := i; j < len(sizes); j++ {
			remainingSum += sizes[j]
		}
		currentSum := sizes[i-1] + remainingSum
		p := 50
		if currentSum > 0 {
			p = remainingSum * 100 / currentSum
		}

		target := fmt.Sprintf("=%s:%d", sessionName, windowIdx)
		if _, err := run("split-window", splitFlag, "-t", target, "-p", strconv.Itoa(p), "-c", startDir); err != nil {
			return fmt.Errorf("splitting pane %d in window %s: %w", i, win.Name, err)
		}
	}

	for i, pane := range win.Panes {
		if pane.Cmd != "" {
			sendCommand(sessionName, windowIdx, i, pane.Cmd)
		}
	}

	run("select-pane", "-t", fmt.Sprintf("=%s:%d.0", sessionName, windowIdx))

	return nil
}

// applyGridPanes creates a grid layout: first splits into columns (-h),
// then splits each column into rows (-v).
// Panes are laid out left-to-right, top-to-bottom.
func applyGridPanes(sessionName string, windowIdx int, startDir string, win config.WindowConfig) error {
	rows := win.Rows
	cols := (len(win.Panes) + rows - 1) / rows // ceil division

	winTarget := fmt.Sprintf("=%s:%d", sessionName, windowIdx)

	// Step 1: create columns by splitting horizontally.
	// After this we have pane indices 0..cols-1, each being one column.
	for c := 1; c < cols; c++ {
		remaining := cols - c
		total := remaining + 1 // current pane represents 1 + remaining columns
		p := remaining * 100 / total
		target := fmt.Sprintf("=%s:%d.0", sessionName, windowIdx)
		if _, err := run("split-window", "-h", "-t", target, "-p", strconv.Itoa(p), "-c", startDir); err != nil {
			return fmt.Errorf("grid: creating column %d in window %s: %w", c, win.Name, err)
		}
	}

	// Step 2: split each column pane vertically into rows.
	// We iterate columns in reverse order so that pane indices don't shift
	// as we add panes to earlier columns.
	// After step 1, pane indices are 0..cols-1 (one per column).
	for c := cols - 1; c >= 0; c-- {
		panesInCol := rows
		startIdx := c * rows
		if startIdx+panesInCol > len(win.Panes) {
			panesInCol = len(win.Panes) - startIdx
		}
		if panesInCol <= 1 {
			continue
		}

		// Split the column pane into rows
		colPane := fmt.Sprintf("=%s:%d.%d", sessionName, windowIdx, c)
		for r := 1; r < panesInCol; r++ {
			remaining := panesInCol - r
			total := remaining + 1
			p := remaining * 100 / total
			if _, err := run("split-window", "-v", "-t", colPane, "-p", strconv.Itoa(p), "-c", startDir); err != nil {
				return fmt.Errorf("grid: creating row %d in column %d of window %s: %w", r, c, win.Name, err)
			}
		}
	}

	// Step 3: send commands. Panes are now ordered by tmux left-to-right, top-to-bottom.
	// The pane ordering after grid creation: column 0 rows, column 1 rows, etc.
	for i, pane := range win.Panes {
		if pane.Cmd != "" {
			sendCommand(sessionName, windowIdx, i, pane.Cmd)
		}
	}

	run("select-pane", "-t", fmt.Sprintf("=%s:%d.0", sessionName, windowIdx))
	// Use tiled layout to even out the grid
	run("select-layout", "-t", winTarget, "tiled")

	return nil
}

func sendCommand(sessionName string, windowIdx, paneIdx int, cmd string) {
	target := fmt.Sprintf("=%s:%d.%d", sessionName, windowIdx, paneIdx)
	run("send-keys", "-t", target, "-l", cmd)
	run("send-keys", "-t", target, "Enter")
}

func computeSizes(panes []config.PaneConfig) []int {
	sizes := make([]int, len(panes))
	totalSpecified := 0
	unspecifiedCount := 0

	for i, p := range panes {
		if p.Size != "" {
			sizes[i] = parseSize(p.Size)
			totalSpecified += sizes[i]
		} else {
			unspecifiedCount++
		}
	}

	if unspecifiedCount > 0 {
		remaining := 100 - totalSpecified
		if remaining < 0 {
			remaining = 0
		}
		each := remaining / unspecifiedCount
		extra := remaining % unspecifiedCount
		for i := range sizes {
			if sizes[i] == 0 {
				sizes[i] = each
				if extra > 0 {
					sizes[i]++
					extra--
				}
			}
		}
	}

	return sizes
}

func parseSize(s string) int {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	n, _ := strconv.Atoi(s)
	return n
}
