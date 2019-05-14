package main

import (
	"fmt"
	"os"

	"github.com/wantedly/subee/cmd/subee/cmd"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	return cmd.NewSubeeCmd().Execute()
}
