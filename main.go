package main

import (
	"context"
	"fmt"
	"github.com/tetratom/cftool/internal/cli"
	"os"
)

func main() {
	err := cli.Entry(context.Background(), os.Args)

	if err != nil {
		fmt.Printf("ERROR: %v", err)
		os.Exit(1)
	}
}
