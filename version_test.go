// Package selfupdate's own test package is used (not selfupdate_test) per
// the task spec, so these tests can be extended later to exercise unexported
// helpers without a second test package.
//
//nolint:testpackage // internal package per task spec envctl-04qa.00bow5.1.1.2
package selfupdate

import "testing"

// normV010 and normRC1 name the normalized forms that both a bare and a
// v-prefixed input map to below. Naming them avoids repeating the string
// literal enough times to trip goconst.
const (
	normV010 = "v0.1.0"
	normRC1  = "v1.2.3-rc.1"
)

func TestNormalizeVersion(t *testing.T) {
	cases := map[string]string{
		"":           "v0.0.0-dev",
		"dev":        "v0.0.0-dev",
		"0.1.0":      normV010,
		"v0.1.0":     normV010,
		"1.2.3-rc.1": normRC1,
		normRC1:      normRC1,
	}
	for in, want := range cases {
		if got := NormalizeVersion(in); got != want {
			t.Errorf("NormalizeVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCompare(t *testing.T) {
	cases := []struct {
		current, latest string
		sign            int
	}{
		{"dev", "v1.0.0", 1},
		{"1.2.3", "v1.2.3", 0},
		{"v1.3.0-rc.1", "v1.3.0", 1},
		{"v2.0.0", "v1.0.0", -1},
		{"v1.3.0-rc.1", "v1.3.0-rc.2", 1},
		{"v1.3.0", "v1.3.0-rc.9", -1},
		{"v0.6.0", "v0.6.1-nightly.20260904", 1},
		{"v0.6.1-nightly.20260904", "v0.6.1", 1},
	}
	for _, c := range cases {
		got := Compare(c.current, c.latest)
		if sign(got) != c.sign {
			t.Errorf("Compare(%q, %q) = %d, want sign %d", c.current, c.latest, got, c.sign)
		}
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
