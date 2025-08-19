package core

import (
	"path/filepath"
	"testing"
)

func TestManifest_SaveAndLoad(t *testing.T) {
	t.Parallel()

	// Given: a test environment and a manifest with sample data
	tempDir := t.TempDir()
	original := createSampleManifest()
	manifestPath := filepath.Join(tempDir, "test-manifest.yaml")

	// When: saving the manifest to a file
	err := original.Save(manifestPath)

	// Then: the save operation should succeed
	requireNoError(t, err, "saving manifest should succeed")

	// When: loading the manifest from the file
	loaded, err := LoadManifest(manifestPath)

	// Then: the load operation should succeed and data should match
	requireNoError(t, err, "loading manifest should succeed")
	verifyManifestDataMatches(t, original, loaded)
}

func TestManifest_GetMemberAgeKey(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Members: []Member{
			{
				ID:     "alice",
				AgeKey: "age1alice",
			},
			{
				ID:     "bob",
				AgeKey: "age1bob",
			},
		},
	}

	// Test existing member
	key, found := manifest.GetMemberAgeKey("alice")
	if !found {
		t.Error("expected to find alice")
	}
	if key != "age1alice" {
		t.Errorf("expected alice's key to be 'age1alice', got %s", key)
	}

	// Test non-existing member
	_, found = manifest.GetMemberAgeKey("charlie")
	if found {
		t.Error("expected not to find charlie")
	}
}

func TestManifest_GetScopeMembers(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Members: []Member{
			{
				ID:     "alice",
				AgeKey: "age1alice",
			},
			{
				ID:     "bob",
				AgeKey: "age1bob",
			},
		},
		Scopes: []Scope{
			{
				Name:    "default",
				Members: []string{"alice", "bob"},
			},
			{
				Name:    "restricted",
				Members: []string{"alice"},
			},
		},
	}

	// Test existing scope with multiple members
	members, err := manifest.GetScopeMembers("default")
	if err != nil {
		t.Errorf("GetScopeMembers('default') failed: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("expected 2 members in default scope, got %d", len(members))
	}

	// Test existing scope with single member
	members, err = manifest.GetScopeMembers("restricted")
	if err != nil {
		t.Errorf("GetScopeMembers('restricted') failed: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("expected 1 member in restricted scope, got %d", len(members))
	}
	if members[0].ID != "alice" {
		t.Errorf("expected alice in restricted scope, got %s", members[0].ID)
	}

	// Test non-existing scope
	_, err = manifest.GetScopeMembers("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent scope")
	}

	// Test scope with non-existent member
	manifest.Scopes = append(manifest.Scopes, Scope{
		Name:    "broken",
		Members: []string{"charlie"}, // charlie doesn't exist in members
	})

	_, err = manifest.GetScopeMembers("broken")
	if err == nil {
		t.Error("expected error for scope with non-existent member")
	}
}

// Manifest test er functions

func createSampleManifest() *Manifest {
	return &Manifest{
		Members: []Member{
			{
				ID:     "alice",
				AgeKey: "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p",
			},
			{
				ID:     "bob",
				AgeKey: "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p",
			},
		},
		Scopes: []Scope{
			{
				Name:     "default",
				Patterns: []string{"*.sops.yaml", "*.sops.json"},
				Members:  []string{"alice", "bob"},
			},
		},
		Settings: Settings{
			SopsVersion: "3.8.0",
		},
	}
}

func verifyManifestDataMatches(t *testing.T, original, loaded *Manifest) {
	t.Helper()

	verifyMembersMatch(t, original.Members, loaded.Members)
	verifyScopesMatch(t, original.Scopes, loaded.Scopes)
	verifySettingsMatch(t, original.Settings, loaded.Settings)
}

func verifyMembersMatch(t *testing.T, originalMembers, loadedMembers []Member) {
	t.Helper()

	if len(loadedMembers) != len(originalMembers) {
		t.Errorf("expected %d members, got %d", len(originalMembers), len(loadedMembers))
		return
	}

	for i, member := range loadedMembers {
		verifyMemberMatches(t, originalMembers[i], member, i)
	}
}

func verifyMemberMatches(t *testing.T, original, loaded Member, index int) {
	t.Helper()

	if loaded.ID != original.ID {
		t.Errorf("member %d ID mismatch: got %s, want %s", index, loaded.ID, original.ID)
	}
	if loaded.AgeKey != original.AgeKey {
		t.Errorf("member %d AgeKey mismatch: got %s, want %s", index, loaded.AgeKey, original.AgeKey)
	}
}

func verifyScopesMatch(t *testing.T, originalScopes, loadedScopes []Scope) {
	t.Helper()

	if len(loadedScopes) != len(originalScopes) {
		t.Errorf("expected %d scopes, got %d", len(originalScopes), len(loadedScopes))
		return
	}

	for i, scope := range loadedScopes {
		verifyScopeMatches(t, originalScopes[i], scope, i)
	}
}

func verifyScopeMatches(t *testing.T, original, loaded Scope, index int) {
	t.Helper()

	if loaded.Name != original.Name {
		t.Errorf("scope %d name mismatch: got %s, want %s", index, loaded.Name, original.Name)
	}
	if len(loaded.Members) != len(original.Members) {
		t.Errorf("scope %d members count mismatch: got %d, want %d", index, len(loaded.Members), len(original.Members))
	}
	if len(loaded.Patterns) != len(original.Patterns) {
		t.Errorf("scope %d patterns count mismatch: got %d, want %d", index, len(loaded.Patterns), len(original.Patterns))
	}
}

func verifySettingsMatch(t *testing.T, original, loaded Settings) {
	t.Helper()

	if loaded.SopsVersion != original.SopsVersion {
		t.Errorf("settings version mismatch: got %s, want %s", loaded.SopsVersion, original.SopsVersion)
	}
}
