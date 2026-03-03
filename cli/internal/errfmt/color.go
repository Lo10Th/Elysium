package errfmt

import (
	"os"
)

var (
	colorEnabled = true
)

func init() {
	if os.Getenv("NO_COLOR") != "" {
		colorEnabled = false
	}
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

func colorize(text, color string) string {
	if !colorEnabled {
		return text
	}
	return color + text + colorReset
}
