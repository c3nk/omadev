package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Confirm asks a yes/no question that defaults to yes and returns the answer.
// An empty line, "y", or "yes" (any case) means yes; anything else, including "n"
// and "no", means no. Defaulting unrecognized input to no keeps state-changing
// actions from proceeding on accidental input.
func Confirm(in io.Reader, out io.Writer, question string) (bool, error) {
	fmt.Fprintf(out, "%s [Y/n] ", question)

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
