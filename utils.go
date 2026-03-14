package prompt

import (
	"fmt"
	"os"
)

func plain(s string) string {
	return reEscCodes.ReplaceAllString(s, "")
}

func readByte() (byte, error) {
	var b []byte = make([]byte, 1)

	if _, e := os.Stdin.Read(b); e != nil {
		//nolint:mnd // NUL
		return 0x0, fmt.Errorf("failed to read from stdin: %w", e)
	}

	return b[0], nil
}
