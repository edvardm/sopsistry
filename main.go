// Package main implements the SOPS Team CLI for managing team-based secret encryption.
package main

import (
	"os"

	"github.com/edvardm/sopsistry/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
