package main

import "github.com/wsafight/agent-lark/internal/auth"

func detectProjectRoot(cwd string) string {
	return auth.DetectProjectRoot(cwd)
}

func mappedProfile(projectRoot string) string {
	return auth.MappedProfile(projectRoot)
}

func resolveEffectiveProfile(explicitProfile, projectRoot string) string {
	_ = projectRoot
	return auth.ResolveEffectiveProfile(explicitProfile)
}

func saveProjectBinding(projectRoot, profile string) error {
	return auth.SaveProjectBinding(projectRoot, profile)
}
