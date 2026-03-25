package prompt

import (
	"fmt"
	"strings"

	"github.com/google/shlex"
)

// DoNothing will simply do nothing. It is the default
// OnStateChangeFunc.
func DoNothing(p *Prompt) {}

// ShellSplit will split the input on semi-colon and then use
// github.com/google/shlex to split the input in a bash-like style.
// This is the default SplitterFunc.
func ShellSplit(input string) ([][]string, error) {
	var cmd []string
	var cmds [][]string
	var e error

	for _, input := range strings.Split(input, ";") {
		if cmd, e = shlex.Split(input); e != nil {
			return nil, fmt.Errorf("failed to parse input: %w", e)
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// SimpleSplit will split the input on semi-colon and then on
// whitespace.
func SimpleSplit(input string) ([][]string, error) {
	var cmds [][]string

	for _, input := range strings.Split(input, ";") {
		cmds = append(cmds, strings.Fields(input))
	}

	return cmds, nil
}
