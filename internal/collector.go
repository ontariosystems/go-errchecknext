package internal

import (
	"fmt"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/phayes/checkstyle"
)

// Collector thread-safe struct for collecting findings from an analysis
type Collector struct {
	findings []finding
	sync.RWMutex
}

type finding struct {
	filename string
	line     int
	column   int
	message  string
}

// NewCollector returns a reference to a new empty Collector
func NewCollector() *Collector {
	return &Collector{}
}

// Print outputs the findings in a Collector to an output stream
func (c *Collector) Print(out io.Writer) {
	c.RLock()
	defer c.RUnlock()

	for _, f := range c.findings {
		_, _ = fmt.Fprintf(out, "%s:%d:%d %s\n", f.filename, f.line, f.column, f.message)
	}
}

// WriteCheckstyle outputs the findings to an outputPath in Checkstyle format
func (c *Collector) WriteCheckstyle(outputPath string) error {
	c.RLock()
	defer c.RUnlock()

	report := checkstyle.New()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	for _, f := range c.findings {
		path := f.filename
		if relPath, err := filepath.Rel(wd, path); err == nil {
			path = relPath
		}

		fileReport := report.EnsureFile(path)
		fileReport.AddError(checkstyle.NewError(
			f.line,
			f.column,
			checkstyle.SeverityWarning,
			f.message,
			name,
		))
	}

	if err := os.WriteFile(outputPath, []byte(report.String()), 0o600); err != nil {
		return err
	}

	return nil
}

// HasErrors returns true if there are any findings
func (c *Collector) HasErrors() bool {
	c.RLock()
	defer c.RUnlock()

	return len(c.findings) > 0
}

// ReportFinding is called by the analyzer when it is scanning the code. needs to be thread-safe
func (c *Collector) ReportFinding(fileSet *token.FileSet, pos token.Pos, message string) {
	c.Lock()
	defer c.Unlock()

	filePos := fileSet.Position(pos)
	c.findings = append(c.findings, finding{
		filename: filePos.Filename,
		line:     filePos.Line,
		column:   filePos.Column,
		message:  message,
	})
}
