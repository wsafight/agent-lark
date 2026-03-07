package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func projectMapPath() string {
	return filepath.Join(HomeDir(), "project-map.json")
}

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

// MappedProfile returns the bound profile for a project root.
func MappedProfile(projectRoot string) string {
	if projectRoot == "" {
		return ""
	}
	data, err := os.ReadFile(projectMapPath())
	if err != nil {
		return ""
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	return strings.TrimSpace(m[projectRoot])
}

// ResolveEffectiveProfile resolves profile precedence:
// explicit flag > AGENT_LARK_PROFILE > project binding > default.
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
		if p := MappedProfile(root); p != "" {
			return p
		}
	}
	return "default"
}

// SaveProjectBinding binds a project root to a profile name.
func SaveProjectBinding(projectRoot, profile string) error {
	if projectRoot == "" || profile == "" {
		return nil
	}

	path := projectMapPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	m := map[string]string{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &m)
	}

	m[projectRoot] = profile
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
