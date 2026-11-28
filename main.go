package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Programming-ark/singularity/command"
)

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(out, "hello world")
		return nil
	}

	switch args[0] {
	case "init":
		return command.Initialize(args[1:], out)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
