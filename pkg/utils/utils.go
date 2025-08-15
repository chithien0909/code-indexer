package utils

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// GenerateID generates a unique ID from a string
func GenerateID(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	return fmt.Sprintf("%x", hasher.Sum(nil))[:16]
}

// NormalizeLanguage normalizes language names to standard forms
func NormalizeLanguage(language string) string {
	language = strings.ToLower(strings.TrimSpace(language))
	
	// Map common variations to standard names
	languageMap := map[string]string{
		"js":         "javascript",
		"ts":         "typescript",
		"py":         "python",
		"rb":         "ruby",
		"cpp":        "c++",
		"cxx":        "c++",
		"cc":         "c++",
		"hpp":        "c++",
		"hxx":        "c++",
		"cs":         "csharp",
		"c#":         "csharp",
		"kt":         "kotlin",
		"rs":         "rust",
		"go":         "go",
		"java":       "java",
		"php":        "php",
		"swift":      "swift",
		"scala":      "scala",
		"clj":        "clojure",
		"cljs":       "clojure",
		"hs":         "haskell",
		"ml":         "ocaml",
		"sh":         "shell",
		"bash":       "shell",
		"zsh":        "shell",
		"fish":       "shell",
		"ps1":        "powershell",
		"sql":        "sql",
		"r":          "r",
		"m":          "matlab",
		"dart":       "dart",
		"lua":        "lua",
		"perl":       "perl",
		"pl":         "perl",
	}

	if normalized, exists := languageMap[language]; exists {
		return normalized
	}

	return language
}

// GetLanguageFromExtension determines programming language from file extension
func GetLanguageFromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	extensionMap := map[string]string{
		".go":     "go",
		".py":     "python",
		".js":     "javascript",
		".ts":     "typescript",
		".jsx":    "javascript",
		".tsx":    "typescript",
		".java":   "java",
		".cpp":    "c++",
		".cxx":    "c++",
		".cc":     "c++",
		".c":      "c",
		".h":      "c",
		".hpp":    "c++",
		".hxx":    "c++",
		".rs":     "rust",
		".rb":     "ruby",
		".php":    "php",
		".cs":     "csharp",
		".kt":     "kotlin",
		".swift":  "swift",
		".scala":  "scala",
		".clj":    "clojure",
		".cljs":   "clojure",
		".hs":     "haskell",
		".ml":     "ocaml",
		".sh":     "shell",
		".bash":   "shell",
		".zsh":    "shell",
		".fish":   "shell",
		".ps1":    "powershell",
		".sql":    "sql",
		".r":      "r",
		".m":      "matlab",
		".dart":   "dart",
		".lua":    "lua",
		".perl":   "perl",
		".pl":     "perl",
		".vim":    "vim",
		".yaml":   "yaml",
		".yml":    "yaml",
		".json":   "json",
		".xml":    "xml",
		".html":   "html",
		".htm":    "html",
		".css":    "css",
		".scss":   "scss",
		".sass":   "sass",
		".less":   "less",
		".md":     "markdown",
		".tex":    "latex",
		".dockerfile": "dockerfile",
	}

	if lang, exists := extensionMap[ext]; exists {
		return lang
	}

	// Special cases for files without extensions
	basename := strings.ToLower(filepath.Base(filename))
	switch basename {
	case "dockerfile":
		return "dockerfile"
	case "makefile":
		return "makefile"
	case "rakefile":
		return "ruby"
	case "gemfile":
		return "ruby"
	case "podfile":
		return "ruby"
	}

	return "unknown"
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	
	if maxLen <= 3 {
		return s[:maxLen]
	}
	
	return s[:maxLen-3] + "..."
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

// FormatFileSize formats a file size in bytes to a human-readable string
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// SanitizeFilename removes or replaces characters that are not safe for filenames
func SanitizeFilename(filename string) string {
	// Replace unsafe characters with underscores
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	
	// Remove leading/trailing spaces and dots
	result = strings.Trim(result, " .")
	
	// Ensure it's not empty
	if result == "" {
		result = "unnamed"
	}
	
	return result
}

// IsTextFile determines if a file is likely to contain text based on its extension
func IsTextFile(filename string) bool {
	textExtensions := map[string]bool{
		".txt": true, ".md": true, ".rst": true, ".asciidoc": true,
		".go": true, ".py": true, ".js": true, ".ts": true, ".java": true,
		".cpp": true, ".c": true, ".h": true, ".hpp": true, ".rs": true,
		".rb": true, ".php": true, ".cs": true, ".kt": true, ".swift": true,
		".scala": true, ".clj": true, ".hs": true, ".ml": true, ".sh": true,
		".bash": true, ".zsh": true, ".fish": true, ".ps1": true, ".sql": true,
		".r": true, ".m": true, ".dart": true, ".lua": true, ".perl": true,
		".pl": true, ".vim": true, ".yaml": true, ".yml": true, ".json": true,
		".xml": true, ".html": true, ".htm": true, ".css": true, ".scss": true,
		".sass": true, ".less": true, ".tex": true, ".dockerfile": true,
		".gitignore": true, ".gitattributes": true, ".editorconfig": true,
		".ini": true, ".cfg": true, ".conf": true, ".config": true,
		".toml": true, ".properties": true, ".env": true,
	}

	ext := strings.ToLower(filepath.Ext(filename))
	return textExtensions[ext]
}

// ExtractSnippet extracts a snippet around a specific line number
func ExtractSnippet(content string, lineNumber, contextLines int) string {
	lines := strings.Split(content, "\n")
	
	start := lineNumber - contextLines - 1
	if start < 0 {
		start = 0
	}
	
	end := lineNumber + contextLines
	if end > len(lines) {
		end = len(lines)
	}
	
	snippet := strings.Join(lines[start:end], "\n")
	return strings.TrimSpace(snippet)
}

// CountLines counts the number of lines in a string
func CountLines(content string) int {
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}

// RemoveCommonIndentation removes common leading whitespace from all lines
func RemoveCommonIndentation(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	// Find minimum indentation (excluding empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		indent := 0
		for _, char := range line {
			if char == ' ' {
				indent++
			} else if char == '\t' {
				indent += 4 // Treat tab as 4 spaces
			} else {
				break
			}
		}
		
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return content
	}

	// Remove common indentation
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
			continue
		}
		
		// Remove minIndent characters (handling tabs)
		removed := 0
		newLine := ""
		for i, char := range line {
			if removed >= minIndent {
				newLine = line[i:]
				break
			}
			
			if char == ' ' {
				removed++
			} else if char == '\t' {
				removed += 4
			} else {
				newLine = line[i:]
				break
			}
		}
		
		result = append(result, newLine)
	}

	return strings.Join(result, "\n")
}
