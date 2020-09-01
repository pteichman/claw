package main

import (
	"os"

	"github.com/pteichman/claw"
)

func main() {
	os.Exit(claw.CLI(os.Args[1:], os.Stdout, os.Stderr))
}
