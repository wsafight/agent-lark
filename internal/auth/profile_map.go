package auth

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectProjectRoot walks up from cwd and returns the nearest project root.
func DetectProjectRoot(cwd string) string {
	dir := cwd
	for {
		for _, marker := range []string{".git", "go.mod", "package.json"} {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd
}

// ProjectHashProfile derives a stable profile name from a project root path.
// Returns "project-" followed by the first 16 hex characters of the SHA-256 hash.
func ProjectHashProfile(projectRoot string) string {
	sum := sha256.Sum256([]byte(projectRoot))
	return fmt.Sprintf("project-%x", sum[:8])
}

// ResolveEffectiveProfile resolves profile precedence:
// explicit flag > AGENT_LARK_PROFILE > project hash > default.
func ResolveEffectiveProfile(explicitProfile string) string {
	if p := strings.TrimSpace(explicitProfile); p != "" {
		return p
	}
	if p := strings.TrimSpace(os.Getenv("AGENT_LARK_PROFILE")); p != "" {
		return p
	}
	cwd, err := os.Getwd()
	if err == nil {
		root := DetectProjectRoot(cwd)
		return ProjectHashProfile(root)
	}
	return "default"
}
