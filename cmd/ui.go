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

func heading(title string) {
	fmt.Println()
	fmt.Println(cyan.Sprint("  " + title))
	fmt.Println(faint.Sprint("  " + strings.Repeat("-", len(title))))
}

func plain(format string, args ...any) {
	fmt.Println(fmt.Sprintf(format, args...))
}

func accent(c *color.Color, format string, args ...any) {
	fmt.Println(c.Sprintf(format, args...))
}

func success(format string, args ...any) {
	fmt.Println(green.Sprintf("  "+format, args...))
}

func warn(format string, args ...any) {
	fmt.Fprintln(os.Stderr, yellow.Sprintf("  ! "+format, args...))
}

func hint(format string, args ...any) {
	fmt.Fprintln(os.Stderr, faint.Sprintf("  "+format, args...))
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
