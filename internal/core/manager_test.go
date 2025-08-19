package core

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestSopsManager_CheckInitialization(t *testing.T) {
	t.Parallel()

	// Given: a fresh directory and SOPS manager
	service := setupTestEnvironment(t)

	// When: checking initialization on fresh directory
	err := service.checkInitialization(false)

	// Then: it should succeed
	requireNoError(t, err, "checkInitialization(false) on fresh directory should succeed")

	// Given: a config file already exists
	if err := os.WriteFile(service.configPath, []byte("test"), 0o644); err != nil { //nolint:gosec // Test file with dummy content
		t.Fatalf("failed to create test config file: %v", err)
	}

	// When: checking initialization with existing file
	err = service.checkInitialization(false)

	// Then: it should fail
	requireError(t, err, "checkInitialization(false) with existing file should fail")

	// When: checking force initialization
	err = service.checkInitialization(true)

	// Then: it should always succeed
	requireNoError(t, err, "checkInitialization(true) should always succeed")
}

func TestSopsManager_SetupEnvironment(t *testing.T) {
	t.Parallel()

	// Given: a fresh directory and SOPS manager
	service := setupTestEnvironment(t)

	// When: setting up the environment
	err := service.setupEnvironment()

	// Then: it should succeed and create the secrets directory with correct permissions
	requireNoError(t, err, "setupEnvironment() should succeed")
	verifySecretsDirectoryCreated(t, service)
	verifySecretsDirectoryHasCorrectPermissions(t, service)
}

func TestSopsManager_CreateInitialManifest(t *testing.T) {
	t.Parallel()

	service := NewSopsManager("sops")
	memberID := "testuser"
	publicKey := testAgeKeyValue

	manifest := service.createInitialManifest(memberID, publicKey, time.Now().UTC())

	if len(manifest.Members) != 1 {
		t.Errorf("expected 1 member, got %d", len(manifest.Members))
	}

	if manifest.Members[0].ID != memberID {
		t.Errorf("expected member ID %s, got %s", memberID, manifest.Members[0].ID)
	}

	if manifest.Members[0].AgeKey != publicKey {
		t.Errorf("expected member key %s, got %s", publicKey, manifest.Members[0].AgeKey)
	}

	if len(manifest.Scopes) != 1 {
		t.Errorf("expected 1 scope, got %d", len(manifest.Scopes))
	}

	if manifest.Scopes[0].Name != defaultScopeName {
		t.Errorf("expected scope name 'default', got %s", manifest.Scopes[0].Name)
	}

	if len(manifest.Scopes[0].Members) != 1 || manifest.Scopes[0].Members[0] != memberID {
		t.Error("default scope should contain the initial member")
	}

	if manifest.Settings.SopsVersion != "3.8.0" {
		t.Errorf("expected SOPS version 3.8.0, got %s", manifest.Settings.SopsVersion)
	}
}

func TestSopsManager_Init_AlreadyExists(t *testing.T) {
	t.Parallel()

	// Given: a fresh directory and SOPS manager
	service := setupTestEnvironment(t)

	// When: first initialization
	err := service.Init(false)

	// Then: it should succeed
	requireNoError(t, err, "first initialization should succeed")

	// When: second initialization without force
	err = service.Init(false)

	// Then: it should fail
	requireError(t, err, "second initialization without force should fail")

	// When: second initialization with force
	err = service.Init(true)

	// Then: it should succeed
	requireNoError(t, err, "force initialization should succeed")
}

