package tmux

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type SplitSpec struct {
	Name        string
	Description string
	KeepCurrent bool
	Rows        int
	Cols        int
}

var splitAliases = map[string]string{
	"2cols":   "12",
	"a22":     "22",
	"c":       "c22",
	"current": "c22",
	"ca22":    "c22",
}

func SplitExamples() []SplitSpec {
	names := []string{"12", "22", "c12", "c22", "34", "c3x4"}
	out := make([]SplitSpec, 0, len(names))
	for _, name := range names {
		spec, ok := ParseSplitSpec(name)
		if ok {
			out = append(out, spec)
		}
	}
	return out
}

func ParseSplitSpec(name string) (SplitSpec, bool) {
	token := strings.ToLower(strings.TrimSpace(name))
	if token == "" {
		return SplitSpec{}, false
	}

	if alias, ok := splitAliases[token]; ok {
		token = alias
	}

	keepCurrent := false
	switch {
	case strings.HasPrefix(token, "c"):
		keepCurrent = true
		token = strings.TrimPrefix(token, "c")
	case strings.HasPrefix(token, "a"):
		token = strings.TrimPrefix(token, "a")
	}

	rows, cols, ok := parseSplitDimensions(token)
	if !ok {
		return SplitSpec{}, false
	}

	return SplitSpec{
		Name:        canonicalSplitSpecName(keepCurrent, rows, cols),
		Description: describeSplitSpec(keepCurrent, rows, cols),
		KeepCurrent: keepCurrent,
		Rows:        rows,
		Cols:        cols,
	}, true
}

func ApplySplitSpec(paneTarget, startDir string, spec SplitSpec) error {
	if spec.KeepCurrent {
		return splitPaneKeepCurrentAddGrid(paneTarget, startDir, spec.Rows, spec.Cols)
	}
	return splitPaneGrid(paneTarget, startDir, spec.Rows, spec.Cols)
}

func parseSplitDimensions(token string) (int, int, bool) {
	if token == "" {
		return 0, 0, false
	}

	if strings.Contains(token, "x") {
		parts := strings.Split(token, "x")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return 0, 0, false
		}
		rows, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, false
		}
		cols, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, false
		}
		if rows < 1 || cols < 1 {
			return 0, 0, false
		}
		return rows, cols, true
	}

	if len(token) != 2 {
		return 0, 0, false
	}
	for _, r := range token {
		if !unicode.IsDigit(r) {
			return 0, 0, false
		}
	}

	rows := int(token[0] - '0')
	cols := int(token[1] - '0')
	if rows < 1 || cols < 1 {
		return 0, 0, false
	}
	return rows, cols, true
}

func canonicalSplitSpecName(keepCurrent bool, rows, cols int) string {
	base := fmt.Sprintf("%dx%d", rows, cols)
	if rows < 10 && cols < 10 {
		base = fmt.Sprintf("%d%d", rows, cols)
	}
	if keepCurrent {
		return "c" + base
	}
	return base
}

func describeSplitSpec(keepCurrent bool, rows, cols int) string {
	if keepCurrent {
		return fmt.Sprintf("Keep current pane, add %s below", gridDimensions(rows, cols))
	}
	return gridDimensions(rows, cols)
}

func gridDimensions(rows, cols int) string {
	return fmt.Sprintf("%d %s, %d %s", rows, pluralize(rows, "row"), cols, pluralize(cols, "column"))
}

func pluralize(n int, singular string) string {
	if n == 1 {
		return singular
	}
	return singular + "s"
}

func splitPaneKeepCurrentAddGrid(paneTarget, startDir string, rows, cols int) error {
	if rows < 1 || cols < 1 {
		return fmt.Errorf("grid must have at least 1 row and 1 column")
	}

	bottomPercent := rows * 100 / (rows + 1)
	bottomPane, err := splitPane(paneTarget, "-v", bottomPercent, startDir)
	if err != nil {
		return err
	}

	if err := splitPaneGrid(bottomPane, startDir, rows, cols); err != nil {
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
