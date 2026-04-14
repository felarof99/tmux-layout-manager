package tmux

import "testing"

func TestParseSplitSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		wantName    string
		wantKeep    bool
		wantRows    int
		wantCols    int
		description string
		ok          bool
	}{
		{name: "12", wantName: "12", wantRows: 1, wantCols: 2, description: "1 row, 2 columns", ok: true},
		{name: "2cols", wantName: "12", wantRows: 1, wantCols: 2, description: "1 row, 2 columns", ok: true},
		{name: "a22", wantName: "22", wantRows: 2, wantCols: 2, description: "2 rows, 2 columns", ok: true},
		{name: "34", wantName: "34", wantRows: 3, wantCols: 4, description: "3 rows, 4 columns", ok: true},
		{name: "3x4", wantName: "34", wantRows: 3, wantCols: 4, description: "3 rows, 4 columns", ok: true},
		{name: "c", wantName: "c22", wantKeep: true, wantRows: 2, wantCols: 2, description: "Keep current pane, add 2 rows, 2 columns below", ok: true},
		{name: "current", wantName: "c22", wantKeep: true, wantRows: 2, wantCols: 2, description: "Keep current pane, add 2 rows, 2 columns below", ok: true},
		{name: "c12", wantName: "c12", wantKeep: true, wantRows: 1, wantCols: 2, description: "Keep current pane, add 1 row, 2 columns below", ok: true},
		{name: "c22", wantName: "c22", wantKeep: true, wantRows: 2, wantCols: 2, description: "Keep current pane, add 2 rows, 2 columns below", ok: true},
		{name: "c3x4", wantName: "c34", wantKeep: true, wantRows: 3, wantCols: 4, description: "Keep current pane, add 3 rows, 4 columns below", ok: true},
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
			if got.KeepCurrent != tt.wantKeep {
				t.Fatalf("ParseSplitSpec(%q) KeepCurrent = %v, want %v", tt.name, got.KeepCurrent, tt.wantKeep)
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
	wantNames := []string{"12", "22", "c12", "c22", "34", "c34"}

	if len(got) != len(wantNames) {
		t.Fatalf("SplitExamples() len = %d, want %d", len(got), len(wantNames))
	}

	for i, spec := range got {
		if spec.Name != wantNames[i] {
			t.Fatalf("SplitExamples()[%d].Name = %q, want %q", i, spec.Name, wantNames[i])
		}
	}
}
