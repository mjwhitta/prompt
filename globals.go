package prompt

import "regexp"

// Version is the package version
const Version string = "1.2.2"

var reEscCodes = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