func TestSopsManager_AddMember(t *testing.T) {
	t.Parallel()

	testAgeKey := testAgeKeyValue

	testCases := []struct {
		name       string
		memberID   string
		memberKey  string
		shouldFail bool
	}{
		{
			name:       "should successfully add alice to team",
			memberID:   "alice",
			memberKey:  testAgeKey,
			shouldFail: false,
		},
		{
			name:       "should successfully add bob to team",
			memberID:   "bob",
			memberKey:  testAgeKey,
			shouldFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given: an initialized SOPS manager
			service := setupSopsManagerInTempDir(t)

			// When: adding a member to the team
			err := service.AddMember(tc.memberID, tc.memberKey)

			// Then: the operation should succeed/fail as expected
			if tc.shouldFail {
				requireError(t, err, "expected AddMember to fail")
			} else {
				requireNoError(t, err, "expected AddMember to succeed")
				verifyMemberWasAddedToTeam(t, service, tc.memberID, tc.memberKey)
			}
		})
	}
}

func TestSopsManager_AddMember_Duplicate(t *testing.T) {
	t.Parallel()

	testAgeKey := testAgeKeyValue

	// Given: an initialized SOPS manager with alice already added
	service := setupSopsManagerInTempDir(t)
	err := service.AddMember("alice", testAgeKey)
	requireNoError(t, err, "first AddMember should succeed")

	// When: attempting to add the same member again
	err = service.AddMember("alice", testAgeKey)

	// Then: the operation should fail
	requireError(t, err, "adding duplicate member should fail")
}

func TestSopsManager_RemoveMember(t *testing.T) {
	t.Parallel()

	testAgeKey := testAgeKeyValue

	testCases := []struct {
		name       string
		memberID   string
		shouldFail bool
	}{
		{
			name:       "should successfully remove existing member alice",
			memberID:   "alice",
			shouldFail: false,
		},
		{
			name:       "should fail to remove non-existent member bob",
			memberID:   "bob",
			shouldFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given: an initialized SOPS manager with alice as a member
			service := setupSopsManagerWithMember(t, "alice", testAgeKey)

			// When: removing a member from the team
			err := service.RemoveMember(tc.memberID)

			// Then: the operation should succeed/fail as expected
			if tc.shouldFail {
				requireError(t, err, "expected RemoveMember to fail")
			} else {
				requireNoError(t, err, "expected RemoveMember to succeed")
				verifyMemberWasRemovedFromTeam(t, service, tc.memberID)
			}
		})
	}
}

func TestNewSopsManager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sopsPath string
		want     *SopsManager
	}{
		{
			name:     "default sops path",
			sopsPath: "sops",
			want: &SopsManager{
				sopsPath:   "sops",
				configPath: "sopsistry.yaml",
				secretsDir: ".secrets",
			},
		},
		{
			name:     "custom sops path",
			sopsPath: "/usr/local/bin/sops",
			want: &SopsManager{
				sopsPath:   "/usr/local/bin/sops",
				configPath: "sopsistry.yaml",
				secretsDir: ".secrets",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewSopsManager(tt.sopsPath)
			if got.sopsPath != tt.want.sopsPath {
				t.Errorf("NewSopsManager().sopsPath = %v, want %v", got.sopsPath, tt.want.sopsPath)
			}
			if got.configPath != tt.want.configPath {
				t.Errorf("NewSopsManager().configPath = %v, want %v", got.configPath, tt.want.configPath)
			}
			if got.secretsDir != tt.want.secretsDir {
				t.Errorf("NewSopsManager().secretsDir = %v, want %v", got.secretsDir, tt.want.secretsDir)
			}
		})
	}
}

// Test Helper Functions - make tests readable as executable specifications

func setupSopsManagerInTempDir(t *testing.T) *SopsManager {
	t.Helper()

	service := setupTestEnvironment(t)
	initializeSopsManager(t, service)
	return service
}

func setupTestEnvironment(t *testing.T) *SopsManager {
	t.Helper()

	// Create unique temp directory for this test (like Python tmpdir fixture)
	tempDir := t.TempDir()

	// Create a service that works with absolute paths - no chdir needed!
	service := createSopsManagerInDir(tempDir)
	return service
}

// createSopsManagerInDir creates a SopsManager that works in a specific directory
// without changing the global working directory (like Python's tmpdir fixture)
func createSopsManagerInDir(workDir string) *SopsManager {
	return &SopsManager{
		sopsPath:   "sops",
		configPath: filepath.Join(workDir, "sopsistry.yaml"),
		secretsDir: filepath.Join(workDir, ".secrets"),
		output:     io.Discard, // Silent output for tests
	}
}

