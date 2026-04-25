package main

import (
	"fmt"
	"os"

	"github.com/chnxq/xkit/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "xkit: %v\n", err)
		os.Exit(1)
	}
}
