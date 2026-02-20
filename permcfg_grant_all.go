//go:build gbkr_grant_all

package gbkr

// GrantAllPrompter implements [Prompter] by automatically granting all
// requested permissions. Only available in builds with the gbkr_grant_all
// build tag to prevent accidental inclusion in production binaries.
type GrantAllPrompter struct{}

// Prompt grants all missing permissions without user interaction.
func (GrantAllPrompter) Prompt(missing []Permission) (PermissionSet, error) {
	return PermissionSet(missing), nil
}
