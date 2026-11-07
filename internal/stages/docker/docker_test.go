package docker

import "testing"

func TestParseSize(t *testing.T) {
	cases := map[string]int64{
		"0B":            0,
		"512MB":         512_000_000,
		"1.093GB (40%)": 1_093_000_000,
		"2.5GB":         2_500_000_000,
		"100MiB":        100 << 20,
		"garbage":       0,
	}
	for in, want := range cases {
		if got := ParseSize(in); got != want {
			t.Errorf("ParseSize(%q) = %d, want %d", in, got, want)
		}
	}
}
