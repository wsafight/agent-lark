package main

import "github.com/wsafight/agent-lark/internal/auth"

func detectProjectRoot(cwd string) string {
	return auth.DetectProjectRoot(cwd)
}

func resolveEffectiveProfile(explicitProfile string) string {
	return auth.ResolveEffectiveProfile(explicitProfile)
}

func saveProjectBinding(projectRoot, profile string) error {
	return auth.SaveProjectBinding(projectRoot, profile)
}
