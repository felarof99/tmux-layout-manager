package tmux

import "testing"

func TestFindSplitPreset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
		ok   bool
	}{
		{name: "2cols", want: "2cols", ok: true},
		{name: "12", want: "2cols", ok: true},
		{name: "c12", want: "2cols", ok: true},
		{name: "c", want: "c", ok: true},
		{name: "current", want: "c", ok: true},
		{name: "ca22", want: "c", ok: true},
		{name: "a22", want: "a22", ok: true},
		{name: "22", want: "a22", ok: true},
		{name: "c22", want: "a22", ok: true},
		{name: "missing", ok: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := FindSplitPreset(tt.name)
			if ok != tt.ok {
				t.Fatalf("FindSplitPreset(%q) ok = %v, want %v", tt.name, ok, tt.ok)
			}
			if !tt.ok {
				return
			}
			if got.Name != tt.want {
				t.Fatalf("FindSplitPreset(%q) name = %q, want %q", tt.name, got.Name, tt.want)
			}
		})
	}
}

func TestSplitPresets(t *testing.T) {
	t.Parallel()

	got := SplitPresets()
	if len(got) != 3 {
		t.Fatalf("SplitPresets() len = %d, want 3", len(got))
	}

	want := []struct {
		name        string
		description string
	}{
		{name: "2cols", description: "1 row, 2 columns"},
		{name: "c", description: "Additional 2 rows, 2 columns at the bottom"},
		{name: "a22", description: "2 rows, 2 columns"},
	}

	for i, preset := range got {
		if preset.Name != want[i].name {
			t.Fatalf("SplitPresets()[%d].Name = %q, want %q", i, preset.Name, want[i].name)
		}
		if preset.Description != want[i].description {
			t.Fatalf("SplitPresets()[%d].Description = %q, want %q", i, preset.Description, want[i].description)
		}
	}
}
