package main

import (
	"fmt"
	"os"

	"github.com/chnxq/xkit/internal/cli"
)

const Version = "v0.2.0 2026-05-28"

func main() {
	if err := cli.Run(os.Args[1:], Version); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "xkit: %v\n", err)
		os.Exit(1)
	}
}