func initializeSopsManager(t *testing.T, service *SopsManager) {
	t.Helper()

	if err := service.Init(false); err != nil {
		t.Fatalf("SOPS manager initialization failed: %v", err)
	}
}

func requireNoError(t *testing.T, err error, message string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: got error %v", message, err)
	}
}

func requireError(t *testing.T, err error, message string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s: expected error but got nil", message)
	}
}

func verifyMemberWasAddedToTeam(t *testing.T, service *SopsManager, memberID, expectedKey string) {
	t.Helper()

	manifest := loadManifestOrFail(t, service.configPath)
	verifyMemberExistsInManifest(t, manifest, memberID, expectedKey)
	verifyMemberExistsInDefaultScope(t, manifest, memberID)
}

func loadManifestOrFail(t *testing.T, configPath string) *Manifest {
	t.Helper()

	manifest, err := LoadManifest(configPath)
	if err != nil {
		t.Fatalf("failed to load manifest from %s: %v", configPath, err)
	}
	return manifest
}

func verifyMemberExistsInManifest(t *testing.T, manifest *Manifest, memberID, expectedKey string) {
	t.Helper()

	for _, member := range manifest.Members {
		if member.ID == memberID {
			if member.AgeKey != expectedKey {
				t.Errorf("member %s has wrong age key: got %s, want %s",
					memberID, member.AgeKey, expectedKey)
			}
			return
		}
	}
	t.Errorf("member %s was not found in manifest", memberID)
}

func verifyMemberExistsInDefaultScope(t *testing.T, manifest *Manifest, memberID string) {
	t.Helper()

	for _, scope := range manifest.Scopes {
		if scope.Name == "default" {
			if slices.Contains(scope.Members, memberID) {
				return // Found member in default scope
			}
			t.Errorf("member %s was not found in default scope", memberID)
			return
		}
	}
	t.Error("default scope was not found in manifest")
}

func setupSopsManagerWithMember(t *testing.T, memberID, memberKey string) *SopsManager {
	t.Helper()

	service := setupSopsManagerInTempDir(t)
	err := service.AddMember(memberID, memberKey)
	requireNoError(t, err, "failed to add initial member")
	return service
}

func verifyMemberWasRemovedFromTeam(t *testing.T, service *SopsManager, memberID string) {
	t.Helper()

	manifest := loadManifestOrFail(t, service.configPath)
	verifyMemberNotInManifest(t, manifest, memberID)
	verifyMemberNotInAnyScope(t, manifest, memberID)
}

func verifyMemberNotInManifest(t *testing.T, manifest *Manifest, memberID string) {
	t.Helper()

	for _, member := range manifest.Members {
		if member.ID == memberID {
			t.Errorf("member %s was not removed from manifest", memberID)
			return
		}
	}
}

func verifyMemberNotInAnyScope(t *testing.T, manifest *Manifest, memberID string) {
	t.Helper()

	for _, scope := range manifest.Scopes {
		if slices.Contains(scope.Members, memberID) {
			t.Errorf("member %s was not removed from scope %s", memberID, scope.Name)
			return
		}
	}
}

func verifySecretsDirectoryCreated(t *testing.T, service *SopsManager) {
	t.Helper()

	if _, err := os.Stat(service.secretsDir); os.IsNotExist(err) {
		t.Errorf(".secrets directory was not created at %s", service.secretsDir)
	}
}

func verifySecretsDirectoryHasCorrectPermissions(t *testing.T, service *SopsManager) {
	t.Helper()

	info, err := os.Stat(service.secretsDir)
	if err != nil {
		t.Errorf("failed to stat .secrets directory at %s: %v", service.secretsDir, err)
		return
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("wrong permissions on .secrets directory: got %o, want 0700", info.Mode().Perm())
	}
}
