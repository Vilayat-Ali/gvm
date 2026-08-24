package internal

import "testing"

func TestCanonicalVersion(t *testing.T) {
	valid := map[string]string{
		"1.22.0":     "go1.22.0",
		"v1.22.0":    "go1.22.0",
		"go1.22.0":   "go1.22.0",
		"  1.22.0 ":  "go1.22.0",
		"1.23":       "go1.23",
		"1.23rc1":    "go1.23rc1",
		"GO1.24.2":   "go1.24.2",
		"1.24beta1":  "go1.24beta1",
		"go1.22.0\n": "go1.22.0",
	}
	for input, want := range valid {
		got, err := CanonicalVersion(input)
		if err != nil {
			t.Fatalf("CanonicalVersion(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Errorf("CanonicalVersion(%q) = %q, want %q", input, got, want)
		}
	}

	invalid := []string{
		"", "latest", "1", "1.2.3.4", "../../etc/passwd", "go1.22.0/../..",
		"$(whoami)", "1.22.0; rm -rf /", "go 1.22.0", "1.-1", "abc",
		"1.22.0/", "/go1.22.0", "go99999.1.1",
	}
	for _, input := range invalid {
		if _, err := CanonicalVersion(input); err == nil {
			t.Errorf("CanonicalVersion(%q) accepted an invalid version", input)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	ordered := []string{
		"1.21.0", "1.21.5", "1.22beta1", "1.22rc1", "1.22rc2", "1.22", "1.22.1", "1.23.0", "2.0.0",
	}
	for i := 1; i < len(ordered); i++ {
		lower, err := ParseVersion(ordered[i-1])
		if err != nil {
			t.Fatal(err)
		}
		higher, err := ParseVersion(ordered[i])
		if err != nil {
			t.Fatal(err)
		}
		if CompareVersions(lower, higher) >= 0 {
			t.Errorf("expected %s < %s", ordered[i-1], ordered[i])
		}
		if CompareVersions(higher, lower) <= 0 {
			t.Errorf("expected %s > %s", ordered[i], ordered[i-1])
		}
	}

	same, _ := ParseVersion("1.22.0")
	if CompareVersions(same, same) != 0 {
		t.Error("expected a version to equal itself")
	}
}

func TestSortVersionsDesc(t *testing.T) {
	versions := []RemoteVersion{
		{Version: "go1.21.0"},
		{Version: "go1.23.1"},
		{Version: "go1.22rc1"},
		{Version: "go1.22.0"},
		{Version: "go1.9.7"},
	}
	SortVersionsDesc(versions)

	want := []string{"go1.23.1", "go1.22.0", "go1.22rc1", "go1.21.0", "go1.9.7"}
	for i, expected := range want {
		if versions[i].Version != expected {
			t.Fatalf("position %d = %s, want %s", i, versions[i].Version, expected)
		}
	}
}

func TestPrereleaseDetection(t *testing.T) {
	for _, input := range []string{"1.22rc1", "1.22beta1", "1.22alpha1"} {
		parsed, err := ParseVersion(input)
		if err != nil {
			t.Fatal(err)
		}
		if !parsed.IsPrerelease() {
			t.Errorf("%s should be a pre-release", input)
		}
	}
	parsed, _ := ParseVersion("1.22.0")
	if parsed.IsPrerelease() {
		t.Error("1.22.0 should not be a pre-release")
	}
}
