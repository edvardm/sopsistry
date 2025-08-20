package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// SafeCommand wraps cobra.Command with guaranteed flag registration
type SafeCommand struct {
	*cobra.Command
	stringFlags map[string]bool
	boolFlags   map[string]bool
}

// NewSafeCommand creates a command with guaranteed flag tracking
func NewSafeCommand(cmd *cobra.Command) *SafeCommand {
	return &SafeCommand{
		Command:     cmd,
		stringFlags: make(map[string]bool),
		boolFlags:   make(map[string]bool),
	}
}

// RegisterStringFlag registers and tracks a string flag
func (sc *SafeCommand) RegisterStringFlag(name, defaultVal, usage string) {
	sc.Command.Flags().String(name, defaultVal, usage)
	sc.stringFlags[name] = true
}

// RegisterBoolFlag registers and tracks a boolean flag
func (sc *SafeCommand) RegisterBoolFlag(name string, defaultVal bool, usage string) {
	sc.Command.Flags().Bool(name, defaultVal, usage)
	sc.boolFlags[name] = true
}

// GetStringFlag safely retrieves a registered string flag (including persistent flags)
func (sc *SafeCommand) GetStringFlag(name string) string {
	value, err := sc.Command.Flags().GetString(name)
	if err != nil {
		panic(fmt.Sprintf("PROGRAMMING ERROR: flag '%s' not accessible: %v", name, err))
	}
	return value
}

// GetBoolFlag safely retrieves a registered boolean flag (including persistent flags)
func (sc *SafeCommand) GetBoolFlag(name string) bool {
	value, err := sc.Command.Flags().GetBool(name)
	if err != nil {
		panic(fmt.Sprintf("PROGRAMMING ERROR: flag '%s' not accessible: %v", name, err))
	}
	return value
}
