package tmux

import "testing"

func TestParseSplitSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		wantName    string
		wantWindow  bool
		wantRows    int
		wantCols    int
		description string
		ok          bool
	}{
		{name: "12", wantName: "12", wantRows: 1, wantCols: 2, description: "1 row, 2 columns", ok: true},
		{name: "2cols", wantName: "12", wantRows: 1, wantCols: 2, description: "1 row, 2 columns", ok: true},
		{name: "a22", wantName: "22", wantRows: 2, wantCols: 2, description: "2 rows, 2 columns", ok: true},
		{name: "23", wantName: "23", wantRows: 2, wantCols: 3, description: "2 rows, 3 columns", ok: true},
		{name: "34", wantName: "34", wantRows: 3, wantCols: 4, description: "3 rows, 4 columns", ok: true},
		{name: "3x4", wantName: "34", wantRows: 3, wantCols: 4, description: "3 rows, 4 columns", ok: true},
		{name: "c", wantName: "c22", wantWindow: true, wantRows: 2, wantCols: 2, description: "Current window to 2 rows, 2 columns", ok: true},
		{name: "current", wantName: "c22", wantWindow: true, wantRows: 2, wantCols: 2, description: "Current window to 2 rows, 2 columns", ok: true},
		{name: "c12", wantName: "c12", wantWindow: true, wantRows: 1, wantCols: 2, description: "Current window to 1 row, 2 columns", ok: true},
		{name: "c22", wantName: "c22", wantWindow: true, wantRows: 2, wantCols: 2, description: "Current window to 2 rows, 2 columns", ok: true},
		{name: "c23", wantName: "c23", wantWindow: true, wantRows: 2, wantCols: 3, description: "Current window to 2 rows, 3 columns", ok: true},
		{name: "c3x4", wantName: "c34", wantWindow: true, wantRows: 3, wantCols: 4, description: "Current window to 3 rows, 4 columns", ok: true},
		{name: "missing", ok: false},
		{name: "0x4", ok: false},
		{name: "c00", ok: false},
		{name: "123", ok: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := ParseSplitSpec(tt.name)
			if ok != tt.ok {
				t.Fatalf("ParseSplitSpec(%q) ok = %v, want %v", tt.name, ok, tt.ok)
			}
			if !tt.ok {
				return
			}
			if got.Name != tt.wantName {
				t.Fatalf("ParseSplitSpec(%q) name = %q, want %q", tt.name, got.Name, tt.wantName)
			}
			if got.UseCurrentWindow != tt.wantWindow {
				t.Fatalf("ParseSplitSpec(%q) UseCurrentWindow = %v, want %v", tt.name, got.UseCurrentWindow, tt.wantWindow)
			}
			if got.Rows != tt.wantRows || got.Cols != tt.wantCols {
				t.Fatalf("ParseSplitSpec(%q) rows/cols = %d/%d, want %d/%d", tt.name, got.Rows, got.Cols, tt.wantRows, tt.wantCols)
			}
			if got.Description != tt.description {
				t.Fatalf("ParseSplitSpec(%q) description = %q, want %q", tt.name, got.Description, tt.description)
			}
		})
	}
}

func TestSplitExamples(t *testing.T) {
	t.Parallel()

	got := SplitExamples()
	wantNames := []string{"12", "22", "23", "c22", "c23", "34", "c34"}

	if len(got) != len(wantNames) {
		t.Fatalf("SplitExamples() len = %d, want %d", len(got), len(wantNames))
	}

	for i, spec := range got {
		if spec.Name != wantNames[i] {
			t.Fatalf("SplitExamples()[%d].Name = %q, want %q", i, spec.Name, wantNames[i])
		}
	}
}

func TestOrderExistingPanes(t *testing.T) {
	t.Parallel()

	panes := []windowPane{
		{ID: "%2", Index: 2, Left: 50, Top: 0},
		{ID: "%1", Index: 1, Left: 0, Top: 0},
		{ID: "%3", Index: 3, Left: 0, Top: 20},
	}

	got := orderExistingPanes(panes, "%3")
	want := []string{"%3", "%1", "%2"}

	if len(got) != len(want) {
		t.Fatalf("orderExistingPanes() len = %d, want %d", len(got), len(want))
	}

	for i, pane := range got {
		if pane.ID != want[i] {
			t.Fatalf("orderExistingPanes()[%d] = %q, want %q", i, pane.ID, want[i])
		}
	}
}
