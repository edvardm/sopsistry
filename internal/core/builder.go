package core

import (
	"fmt"
	"os/exec"
)

// SOPSCommandState phantom type interface
type SOPSCommandState interface {
	Incomplete | WithFile | WithRecipients | Complete
}

// Phantom type markers for SOPS command builder states
type Incomplete struct{}
type WithFile struct{}
type WithRecipients struct{}
type Complete struct{}

// SOPSCommandBuilder builds SOPS commands with compile-time validation
type SOPSCommandBuilder[T SOPSCommandState] struct {
	sopsPath   string
	args       []string
	file       string
	recipients []string
}

// NewSOPSCommand creates a new SOPS command builder
func NewSOPSCommand(sopsPath string) SOPSCommandBuilder[Incomplete] {
	return SOPSCommandBuilder[Incomplete]{
		sopsPath: sopsPath,
		args:     make([]string, 0),
	}
}

// WithFile specifies the file to operate on (required)
func (b SOPSCommandBuilder[Incomplete]) WithFile(file string) SOPSCommandBuilder[WithFile] {
	return SOPSCommandBuilder[WithFile]{
		sopsPath: b.sopsPath,
		args:     b.args,
		file:     file,
	}
}

// WithRecipients specifies the age recipients (required for encryption)
func (b SOPSCommandBuilder[WithFile]) WithRecipients(recipients []string) SOPSCommandBuilder[WithRecipients] {
	return SOPSCommandBuilder[WithRecipients]{
		sopsPath:   b.sopsPath,
		args:       b.args,
		file:       b.file,
		recipients: recipients,
	}
}

// ForEncryption configures the builder for encryption operations
func (b SOPSCommandBuilder[WithRecipients]) ForEncryption() SOPSCommandBuilder[Complete] {
	args := append(b.args, "-e") //nolint:gocritic // Intentionally creating new slice for immutable builder
	for _, recipient := range b.recipients {
		args = append(args, "--age", recipient)
	}
	args = append(args, b.file)

	return SOPSCommandBuilder[Complete]{
		sopsPath: b.sopsPath,
		args:     args,
		file:     b.file,
	}
}

// ForDecryption configures the builder for decryption operations
func (b SOPSCommandBuilder[WithFile]) ForDecryption() SOPSCommandBuilder[Complete] {
	args := append(b.args, "-d", b.file) //nolint:gocritic // Intentionally creating new slice for immutable builder

	return SOPSCommandBuilder[Complete]{
		sopsPath: b.sopsPath,
		args:     args,
		file:     b.file,
	}
}

// WithInPlace adds the --in-place flag
func (b SOPSCommandBuilder[T]) WithInPlace() SOPSCommandBuilder[T] {
	args := append(b.args, "--in-place") //nolint:gocritic // Intentionally creating new slice for immutable builder
	return SOPSCommandBuilder[T]{
		sopsPath:   b.sopsPath,
		args:       args,
		file:       b.file,
		recipients: b.recipients,
	}
}

// WithRegex adds encryption regex pattern
func (b SOPSCommandBuilder[T]) WithRegex(pattern string) SOPSCommandBuilder[T] {
	if pattern != "" {
		args := append(b.args, "--encrypted-regex", pattern) //nolint:gocritic // Intentionally creating new slice for immutable builder
		return SOPSCommandBuilder[T]{
			sopsPath:   b.sopsPath,
			args:       args,
			file:       b.file,
			recipients: b.recipients,
		}
	}
	return b
}

// Build creates the final exec.Cmd (only available when Complete)
func (b SOPSCommandBuilder[Complete]) Build() *exec.Cmd {
	return exec.Command(b.sopsPath, b.args...) //nolint:gosec // sopsPath is validated by ValidSOPSPath type system
}

// Args returns the command arguments (only available when Complete)
func (b SOPSCommandBuilder[Complete]) Args() []string {
	return b.args
}

// ManifestBuilder builds manifest configurations with validation
type ManifestBuilder struct {
	members  []Member
	scopes   []Scope
	settings Settings
}

// NewManifestBuilder creates a new manifest builder
func NewManifestBuilder() *ManifestBuilder {
	return &ManifestBuilder{
		members: make([]Member, 0),
		scopes:  make([]Scope, 0),
		settings: Settings{
			MaxKeyAgeDays: DefaultMaxKeyAgeDays,
		},
	}
}

// WithMember adds a member to the manifest
func (b *ManifestBuilder) WithMember(member Member) *ManifestBuilder {
	b.members = append(b.members, member)
	return b
}

// WithScope adds a scope to the manifest
func (b *ManifestBuilder) WithScope(scope Scope) *ManifestBuilder {
	b.scopes = append(b.scopes, scope)
	return b
}

// WithSettings configures manifest settings
func (b *ManifestBuilder) WithSettings(settings Settings) *ManifestBuilder {
	b.settings = settings
	return b
}

// Build creates the final manifest with validation
func (b *ManifestBuilder) Build() Result[*Manifest] {
	if len(b.members) == 0 {
		return Err[*Manifest](NewManifestError("build", "", fmt.Errorf("manifest must have at least one member")))
	}

	manifest := &Manifest{
		Members:  b.members,
		Scopes:   b.scopes,
		Settings: b.settings,
	}

	return Ok(manifest)
}
