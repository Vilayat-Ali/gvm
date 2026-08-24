package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const LatestAlias = "latest"

var versionPattern = regexp.MustCompile(`^go(\d{1,4})\.(\d{1,4})(?:\.(\d{1,4}))?(?:(alpha|beta|rc)(\d{1,4}))?$`)

type ParsedVersion struct {
	Major  int
	Minor  int
	Patch  int
	PreTag string
	PreNum int
}

func CanonicalVersion(input string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	normalized = strings.TrimPrefix(normalized, "v")
	if !strings.HasPrefix(normalized, "go") {
		normalized = "go" + normalized
	}
	if !versionPattern.MatchString(normalized) {
		return "", fmt.Errorf("%q is not a valid Go version (expected forms: 1.22.0, go1.22.0, 1.23rc1)", input)
	}
	return normalized, nil
}

func IsValidVersion(input string) bool {
	_, err := CanonicalVersion(input)
	return err == nil
}

func ParseVersion(input string) (ParsedVersion, error) {
	canonical, err := CanonicalVersion(input)
	if err != nil {
		return ParsedVersion{}, err
	}
	m := versionPattern.FindStringSubmatch(canonical)
	parsed := ParsedVersion{PreTag: m[4]}
	parsed.Major, _ = strconv.Atoi(m[1])
	parsed.Minor, _ = strconv.Atoi(m[2])
	if m[3] != "" {
		parsed.Patch, _ = strconv.Atoi(m[3])
	}
	if m[5] != "" {
		parsed.PreNum, _ = strconv.Atoi(m[5])
	}
	return parsed, nil
}

func (p ParsedVersion) IsPrerelease() bool {
	return p.PreTag != ""
}

func (p ParsedVersion) preRank() int {
	switch p.PreTag {
	case "alpha":
		return 0
	case "beta":
		return 1
	case "rc":
		return 2
	default:
		return 3
	}
}

func CompareVersions(a, b ParsedVersion) int {
	fields := [][2]int{
		{a.Major, b.Major},
		{a.Minor, b.Minor},
		{a.Patch, b.Patch},
		{a.preRank(), b.preRank()},
		{a.PreNum, b.PreNum},
	}
	for _, f := range fields {
		if f[0] != f[1] {
			if f[0] < f[1] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func DisplayVersion(version string) string {
	return strings.TrimPrefix(version, "go")
}
