package internal

import (
	"go/ast"
	"slices"
	"strings"
)

// hasNoLint returns true if we have //nolint or //nolint:errchecknext,
// or it's in the comma-separated list of ignored linters.
func hasNoLint(comment *ast.Comment) bool {
	text := strings.TrimPrefix(comment.Text, "//")

	if !strings.HasPrefix(text, "nolint") {
		return false
	}

	parts := strings.SplitN(text, ":", 2)
	if len(parts) == 1 {
		// bare nolint comment
		return true
	}

	// contains our linter name
	return slices.Contains(strings.Split(parts[1], ","), name)
}
