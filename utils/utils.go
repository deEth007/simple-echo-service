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

func ReadJson[T any](path string, data *T) {
	jsonFile, _ := os.Open(path)
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	json.Unmarshal(byteValue, data)
}

func WriteJson[T any](path string, data T) {
	file, _ := os.Create(path)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}
