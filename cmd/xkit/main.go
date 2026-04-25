package main

import (
	"fmt"
	"os"

	"github.com/chnxq/xkit/internal/cli"
)

const Version = "v0.1.0"

func main() {
	if err := cli.Run(os.Args[1:], Version); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "xkit: %v\n", err)
		os.Exit(1)
	}
}
