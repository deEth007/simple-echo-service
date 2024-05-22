package utils

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func PrefixEnvVars(prefix, name string) []string {
	return []string{prefix + "_" + name}
}

func ConvertToAbsPath(path string) string {
	absPath, _ := filepath.Abs(path)
	return absPath
}

func ReadJsonUnknown(path string) map[string]interface{} {
	jsonFile, _ := os.Open(path)
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var result map[string]interface{}

	json.Unmarshal(byteValue, &result)
	return result
}
