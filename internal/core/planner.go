package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ActionType represents the type of action to be performed
type ActionType string

// Action types for SOPS operations
const (
	ActionEncrypt   ActionType = "encrypt"    // Encrypt a new file
	ActionReencrypt ActionType = "re-encrypt" // Re-encrypt existing file with new keys
	ActionSkip      ActionType = "skip"       // Skip file (no members in scope)
)

// Action represents a single planned action
type Action struct { //nolint:govet // Field alignment optimization not critical for this struct
	Recipients  []string   `json:"recipients"`
	File        string     `json:"file"`
	Scope       string     `json:"scope"`
	Description string     `json:"description"`
	Type        ActionType `json:"type"`
}

// Plan contains all planned actions
type Plan struct {
	Actions []Action `json:"actions"`
}

// Planner computes execution plans for SOPS operations
type Planner struct {
	sopsPath string
}

// NewPlanner creates a new planner instance
func NewPlanner(sopsPath string) *Planner {
	return &Planner{
		sopsPath: sopsPath,
	}
}

// ComputePlan calculates what actions need to be taken
func (p *Planner) ComputePlan(manifest *Manifest) (*Plan, error) {
	plan := &Plan{Actions: []Action{}}

	for _, scope := range manifest.Scopes {
		actions, err := p.planScopeActions(scope, manifest)
		if err != nil {
			return nil, err
		}
		plan.Actions = append(plan.Actions, actions...)
	}

	return plan, nil
}

func (p *Planner) planScopeActions(scope Scope, manifest *Manifest) ([]Action, error) {
	files, err := p.findMatchingFiles(scope.Patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to find files for scope %s: %w", scope.Name, err)
	}

	members, err := manifest.GetScopeMembers(scope.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get members for scope %s: %w", scope.Name, err)
	}

	if len(members) == 0 {
		return p.createSkipActions(files, scope.Name), nil
	}

	recipients := p.extractAgeKeys(members)
	return p.createFileActions(files, scope.Name, recipients), nil
}

func (p *Planner) createSkipActions(files []string, scopeName string) []Action {
	actions := make([]Action, 0, DefaultSliceCapacity)
	for _, file := range files {
		actions = append(actions, Action{
			Type:        ActionSkip,
			File:        file,
			Scope:       scopeName,
			Recipients:  []string{},
			Description: "No members in scope",
		})
	}
	return actions
}

func (p *Planner) extractAgeKeys(members []Member) []string {
	recipients := make([]string, 0, DefaultSliceCapacity)
	for _, member := range members {
		recipients = append(recipients, member.AgeKey)
	}
	return recipients
}

func (p *Planner) createFileActions(files []string, scopeName string, recipients []string) []Action {
	actions := make([]Action, 0, DefaultSliceCapacity)
	for _, file := range files {
		actionType := ActionEncrypt
		description := "Encrypt with current team"

		if p.isSOPSFile(file) {
			actionType = ActionReencrypt
			description = "Re-encrypt with updated team"
		}

		actions = append(actions, Action{
			Type:        actionType,
			File:        file,
			Scope:       scopeName,
			Recipients:  recipients,
			Description: description,
		})
	}
	return actions
}

// Display shows the plan in human-readable format
func (p *Plan) Display(noColor bool) {
	if len(p.Actions) == 0 {
		fmt.Println("No changes planned")
		return
	}

	p.displayHeader()
	p.displayActions(noColor)
	p.displayLegend()
}

func (p *Plan) displayHeader() {
	fmt.Printf("Planned actions (%d files):\n\n", len(p.Actions))
}

func (p *Plan) displayActions(noColor bool) {
	for _, action := range p.Actions {
		prefix := p.getActionPrefix(action.Type, noColor)
		p.displayAction(&action, prefix)
	}
}

// ActionDisplay provides type-safe display formatting for actions
type ActionDisplay struct {
	actionType ActionType
}

// NewActionDisplay creates a type-safe display formatter for the given action type
func NewActionDisplay(actionType ActionType) ActionDisplay {
	return ActionDisplay{actionType: actionType}
}

// ColoredFormat returns the ANSI-formatted display symbol for this action type
func (a ActionDisplay) ColoredFormat() string {
	switch a.actionType {
	case ActionEncrypt:
		return "\033[32m+\033[0m" // Green +
	case ActionReencrypt:
		return "\033[33m~\033[0m" // Yellow ~
	case ActionSkip:
		return "\033[90m-\033[0m" // Gray -
	default:
		return "?"
	}
}

// PlainFormat returns the plain text display symbol for this action type
func (a ActionDisplay) PlainFormat() string {
	switch a.actionType {
	case ActionEncrypt:
		return "+"
	case ActionReencrypt:
		return "~"
	case ActionSkip:
		return "-"
	default:
		return "?"
	}
}

func (p *Plan) getActionPrefix(actionType ActionType, noColor bool) string { //nolint:revive // noColor is a legitimate CLI flag parameter
	display := NewActionDisplay(actionType)
	if noColor {
		return display.PlainFormat()
	}
	return display.ColoredFormat()
}

func (p *Plan) displayAction(action *Action, prefix string) { //nolint:gocritic // Passing by pointer for performance
	fmt.Printf("%s %s (%s): %s\n",
		prefix, action.File, action.Scope, action.Description)

	if action.Type != ActionSkip && len(action.Recipients) > 0 {
		fmt.Printf("  Recipients: %d keys\n", len(action.Recipients))
	}
}

func (p *Plan) displayLegend() {
	fmt.Printf("\nLegend:\n")
	fmt.Printf("  + = new encryption\n")
	fmt.Printf("  ~ = re-encryption\n")
	fmt.Printf("  - = skipped\n")
}

// findMatchingFiles finds all files matching the given patterns
func (p *Planner) findMatchingFiles(patterns []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		patternFiles, err := p.findFilesForPattern(pattern, seen)
		if err != nil {
			return nil, err
		}
		files = append(files, patternFiles...)
	}

	return files, nil
}

func (p *Planner) findFilesForPattern(pattern string, seen map[string]bool) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
	}

	var files []string
	for _, match := range matches {
		if p.shouldIncludeFile(match, seen) {
			files = append(files, match)
			seen[match] = true
		}
	}

	return files, nil
}

func (p *Planner) shouldIncludeFile(filePath string, seen map[string]bool) bool {
	if seen[filePath] {
		return false
	}

	if info, err := os.Stat(filePath); err == nil && info.IsDir() {
		return false
	}

	return true
}

// isSOPSFile checks if a file is already encrypted with SOPS
func (p *Planner) isSOPSFile(file string) bool {
	data, err := os.ReadFile(file) //nolint:gosec // Reading project files for analysis is expected
	if err != nil {
		return false
	}

	content := string(data)

	return strings.Contains(content, "sops:") ||
		strings.Contains(content, "\"sops\"") ||
		strings.Contains(content, "lastmodified") ||
		strings.Contains(content, "mac")
}
