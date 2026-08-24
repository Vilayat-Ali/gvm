package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
	progressbar "github.com/schollz/progressbar/v3"
)

var (
	faint   = color.New(color.Faint)
	bold    = color.New(color.Bold)
	green   = color.New(color.FgGreen)
	yellow  = color.New(color.FgYellow)
	cyan    = color.New(color.FgCyan)
	magenta = color.New(color.FgMagenta)
)

// Narration goes to stderr so that it stays in one stream and keeps its
// order in CI logs. Only command results go to stdout, where they can be
// piped into other tools.

func heading(title string) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, cyan.Sprint("  "+title))
	fmt.Fprintln(os.Stderr, faint.Sprint("  "+strings.Repeat("-", len(title))))
}

func blank() {
	fmt.Fprintln(os.Stderr)
}

func plain(format string, args ...any) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, args...))
}

func accent(c *color.Color, format string, args ...any) {
	fmt.Fprintln(os.Stderr, c.Sprintf(format, args...))
}

func success(format string, args ...any) {
	fmt.Fprintln(os.Stderr, green.Sprintf("  "+format, args...))
}

func warn(format string, args ...any) {
	fmt.Fprintln(os.Stderr, yellow.Sprintf("  ! "+format, args...))
}

func hint(format string, args ...any) {
	fmt.Fprintln(os.Stderr, faint.Sprintf("  "+format, args...))
}

func out(format string, args ...any) {
	fmt.Println(fmt.Sprintf(format, args...))
}

func outColor(c *color.Color, format string, args ...any) {
	fmt.Println(c.Sprintf(format, args...))
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func newProgress(total int64, label string) io.Writer {
	if !color.NoColor && total > 0 {
		return progressbar.DefaultBytes(total, label)
	}
	return io.Discard
}
