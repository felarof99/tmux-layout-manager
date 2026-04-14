package tmux

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type SplitSpec struct {
	Name             string
	Description      string
	UseCurrentWindow bool
	Rows             int
	Cols             int
}

var splitAliases = map[string]string{
	"2cols":   "12",
	"a22":     "22",
	"c":       "c22",
	"current": "c22",
	"ca22":    "c22",
}

func SplitExamples() []SplitSpec {
	names := []string{"12", "22", "23", "c22", "c23", "34", "c3x4"}
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

	useCurrentWindow := false
	switch {
	case strings.HasPrefix(token, "c"):
		useCurrentWindow = true
		token = strings.TrimPrefix(token, "c")
	case strings.HasPrefix(token, "a"):
		token = strings.TrimPrefix(token, "a")
	}

	rows, cols, ok := parseSplitDimensions(token)
	if !ok {
		return SplitSpec{}, false
	}

	return SplitSpec{
		Name:             canonicalSplitSpecName(useCurrentWindow, rows, cols),
		Description:      describeSplitSpec(useCurrentWindow, rows, cols),
		UseCurrentWindow: useCurrentWindow,
		Rows:             rows,
		Cols:             cols,
	}, true
}

func ApplySplitSpec(windowTarget, paneTarget, startDir string, spec SplitSpec) error {
	if spec.UseCurrentWindow {
		return reshapeWindowToGrid(windowTarget, paneTarget, startDir, spec.Rows, spec.Cols)
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

func canonicalSplitSpecName(useCurrentWindow bool, rows, cols int) string {
	base := fmt.Sprintf("%dx%d", rows, cols)
	if rows < 10 && cols < 10 {
		base = fmt.Sprintf("%d%d", rows, cols)
	}
	if useCurrentWindow {
		return "c" + base
	}
	return base
}

func describeSplitSpec(useCurrentWindow bool, rows, cols int) string {
	if useCurrentWindow {
		return fmt.Sprintf("Current window to %s", gridDimensions(rows, cols))
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

type windowPane struct {
	ID    string
	Left  int
	Top   int
	Index int
}

func reshapeWindowToGrid(windowTarget, paneTarget, startDir string, rows, cols int) error {
	targetPaneCount := rows * cols
	if targetPaneCount < 1 {
		return fmt.Errorf("grid must have at least 1 row and 1 column")
	}

	zoomed, err := CurrentWindowZoomed()
	if err != nil {
		return fmt.Errorf("checking zoom state: %w", err)
	}
	if zoomed {
		if _, err := run("resize-pane", "-t", paneTarget, "-Z"); err != nil {
			return fmt.Errorf("restoring zoom before relayout: %w", err)
		}
	}

	panes, err := listWindowPanes(windowTarget)
	if err != nil {
		return err
	}
	if len(panes) == 0 {
		return fmt.Errorf("current window has no panes")
	}
	if len(panes) > targetPaneCount {
		return fmt.Errorf("current window already has %d panes, cannot fit them into %d rows and %d columns", len(panes), rows, cols)
	}

	panes = orderExistingPanes(panes, paneTarget)

	stagingWindows := make([]string, 0, len(panes)-1)
	defer func() {
		for _, stagingWindow := range stagingWindows {
			run("kill-window", "-t", stagingWindow)
		}
	}()

	for _, pane := range panes[1:] {
		stagingWindow, err := stagePane(pane.ID)
		if err != nil {
			return fmt.Errorf("staging pane %s: %w", pane.ID, err)
		}
		stagingWindows = append(stagingWindows, stagingWindow)
	}

	if err := splitPaneGrid(paneTarget, startDir, rows, cols); err != nil {
		return err
	}

	targetPanes, err := listWindowPanes(windowTarget)
	if err != nil {
		return err
	}
	sortWindowPanes(targetPanes)
	if len(targetPanes) != targetPaneCount {
		return fmt.Errorf("expected %d panes after relayout, got %d", targetPaneCount, len(targetPanes))
	}

	for i, pane := range panes[1:] {
		if _, err := run("swap-pane", "-d", "-s", pane.ID, "-t", targetPanes[i+1].ID); err != nil {
			return fmt.Errorf("placing pane %s into grid cell %d: %w", pane.ID, i+2, err)
		}
	}

	run("select-pane", "-t", paneTarget)
	return nil
}

func listWindowPanes(windowTarget string) ([]windowPane, error) {
	out, err := run("list-panes", "-t", "="+windowTarget, "-F", "#{pane_id}\t#{pane_index}\t#{pane_left}\t#{pane_top}")
	if err != nil {
		return nil, fmt.Errorf("listing panes in current window: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	panes := make([]windowPane, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			return nil, fmt.Errorf("unexpected pane metadata %q", line)
		}
		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("parsing pane index from %q: %w", line, err)
		}
		left, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("parsing pane left from %q: %w", line, err)
		}
		top, err := strconv.Atoi(fields[3])
		if err != nil {
			return nil, fmt.Errorf("parsing pane top from %q: %w", line, err)
		}
		panes = append(panes, windowPane{
			ID:    fields[0],
			Index: index,
			Left:  left,
			Top:   top,
		})
	}

	sortWindowPanes(panes)
	return panes, nil
}

func orderExistingPanes(panes []windowPane, paneTarget string) []windowPane {
	if len(panes) == 0 {
		return nil
	}

	ordered := make([]windowPane, len(panes))
	copy(ordered, panes)
	sortWindowPanes(ordered)

	out := make([]windowPane, 0, len(panes))
	for _, pane := range ordered {
		if pane.ID == paneTarget {
			out = append(out, pane)
			break
		}
	}
	for _, pane := range ordered {
		if pane.ID == paneTarget {
			continue
		}
		out = append(out, pane)
	}
	return out
}

func sortWindowPanes(panes []windowPane) {
	sort.Slice(panes, func(i, j int) bool {
		if panes[i].Top != panes[j].Top {
			return panes[i].Top < panes[j].Top
		}
		if panes[i].Left != panes[j].Left {
			return panes[i].Left < panes[j].Left
		}
		return panes[i].Index < panes[j].Index
	})
}

func stagePane(paneTarget string) (string, error) {
	out, err := run("break-pane", "-d", "-s", paneTarget, "-P", "-F", "#{window_id}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
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
