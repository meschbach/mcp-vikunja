package vikunja

import (
	"bytes"
	"io"
)

// Formatter handles output formatting for CLI
type Formatter struct {
	useColor bool
	output   io.Writer
}

// NewFormatter creates a new formatter
func NewFormatter(useColor bool, output io.Writer) *Formatter {
	return &Formatter{
		useColor: useColor,
		output:   output,
	}
}

// CaptureOutput captures formatted output to a string
func (f *Formatter) CaptureOutput(fn func() error) (string, error) {
	buf := &bytes.Buffer{}
	oldOutput := f.output
	f.output = buf
	defer func() { f.output = oldOutput }()

	if err := fn(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
