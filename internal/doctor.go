package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var goVersionPattern = regexp.MustCompile(`go version go(\S+)`)

type Diagnosis struct {
	Root       string
	ShimDir    string
	Current    string
	CurrentErr error
	ResolvedGo string
	ActiveGo   string
	OnPath     bool
	Shadowed   bool
	ShadowedBy string
	GoRootEnv  string
	Toolchain  string
	ModuleFile string
	ModuleGo   string
	Installed  int
	Problems   []string
	Warnings   []string
	Hints      []string
}

func pathEntries() []string {
	raw := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	entries := make([]string, 0, len(raw))
	for _, entry := range raw {
		if entry == "" {
			continue
		}
		entries = append(entries, filepath.Clean(entry))
	}
	return entries
}

func goVersionOf(binary string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := probeCommand(ctx, binary, "").Output()
	if err != nil {
		return ""
	}
	matches := goVersionPattern.FindStringSubmatch(string(out))
	if len(matches) < 2 {
		return ""
	}
	return "go" + matches[1]
}

func ActiveGoVersion() string {
	binary, err := exec.LookPath("go")
	if err != nil {
		return ""
	}
	return goVersionOf(binary)
}

func findGoMod(start string) string {
	dir := filepath.Clean(start)
	for {
		candidate := filepath.Join(dir, "go.mod")
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func goDirectiveOf(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, raw := range strings.Split(string(data), "\n") {
		fields := strings.Fields(raw)
		if len(fields) < 2 || fields[0] != "go" {
			continue
		}
		if canonical, err := CanonicalVersion(fields[1]); err == nil {
			return canonical
		}
	}
	return ""
}

func Diagnose() Diagnosis {
	var d Diagnosis

	root, err := Root()
	if err != nil {
		d.Problems = append(d.Problems, err.Error())
		return d
	}
	d.Root = root
	d.ShimDir, _ = ShimDir()
	d.Current, d.CurrentErr = CurrentVersion()
	d.GoRootEnv = os.Getenv("GOROOT")

	if installed, err := InstalledVersions(); err == nil {
		d.Installed = len(installed)
	}

	entries := pathEntries()
	shimIndex := -1
	for i, entry := range entries {
		if entry == filepath.Clean(d.ShimDir) {
			shimIndex = i
			break
		}
	}
	d.OnPath = shimIndex >= 0

	if resolved, err := exec.LookPath("go"); err == nil {
		d.ResolvedGo, _ = filepath.Abs(resolved)
		d.ActiveGo = goVersionOf(resolved)
		owner := filepath.Clean(filepath.Dir(d.ResolvedGo))
		if d.OnPath && owner != filepath.Clean(d.ShimDir) {
			d.Shadowed = true
			d.ShadowedBy = owner
		}
	}

	if d.CurrentErr != nil && d.Installed > 0 {
		d.Problems = append(d.Problems, "no Go version is active. Run `gvm use <version>`")
	}
	if d.Installed == 0 {
		d.Hints = append(d.Hints, "no Go versions installed yet. Run `gvm use latest`")
	}
	if !d.OnPath {
		d.Problems = append(d.Problems, fmt.Sprintf("%s is not on your PATH", d.ShimDir))
		d.Hints = append(d.Hints, "add this to your shell profile: "+ShellExportLine(d.ShimDir))
	}
	if d.Shadowed {
		d.Problems = append(d.Problems, fmt.Sprintf("another Go in %s takes priority over gvm", d.ShadowedBy))
		d.Hints = append(d.Hints, fmt.Sprintf("put %s before %s in your PATH", d.ShimDir, d.ShadowedBy))
	}
	if d.GoRootEnv != "" {
		d.Warnings = append(d.Warnings, fmt.Sprintf("GOROOT is pinned to %s in your environment, which overrides gvm", d.GoRootEnv))
		d.Hints = append(d.Hints, "unset GOROOT in your shell profile")
	}
	if warning := RootWarning(); warning != "" {
		d.Warnings = append(d.Warnings, warning)
	}

	d.Toolchain = os.Getenv("GOTOOLCHAIN")
	if d.Toolchain == "" {
		d.Toolchain = "auto"
	}
	if cwd, err := os.Getwd(); err == nil {
		if mod := findGoMod(cwd); mod != "" {
			d.ModuleFile = mod
			d.ModuleGo = goDirectiveOf(mod)
		}
	}
	if switchesToolchain(d.Current, d.ModuleGo, d.Toolchain) {
		d.Warnings = append(d.Warnings, fmt.Sprintf(
			"%s requires Go %s, which is newer than the active %s — Go will download and run %s here instead (GOTOOLCHAIN=%s)",
			d.ModuleFile, DisplayVersion(d.ModuleGo), DisplayVersion(d.Current), DisplayVersion(d.ModuleGo), d.Toolchain))
		d.Hints = append(d.Hints, fmt.Sprintf("run `gvm use %s`, or set GOTOOLCHAIN=local to make Go fail instead of switching", DisplayVersion(d.ModuleGo)))
	}

	return d
}

func switchesToolchain(active, moduleGo, toolchain string) bool {
	if active == "" || moduleGo == "" || toolchain == "local" {
		return false
	}
	current, err := ParseVersion(active)
	if err != nil {
		return false
	}
	required, err := ParseVersion(moduleGo)
	if err != nil {
		return false
	}
	return CompareVersions(current, required) < 0
}

func ShellExportLine(shimDir string) string {
	return fmt.Sprintf("export PATH=%q:$PATH", shimDir)
}

func (d Diagnosis) Healthy() bool {
	return len(d.Problems) == 0
}
