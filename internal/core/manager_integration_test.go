package core

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

const (
	testAgeKeyValue  = "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p"
	defaultScopeName = "default"
)

// Integration tests that require age-keygen to be installed
// These tests are separate to allow unit tests to run without external dependencies

func TestSopsManager_Init_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfAgeKeygenUnavailable(t)
	t.Parallel()

	// Given: a fresh directory and SOPS manager
	service := setupIntegrationTestEnvironment(t)

	// When: initializing the SOPS manager
	err := service.Init(false)

	// Then: initialization should succeed and create all required files
	requireNoError(t, err, "SOPS manager initialization should succeed")
	verifyInitializationArtifacts(t, service)
}

func TestSopsManager_AddMember_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfAgeKeygenUnavailable(t)
	t.Parallel()

	testAgeKey := testAgeKeyValue

	// Given: an initialized SOPS manager
	service := setupInitializedIntegrationService(t)

	// When: adding a valid member
	err := service.AddMember("alice", testAgeKey)

	// Then: the member should be added successfully
	requireNoError(t, err, "adding valid member should succeed")
	verifyMemberWasAddedToTeam(t, service, "alice", testAgeKey)

	// When: attempting to add the same member again
	err = service.AddMember("alice", testAgeKey)

	// Then: the operation should fail
	requireError(t, err, "adding duplicate member should fail")

	// When: removing the member
	err = service.RemoveMember("alice")

	// Then: the member should be removed successfully
	requireNoError(t, err, "removing member should succeed")
	verifyMemberWasRemovedFromTeam(t, service, "alice")
}

// Integration test er functions

func skipIfAgeKeygenUnavailable(t *testing.T) {
	t.Helper()

	if err := ensureBinaryAvailable("age-keygen", "Please install age for integration tests"); err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
}

func setupIntegrationTestEnvironment(t *testing.T) *SopsManager {
	t.Helper()

	// Create unique temp directory (like Python tmpdir fixture)
	tempDir := t.TempDir()

	// Create service with absolute paths - no chdir needed!
	return &SopsManager{
		sopsPath:   "sops",
		configPath: filepath.Join(tempDir, "sopsistry.yaml"),
		secretsDir: filepath.Join(tempDir, ".secrets"),
		output:     io.Discard, // Silent output for tests
	}
}

func setupInitializedIntegrationService(t *testing.T) *SopsManager {
	t.Helper()

	service := setupIntegrationTestEnvironment(t)
	err := service.Init(false)
	requireNoError(t, err, "service initialization should succeed")
	return service
}

func verifyInitializationArtifacts(t *testing.T, service *SopsManager) {
	t.Helper()

	verifyConfigFileExists(t, service.configPath)
	verifySecretsDirectoryExists(t, service.secretsDir)
	verifyAgeKeyExists(t, service.secretsDir)
	verifyManifestIsValid(t, service.configPath)
}

func verifyConfigFileExists(t *testing.T, configPath string) {
	t.Helper()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file was not created at %s", configPath)
	}
}

func verifySecretsDirectoryExists(t *testing.T, secretsDir string) {
	t.Helper()

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		t.Errorf(".secrets directory was not created at %s", secretsDir)
	}
}

func verifyAgeKeyExists(t *testing.T, secretsDir string) {
	t.Helper()

	pattern := filepath.Join(secretsDir, "key-*.txt")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Errorf("failed to search for age key files: %v", err)
		return
	}
	if len(matches) == 0 {
		t.Errorf("no age key files found matching pattern %s", pattern)
		return
	}
	// Verify at least one key file exists and is readable
	keyPath := matches[0]
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("age key file was not created at %s", keyPath)
	}
}

func verifyManifestIsValid(t *testing.T, configPath string) {
	t.Helper()

	manifest := loadManifestOrFail(t, configPath)
	verifyManifestHasMembers(t, manifest)
	verifyManifestHasScopes(t, manifest)
	verifyDefaultScopeIsValid(t, manifest)
}

func verifyManifestHasMembers(t *testing.T, manifest *Manifest) {
	t.Helper()

	if len(manifest.Members) == 0 {
		t.Error("no members were added to manifest")
	}
}

func verifyManifestHasScopes(t *testing.T, manifest *Manifest) {
	t.Helper()

	if len(manifest.Scopes) == 0 {
		t.Error("no scopes were added to manifest")
	}
}

func verifyDefaultScopeIsValid(t *testing.T, manifest *Manifest) {
	t.Helper()

	for _, scope := range manifest.Scopes {
		if scope.Name == defaultScopeName {
			verifyDefaultScopeHasMembers(t, scope)
			verifyDefaultScopeHasPatterns(t, scope)
			return
		}
	}
	t.Error("default scope was not created")
}

func verifyDefaultScopeHasMembers(t *testing.T, scope Scope) {
	t.Helper()

	if len(scope.Members) == 0 {
		t.Error("default scope has no members")
	}
}

func verifyDefaultScopeHasPatterns(t *testing.T, scope Scope) {
	t.Helper()

	if len(scope.Patterns) == 0 {
		t.Error("default scope has no patterns")
	}
}
