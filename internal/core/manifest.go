package core

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Member represents a team member with their age key
type Member struct {
	ID     string `yaml:"id" json:"id"`
	AgeKey string `yaml:"age_key" json:"age_key"`
}

// Scope defines which files are encrypted for which members
type Scope struct {
	Name     string   `yaml:"name" json:"name"`
	Patterns []string `yaml:"patterns" json:"patterns"`
	Members  []string `yaml:"members" json:"members"`
}

// Settings contains global configuration
type Settings struct {
	SopsVersion string `yaml:"sops_version" json:"sops_version"`
}

// Manifest represents the sopsistry.yaml configuration
type Manifest struct {
	Members  []Member `yaml:"members" json:"members"`
	Scopes   []Scope  `yaml:"scopes" json:"scopes"`
	Settings Settings `yaml:"settings" json:"settings"`
}

// LoadManifest loads the team manifest from file
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path) //nolint:gosec // Reading user-provided config file is expected
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// Save writes the manifest to file
func (m *Manifest) Save(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil { //nolint:gosec // Config files use standard permissions
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// Display shows the manifest in human-readable format
func (m *Manifest) Display() {
	fmt.Println("Team Members:")
	if len(m.Members) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, member := range m.Members {
			fmt.Printf("  %s: %s\n", member.ID, member.AgeKey[:16]+"...")
		}
	}

	fmt.Println("\nScopes:")
	for _, scope := range m.Scopes {
		fmt.Printf("  %s:\n", scope.Name)
		fmt.Printf("    Patterns: %v\n", scope.Patterns)
		fmt.Printf("    Members: %v\n", scope.Members)
	}

	fmt.Printf("\nSettings:\n")
	fmt.Printf("  SOPS Version: %s\n", m.Settings.SopsVersion)
}

// DisplayJSON outputs the manifest as JSON
func (m *Manifest) DisplayJSON() error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// GetMemberAgeKey returns the age key for a member ID
func (m *Manifest) GetMemberAgeKey(id string) (string, bool) {
	for _, member := range m.Members {
		if member.ID == id {
			return member.AgeKey, true
		}
	}
	return "", false
}

// GetScopeMembers returns all members for a given scope
func (m *Manifest) GetScopeMembers(scopeName string) ([]Member, error) {
	var scope *Scope
	for i := range m.Scopes {
		if m.Scopes[i].Name == scopeName {
			scope = &m.Scopes[i]
			break
		}
	}

	if scope == nil {
		return nil, fmt.Errorf("scope %s not found", scopeName)
	}

	var members []Member //nolint:prealloc // Small team sizes, optimization not worth it
	for _, memberID := range scope.Members {
		ageKey, found := m.GetMemberAgeKey(memberID)
		if !found {
			return nil, fmt.Errorf("member %s not found", memberID)
		}
		members = append(members, Member{ID: memberID, AgeKey: ageKey})
	}

	return members, nil
}
