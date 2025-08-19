package core

import (
	"os"
	"testing"
)

func TestGetCurrentMemberID_EnvironmentOverride(t *testing.T) {
	t.Parallel()

	service := NewSopsManager("sops")

	// Test 1: No override set - should use system user
	originalSopsistryUserID := os.Getenv("SOPSISTRY_USER_ID")
	_ = os.Unsetenv("SOPSISTRY_USER_ID")
	defer func() {
		if originalSopsistryUserID != "" {
			_ = os.Setenv("SOPSISTRY_USER_ID", originalSopsistryUserID)
		}
	}()

	systemUserID, err := service.getCurrentMemberID()
	if err != nil {
		t.Fatalf("Expected getCurrentMemberID to succeed without override: %v", err)
	}
	if systemUserID == "" {
		t.Error("Expected non-empty user ID from system")
	}

	// Test 2: Override set - should use override value
	expectedOverride := "custom-user-123"
	_ = os.Setenv("SOPSISTRY_USER_ID", expectedOverride)

	overrideUserID, err := service.getCurrentMemberID()
	if err != nil {
		t.Fatalf("Expected getCurrentMemberID to succeed with override: %v", err)
	}
	if overrideUserID != expectedOverride {
		t.Errorf("Expected override user ID '%s', got '%s'", expectedOverride, overrideUserID)
	}

	// Test 3: Empty override should fall back to system user
	_ = os.Setenv("SOPSISTRY_USER_ID", "")

	fallbackUserID, err := service.getCurrentMemberID()
	if err != nil {
		t.Fatalf("Expected getCurrentMemberID to succeed with empty override: %v", err)
	}
	if fallbackUserID != systemUserID {
		t.Errorf("Expected fallback to system user ID '%s', got '%s'", systemUserID, fallbackUserID)
	}
}

func TestInit_WithUserOverride(t *testing.T) {
	t.Parallel()

	// Setup temporary environment
	service := setupTestEnvironment(t)

	// Set override user ID
	expectedUserID := "test-override-user"
	originalSopsistryUserID := os.Getenv("SOPSISTRY_USER_ID")
	_ = os.Setenv("SOPSISTRY_USER_ID", expectedUserID)
	defer func() {
		if originalSopsistryUserID != "" {
			_ = os.Setenv("SOPSISTRY_USER_ID", originalSopsistryUserID)
		} else {
			_ = os.Unsetenv("SOPSISTRY_USER_ID")
		}
	}()

	// Initialize with override
	err := service.Init(false)
	requireNoError(t, err, "Init should succeed with user override")

	// Verify the manifest contains the override user ID
	manifest := loadManifestOrFail(t, service.configPath)

	found := false
	for _, member := range manifest.Members {
		if member.ID == expectedUserID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected manifest to contain member with override user ID '%s'", expectedUserID)
	}
}
