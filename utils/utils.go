package utils

import "path/filepath"

func PrefixEnvVars(prefix, name string) []string {
	return []string{prefix + "_" + name}
}

func ConvertToAbsPath(path string) string {
	absPath, _ := filepath.Abs(path)
	return absPath
}
