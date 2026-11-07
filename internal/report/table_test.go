package report

import "testing"

func TestHumanize(t *testing.T) {
	cases := map[int64]string{
		0:          "0 B",
		512:        "512 B",
		1024:       "1.0 KB",
		213038015:  "203.2 MB",
		7516192768: "7.0 GB",
	}
	for in, want := range cases {
		if got := Humanize(in); got != want {
			t.Errorf("Humanize(%d) = %q, want %q", in, got, want)
		}
	}
}
