package safety

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Interactive reports whether stdin is a terminal we can prompt on.
func Interactive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

var reader = bufio.NewReader(os.Stdin)

// Ask prompts a yes/no question that defaults to No. assumeYes (from --yes)
// answers yes without prompting, for non-interactive use.
func Ask(prompt string, assumeYes bool) bool {
	if assumeYes {
		return true
	}
	if !Interactive() {
		return false
	}
	fmt.Printf("%s [y/N]: ", prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
