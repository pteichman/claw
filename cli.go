package claw

import (
	"fmt"
	"io"
)

func CLI(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintf(stderr, "ERROR: No command specified\n")
		return 2
	}

	cmd, rest := args[0], args[1:]

	switch cmd {
	case "links":
		return cmdLinks(rest, stdout, stderr)

	case "unlinked":
		return cmdUnlinked(rest, stdout, stderr)

	default:
		fmt.Fprintf(stderr, "ERROR: Unknown command: %s\n", cmd)
		return 2
	}
}
