package core

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// TeamService handles all SOPS team management operations
type TeamService struct {
	sopsPath   string
	configPath string
	secretsDir string
	output     io.Writer
}

// NewTeamService creates a new team service instance
func NewTeamService(sopsPath string) *TeamService {
	return &TeamService{
		sopsPath:   sopsPath,
		configPath: "sopsistry.yaml",
		secretsDir: ".secrets",
		output:     os.Stdout,
	}
}

// Init initializes a new SOPS team configuration
func (s *TeamService) Init(force bool) error {
	if err := s.checkInitialization(force); err != nil {
		return err
	}

	if err := s.setupEnvironment(); err != nil {
		return err
	}

	publicKey, err := s.setupAgeKey()
	if err != nil {
		return err
	}

	memberID, err := s.getCurrentMemberID()
	if err != nil {
		return err
	}

	manifest := s.createInitialManifest(memberID, publicKey, time.Now().UTC())
	if err := manifest.Save(s.configPath); err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	s.printInitializationSuccess(force, memberID)
	s.showSOPSCoexistenceAdvice()
	s.printNextSteps()

	return nil
}

func (s *TeamService) checkInitialization(force bool) error {
	if !force {
		if _, err := os.Stat(s.configPath); err == nil {
			return fmt.Errorf("sopsistry.yaml already exists (use --force to overwrite)")
		}
	}
	return nil
}

func (s *TeamService) setupEnvironment() error {
	if err := os.MkdirAll(s.secretsDir, 0o700); err != nil {
		return fmt.Errorf("failed to create .secrets directory: %w", err)
	}

	if err := s.ensureGitignore(); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	return nil
}

func (s *TeamService) setupAgeKey() (string, error) {
	keyPath := filepath.Join(s.secretsDir, "key.txt")

	if _, err := os.Stat(keyPath); err != nil {
		return s.generateAgeKey(keyPath)
	}

	_, _ = fmt.Fprintf(s.output, "Using existing age key at %s\n", keyPath)
	publicKey, err := s.getPublicKeyFromPrivateKey(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to extract public key: %w", err)
	}
	return publicKey, nil
}

func (s *TeamService) getCurrentMemberID() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	memberID := currentUser.Username
	if memberID == "" {
		memberID = "me"
	}
	return memberID, nil
}

func (s *TeamService) createInitialManifest(memberID, publicKey string, created time.Time) *Manifest {
	return &Manifest{
		Members: []Member{
			{
				ID:      memberID,
				AgeKey:  publicKey,
				Created: created,
			},
		},
		Scopes: []Scope{
			{
				Name:     "default",
				Patterns: []string{"*.sops.yaml", "*.sops.json", "secrets/*"},
				Members:  []string{memberID},
			},
		},
		Settings: Settings{
			SopsVersion:   "3.8.0",
			MaxKeyAgeDays: 180, // default 6 months
		},
	}
}

func (s *TeamService) printInitializationSuccess(force bool, memberID string) {
	if force {
		_, _ = fmt.Fprintf(s.output, "Re-initialized SOPS team configuration (force mode)\n")
	} else {
		_, _ = fmt.Fprintf(s.output, "Initialized SOPS team configuration\n")
	}
	_, _ = fmt.Fprintf(s.output, "📄  Created %s\n", s.configPath)
	_, _ = fmt.Fprintf(s.output, "🔒  Created %s directory\n", s.secretsDir)

	keyPath := filepath.Join(s.secretsDir, "key.txt")
	if _, err := os.Stat(keyPath); err == nil {
		_, _ = fmt.Fprintf(s.output, "🗝️   Using existing age key at %s\n", keyPath)
	}
	_, _ = fmt.Fprintf(s.output, "🧑‍💻  Added %s as team member\n", memberID)
}

func (s *TeamService) showSOPSCoexistenceAdvice() {
	detector := NewSOPSDetector()
	sopsInfo, err := detector.DetectSOPSConfig()
	if err == nil && sopsInfo.Exists {
		_, _ = fmt.Fprintf(s.output, "\n%s\n", sopsInfo.GetCoexistenceAdvice())
	}
}

