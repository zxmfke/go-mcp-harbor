package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BuildOutputPath constructs the output path
func BuildOutputPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	return filepath.Join(homeDir, ".go-mcp-server", ".minimax-mcp-server")
}

// BuildOutputFile constructs the output file name
func BuildOutputFile(prefix string, text string, outputPath string, ext string) string {
	timestamp := time.Now().Format("20060102_150405")
	sanitizedText := sanitizeFilename(text)
	if len(sanitizedText) > 20 {
		sanitizedText = sanitizedText[:20]
	}

	return filepath.Join(outputPath, fmt.Sprintf("%s_%s_%s.%s", prefix, sanitizedText, timestamp, ext))
}

// Sanitize filename
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	return name
}
