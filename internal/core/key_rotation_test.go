package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestKeyRotation_ValidationLogic(t *testing.T) {
	// Test the key expiry validation logic without external dependencies
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	configPath := filepath.Join(tempDir, "sopsistry.yaml")
	secretsDir := filepath.Join(tempDir, ".secrets")

	// Create initial manifest with a fresh key
	createdTime := time.Now().UTC().AddDate(0, 0, -7) // 7 days ago
	manifest := &Manifest{
		Members: []Member{
			{
				ID:      "testuser",
				AgeKey:  "age1testkey12345...",
				Created: createdTime,
			},
		},
		Settings: Settings{
			SopsVersion:   "3.8.0",
			MaxKeyAgeDays: 180,
		},
	}

	if err := manifest.Save(configPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	if err := os.MkdirAll(secretsDir, 0o700); err != nil {
		t.Fatalf("Failed to create secrets dir: %v", err)
	}

	// Mock the current user
	_ = os.Setenv("USER", "testuser")
	defer func() { _ = os.Unsetenv("USER") }()

	var output bytes.Buffer
	service := &SopsManager{
		sopsPath:   "/nonexistent/sops", // Will fail, but we test validation first
		configPath: configPath,
		secretsDir: secretsDir,
		output:     &output,
	}

	// Test key expiry check logic directly
	member := &manifest.Members[0]
	err := service.checkKeyExpiry(member, 180)
	if err != nil {
		t.Errorf("Fresh key should not be expired: %v", err)
	}

	// Test expired key
	member.Created = time.Now().UTC().AddDate(0, 0, -200) // 200 days ago
	err = service.checkKeyExpiry(member, 180)
	if err == nil {
		t.Error("Expired key should return error")
	}
	if !containsString(err.Error(), "key has expired") {
		t.Errorf("Expected expiry error, got: %v", err)
	}
}

func TestKeyRotation_ExpiredKey(t *testing.T) {
	// Setup test environment
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	configPath := filepath.Join(tempDir, "sopsistry.yaml")
	secretsDir := filepath.Join(tempDir, ".secrets")

	// Get the actual current user for this test
	service := &SopsManager{}
	currentUser, err := service.getCurrentMemberID()
	if err != nil {
		t.Skipf("Cannot get current user: %v", err)
	}

	// Create manifest with expired key (created 200 days ago, max age 180)
	expiredTime := time.Now().UTC().AddDate(0, 0, -200)
	manifest := &Manifest{
		Members: []Member{
			{
				ID:      currentUser, // Use actual current user
				AgeKey:  "age1expiredkey...",
				Created: expiredTime,
			},
		},
		Settings: Settings{
			SopsVersion:   "3.8.0",
			MaxKeyAgeDays: 180,
		},
	}

	if err := manifest.Save(configPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Create secrets directory
	if err := os.MkdirAll(secretsDir, 0o700); err != nil {
		t.Fatalf("Failed to create secrets dir: %v", err)
	}

	var output bytes.Buffer
	service = &SopsManager{
		sopsPath:   "echo",
		configPath: configPath,
		secretsDir: secretsDir,
		output:     &output,
	}

	// Test rotation without force - should fail
	err = service.RotateKey(false)
	if err == nil {
		t.Fatal("Expected error for expired key without force, got nil")
	}

	expectedError := "key has expired"
	if !containsString(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: %s", expectedError, err.Error())
	}

	// Test with force - will fail due to missing binaries, but that's expected
	err = service.RotateKey(true)
	if err == nil {
		t.Skip("Unexpected success - would require real binaries")
	}

	// Should get a different error (not expiry) when using force
	if containsString(err.Error(), "key has expired") {
		t.Error("With force=true, should not get expiry error")
	}
}

func TestCheckKeyExpiry_Warnings(t *testing.T) {
	// Setup test environment
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	configPath := filepath.Join(tempDir, "sopsistry.yaml")
	now := time.Now().UTC()

	// Create manifest with keys in different states
	manifest := &Manifest{
		Members: []Member{
			{
				ID:      "fresh-user",
				AgeKey:  "age1fresh...",
				Created: now.AddDate(0, 0, -30), // 30 days old - fresh
			},
			{
				ID:      "warning-user",
				AgeKey:  "age1warning...",
				Created: now.AddDate(0, 0, -170), // 170 days old - warning (expires in 10 days)
			},
			{
				ID:      "expired-user",
				AgeKey:  "age1expired...",
				Created: now.AddDate(0, 0, -200), // 200 days old - expired
			},
		},
		Settings: Settings{
			SopsVersion:   "3.8.0",
			MaxKeyAgeDays: 180,
		},
	}

	if err := manifest.Save(configPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	var output bytes.Buffer
	service := &SopsManager{
		sopsPath:   "sops",
		configPath: configPath,
		secretsDir: ".secrets",
		output:     &output,
	}

	// Check key expiry
	err := service.CheckKeyExpiry(false)
	if err != nil {
		t.Fatalf("CheckKeyExpiry failed: %v", err)
	}

	outputStr := output.String()

	// Verify fresh key shows as OK
	if !containsString(outputStr, "fresh-user: key is 30 days old") {
		t.Errorf("Expected fresh key status, got: %s", outputStr)
	}

	// Verify warning for soon-to-expire key
	if !containsString(outputStr, "warning-user: key expires in") {
		t.Errorf("Expected warning for expiring key, got: %s", outputStr)
	}

	// Verify error for expired key
	if !containsString(outputStr, "expired-user: key expired") {
		t.Errorf("Expected error for expired key, got: %s", outputStr)
	}

	// Verify summary messages
	if !containsString(outputStr, "1 expired keys found") {
		t.Errorf("Expected expired keys summary, got: %s", outputStr)
	}
	if !containsString(outputStr, "1 keys expiring soon") {
		t.Errorf("Expected warning keys summary, got: %s", outputStr)
	}
}

func TestKeyRotation_UserNotFound(t *testing.T) {
	// Test error handling when current user is not in the manifest
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	configPath := filepath.Join(tempDir, "sopsistry.yaml")
	secretsDir := filepath.Join(tempDir, ".secrets")

	// Create manifest without the current user
	manifest := &Manifest{
		Members: []Member{
			{
				ID:      "someoneelse",
				AgeKey:  "age1someoneelse...",
				Created: time.Now().UTC(),
			},
		},
		Settings: Settings{
			SopsVersion:   "3.8.0",
			MaxKeyAgeDays: 180,
		},
	}

	if err := manifest.Save(configPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	if err := os.MkdirAll(secretsDir, 0o700); err != nil {
		t.Fatalf("Failed to create secrets dir: %v", err)
	}

	// Mock current user not in manifest
	originalUser := os.Getenv("USER")
	_ = os.Setenv("USER", "notfounduser")
	defer func() {
		if originalUser != "" {
			_ = os.Setenv("USER", originalUser)
		} else {
			_ = os.Unsetenv("USER")
		}
	}()

	var output bytes.Buffer
	service := &SopsManager{
		sopsPath:   "sops",
		configPath: configPath,
		secretsDir: secretsDir,
		output:     &output,
	}

	// Attempt rotation - should fail with user not found
	err := service.RotateKey(false)
	if err == nil {
		t.Fatal("Expected error for user not found, got nil")
	}

	expectedError := "not found in team"
	if !containsString(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: %s", expectedError, err.Error())
	}
}

// Helper functions
func setupTestDir(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "sopsistry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tempDir
}

func cleanupTestDir(dir string) {
	_ = os.RemoveAll(dir)
}

func containsString(haystack, needle string) bool {
	return bytes.Contains([]byte(haystack), []byte(needle))
}