func (s *TeamService) printNextSteps() {
	_, _ = fmt.Fprintf(s.output, "\n🚀 Next steps:\n")
	_, _ = fmt.Fprintf(s.output, "1. Encrypt files: sistry encrypt <file> or sistry encrypt --regex '^(password|key)' <file>\n")
	_, _ = fmt.Fprintf(s.output, "2. Add more team members: sistry add-member <id> --key <age-pubkey>\n")
	_, _ = fmt.Fprintf(s.output, "3. Review planned changes: sistry plan\n")
	_, _ = fmt.Fprintf(s.output, "4. Apply changes: sistry apply\n")
}

// Plan shows what changes would be made
func (s *TeamService) Plan(noColor bool) error {
	manifest, err := LoadManifest(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	planner := NewPlanner(s.sopsPath)
	plan, err := planner.ComputePlan(manifest)
	if err != nil {
		return fmt.Errorf("failed to compute plan: %w", err)
	}

	plan.Display(noColor)
	return nil
}

// Apply executes planned changes
func (s *TeamService) Apply(requireCleanGit, skipConfirmation bool) error {
	if requireCleanGit {
		if err := s.checkGitClean(); err != nil {
			return err
		}
	}

	manifest, err := LoadManifest(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	planner := NewPlanner(s.sopsPath)
	plan, err := planner.ComputePlan(manifest)
	if err != nil {
		return fmt.Errorf("failed to compute plan: %w", err)
	}

	if len(plan.Actions) == 0 {
		_, _ = fmt.Fprintln(s.output, "No changes to apply")
		return nil
	}

	if !skipConfirmation {
		plan.Display(false)
		fmt.Print("\nApply these changes? [y/N]: ")
		var response string
		_, _ = fmt.Scanln(&response) // User input, ignore errors
		if response != "y" && response != "Y" {
			_, _ = fmt.Fprintln(s.output, "Cancelled")
			return nil
		}
	}

	executor := NewExecutor(s.sopsPath)
	return executor.Execute(plan)
}

// AddMember adds a new team member
func (s *TeamService) AddMember(id, ageKey string) error {
	manifest, err := LoadManifest(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	for _, member := range manifest.Members {
		if member.ID == id {
			return fmt.Errorf("member %s already exists", id)
		}
	}

	manifest.Members = append(manifest.Members, Member{
		ID:      id,
		AgeKey:  ageKey,
		Created: time.Now().UTC(),
	})

	for i := range manifest.Scopes {
		if manifest.Scopes[i].Name == "default" {
			manifest.Scopes[i].Members = append(manifest.Scopes[i].Members, id)
			break
		}
	}

	if err := manifest.Save(s.configPath); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	_, _ = fmt.Fprintf(s.output, "Added member %s to team\n", id)
	_, _ = fmt.Fprintln(s.output, "Run 'sistry plan' to see changes, then 'sistry apply' to re-encrypt files")
	return nil
}

// RemoveMember removes a team member
func (s *TeamService) RemoveMember(id string) error {
	manifest, err := LoadManifest(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	if err := s.removeMemberFromManifest(manifest, id); err != nil {
		return err
	}

	s.removeMemberFromAllScopes(manifest, id)

	if err := manifest.Save(s.configPath); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	s.printRemovalSuccess(id)
	return nil
}

func (s *TeamService) removeMemberFromManifest(manifest *Manifest, id string) error {
	for i, member := range manifest.Members {
		if member.ID == id {
			manifest.Members = append(manifest.Members[:i], manifest.Members[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("member %s not found", id)
}

func (s *TeamService) removeMemberFromAllScopes(manifest *Manifest, id string) {
	for i := range manifest.Scopes {
		s.removeMemberFromScope(&manifest.Scopes[i], id)
	}
}

func (s *TeamService) removeMemberFromScope(scope *Scope, id string) {
	for j, memberID := range scope.Members {
		if memberID == id {
			scope.Members = append(scope.Members[:j], scope.Members[j+1:]...)
			break
		}
	}
}

func (s *TeamService) printRemovalSuccess(id string) {
	_, _ = fmt.Fprintf(s.output, "Removed member %s from team\n", id)
	_, _ = fmt.Fprintln(s.output, "Run 'sistry plan' to see changes, then 'sistry apply' to re-encrypt files")
}

// List displays current team configuration
func (s *TeamService) List(jsonOutput bool) error {
	manifest, err := LoadManifest(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	if jsonOutput {
		return manifest.DisplayJSON()
	}

	manifest.Display()
	return nil
}

// EncryptFile encrypts a file using the current team configuration
func (s *TeamService) EncryptFile(filePath string, inPlace bool, regex string) error {
	manifest, err := LoadManifest(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	var ageKeys []string //nolint:prealloc // Small team sizes, optimization not worth it
	for _, member := range manifest.Members {
		ageKeys = append(ageKeys, member.AgeKey)
	}

	if len(ageKeys) == 0 {
		return fmt.Errorf("no team members found in configuration")
	}

	encryptor := NewEncryptor(s.sopsPath)
	return encryptor.EncryptFile(filePath, ageKeys, inPlace, regex)
}

// DecryptFile decrypts a SOPS-encrypted file
func (s *TeamService) DecryptFile(filePath string, inPlace bool) error {
	keyPath := filepath.Join(s.secretsDir, "key.txt")

	decryptor := NewDecryptor(s.sopsPath)
	return decryptor.DecryptFile(filePath, keyPath, inPlace)
}

// ShowSOPSCommand displays the SOPS command with proper environment variables
func (s *TeamService) ShowSOPSCommand(args []string, execute bool) error {
	manifest, err := LoadManifest(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	var ageKeys []string //nolint:prealloc // Small team sizes, optimization not worth it
	for _, member := range manifest.Members {
		ageKeys = append(ageKeys, member.AgeKey)
	}

	if len(ageKeys) == 0 {
		return fmt.Errorf("no team members found in configuration")
	}

	er := NewSOPSHelper(s.sopsPath, s.secretsDir)
	return er.ShowCommand(args, ageKeys, execute)
}

// RotateKey rotates the current user's age key
func (s *TeamService) RotateKey(force bool) error {
	manifest, currentMember, err := s.prepareKeyRotation(force)
	if err != nil {
		return err
	}

	keyPath := filepath.Join(s.secretsDir, "key.txt")
	backupPath := keyPath + ".backup"

	if err := s.backupCurrentKey(keyPath, backupPath); err != nil {
		return err
	}
	defer func() { _ = os.Remove(backupPath) }()

	return s.executeKeyRotation(manifest, currentMember, keyPath, backupPath)
}

func (s *TeamService) prepareKeyRotation(force bool) (*Manifest, *Member, error) {
	manifest, err := LoadManifest(s.configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	currentUser, err := s.getCurrentMemberID()
	if err != nil {
		return nil, nil, err
	}

	currentMember := s.findMemberByID(manifest, currentUser)
	if currentMember == nil {
		return nil, nil, fmt.Errorf("current user %s not found in team", currentUser)
	}

	if !force {
		if err := s.checkKeyExpiry(currentMember, manifest.Settings.MaxKeyAgeDays); err != nil {
			return nil, nil, err
		}
	}

	return manifest, currentMember, nil
}

func (s *TeamService) findMemberByID(manifest *Manifest, userID string) *Member {
	for i := range manifest.Members {
		if manifest.Members[i].ID == userID {
			return &manifest.Members[i]
		}
	}
	return nil
}

func (s *TeamService) executeKeyRotation(manifest *Manifest, currentMember *Member, keyPath, backupPath string) error {
	newPublicKey, err := s.generateAgeKey(keyPath)
	if err != nil {
		return s.handleRotationError("failed to generate new key", err, keyPath, backupPath)
	}

	currentMember.AgeKey = newPublicKey
	currentMember.Created = time.Now().UTC()

	if err := manifest.Save(s.configPath); err != nil {
		return s.handleRotationError("failed to save manifest", err, keyPath, backupPath)
	}

	if err := s.reencryptAllFiles(manifest, keyPath, backupPath); err != nil {
		return err
	}

	s.printRotationSuccess(currentMember)
	return nil
}

func (s *TeamService) reencryptAllFiles(manifest *Manifest, keyPath, backupPath string) error {
	planner := NewPlanner(s.sopsPath)
	plan, err := planner.ComputePlan(manifest)
	if err != nil {
		return s.handleRotationError("failed to compute plan", err, keyPath, backupPath)
	}

	executor := NewExecutor(s.sopsPath)
	if err := executor.Execute(plan); err != nil {
		return s.handleRotationError("failed to re-encrypt files", err, keyPath, backupPath)
	}

	return nil
}

func (s *TeamService) printRotationSuccess(member *Member) {
	_, _ = fmt.Fprintf(s.output, "🔄 Successfully rotated key for %s\n", member.ID)
	_, _ = fmt.Fprintf(s.output, "📅 New key created: %s\n", member.Created.Format("2006-01-02T15:04:05Z"))
}

func (s *TeamService) backupCurrentKey(keyPath, backupPath string) error {
	if err := validateFilePath(keyPath); err != nil {
		return fmt.Errorf("invalid key path: %w", err)
	}
	if err := validateFilePath(backupPath); err != nil {
		return fmt.Errorf("invalid backup path: %w", err)
	}

	if _, err := os.Stat(keyPath); err == nil {
		data, err := os.ReadFile(keyPath) //nolint:gosec // Path validated by validateFilePath
		if err != nil {
			return fmt.Errorf("failed to read current key: %w", err)
		}
		if err := os.WriteFile(backupPath, data, 0o600); err != nil {
			return fmt.Errorf("failed to backup key: %w", err)
		}
	}
	return nil
}

func (s *TeamService) handleRotationError(msg string, err error, keyPath, backupPath string) error {
	// Restore backup
	if _, backupErr := os.Stat(backupPath); backupErr == nil {
		if data, readErr := os.ReadFile(backupPath); readErr == nil { //nolint:gosec // Path validated during backup creation
			_ = os.WriteFile(keyPath, data, 0o600)
		}
	}
	return fmt.Errorf("%s: %w", msg, err)
}

func (s *TeamService) checkKeyExpiry(member *Member, maxAgeDays int) error {
	if maxAgeDays <= 0 {
		maxAgeDays = 180 // default 6 months
	}

	age := time.Since(member.Created)
	maxAge := time.Duration(maxAgeDays) * 24 * time.Hour

	if age > maxAge {
		return fmt.Errorf("key has expired (age: %d days, max: %d days). Use --force to rotate anyway",
			int(age.Hours()/24), maxAgeDays)
	}

	return nil
}

// CheckKeyExpiry checks if any keys are expired or expiring soon
func (s *TeamService) CheckKeyExpiry() error {
	manifest, err := LoadManifest(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	maxAgeDays := manifest.Settings.MaxKeyAgeDays
	if maxAgeDays <= 0 {
		maxAgeDays = 180 // default 6 months
	}

	warnings := 0
	errors := 0
	now := time.Now()

	for _, member := range manifest.Members {
		age := now.Sub(member.Created)
		maxAge := time.Duration(maxAgeDays) * 24 * time.Hour
		warningThreshold := maxAge - (14 * 24 * time.Hour) // 2 weeks before expiry

		switch {
		case age > maxAge:
			_, _ = fmt.Fprintf(s.output, "❌ %s: key expired %d days ago (created: %s)\n",
				member.ID, int((age-maxAge).Hours()/24), member.Created.Format("2006-01-02"))
			errors++
		case age > warningThreshold:
			daysUntilExpiry := int((maxAge - age).Hours() / 24)
			_, _ = fmt.Fprintf(s.output, "⚠️  %s: key expires in %d days (created: %s)\n",
				member.ID, daysUntilExpiry, member.Created.Format("2006-01-02"))
			warnings++
		default:
			daysAge := int(age.Hours() / 24)
			_, _ = fmt.Fprintf(s.output, "✅ %s: key is %d days old (created: %s)\n",
				member.ID, daysAge, member.Created.Format("2006-01-02"))
		}
	}

	if errors > 0 {
		_, _ = fmt.Fprintf(s.output, "\n%d expired keys found. Run 'sopsistry rotate-key' to rotate.\n", errors)
	}
	if warnings > 0 {
		_, _ = fmt.Fprintf(s.output, "\n%d keys expiring soon. Consider running 'sopsistry rotate-key'.\n", warnings)
	}

	return nil
}
