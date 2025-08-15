package analyze

import (
	"path/filepath"
	"strings"
)

// Filter handles filtering of unwanted files and directories
type Filter struct {
	// Hard-coded patterns to ignore
	ignoreDirs       []string
	ignoreFiles      []string
	ignoreExtensions []string
	ignorePrefixes   []string
}

// NewFilter creates a new filter with default ignore patterns
func NewFilter() *Filter {
	return &Filter{
		// Directories to ignore (exact match)
		ignoreDirs: []string{
			// Version control
			".git",
			".svn",
			".hg",
			".bzr",

			// Dependencies
			"node_modules",
			"vendor",
			"venv",
			".venv",
			"env",
			".env.local",
			"virtualenv",
			"bower_components",
			"jspm_packages",

			// Build outputs
			"dist",
			"build",
			"target",
			"out",
			"output",
			"bin",
			"obj",
			"_build",

			// Cache and compiled
			"__pycache__",
			".cache",
			".pytest_cache",
			".mypy_cache",
			".tox",
			".coverage",
			".nyc_output",

			// IDE and editors
			".idea",
			".vscode",
			".vs",
			".sublime-workspace",
			"*.xcworkspace",
			".project",
			".classpath",
			".settings",

			// Temporary
			"tmp",
			"temp",
			".tmp",
			".temp",

			// Package managers
			".bundle",
			".dart_tool",
			".packages",
			".pub-cache",
			".pub",
			"Pods",

			// Other
			"coverage",
			"htmlcov",
			".hypothesis",
			".phpunit.result.cache",
			"*.egg-info",
			".eggs",
		},

		// Files to ignore (exact match)
		ignoreFiles: []string{
			// OS files
			".DS_Store",
			"Thumbs.db",
			"desktop.ini",

			// Lock files (might want to keep some)
			"package-lock.json",
			"yarn.lock",
			"pnpm-lock.yaml",
			"Gemfile.lock",
			"poetry.lock",
			"Pipfile.lock",
			"composer.lock",
			"Cargo.lock",
			"go.sum",

			// Environment files with secrets
			".env",
			".env.local",
			".env.production",
			".env.development",
		},

		// Extensions to ignore
		ignoreExtensions: []string{
			// Compiled/binary
			".exe", ".dll", ".so", ".dylib", ".a",
			".o", ".obj", ".class", ".jar", ".war",
			".pyc", ".pyo", ".pyd",

			// Archives
			".zip", ".tar", ".gz", ".bz2", ".7z", ".rar",

			// Database
			".db", ".sqlite", ".sqlite3",

			// Logs
			".log",

			// Backup
			".bak", ".backup", ".old", ".orig",
			"~", // Backup files ending with ~

			// IDE specific
			".iml", ".suo", ".user",
			".swp", ".swo", ".swn", // Vim swap
		},

		// Prefixes to ignore
		ignorePrefixes: []string{
			".", // Hidden files/folders (but we check specific ones above)
			"~", // Temporary files
			"#", // Emacs temp files
		},
	}
}

// ShouldFilter returns true if the given path should be filtered out
func (f *Filter) ShouldFilter(name string, isDir bool) bool {
	// Check for empty name
	if name == "" {
		return true
	}

	// Get lowercase name for case-insensitive matching
	lowerName := strings.ToLower(name)

	if isDir {
		// Check directory ignore list
		for _, ignore := range f.ignoreDirs {
			if lowerName == strings.ToLower(ignore) {
				return true
			}
		}
	} else {
		// Check file ignore list
		for _, ignore := range f.ignoreFiles {
			if lowerName == strings.ToLower(ignore) {
				return true
			}
		}

		// Check extensions
		for _, ext := range f.ignoreExtensions {
			if strings.HasSuffix(lowerName, strings.ToLower(ext)) {
				return true
			}
		}
	}

	// Check prefixes (for both files and directories)
	// But be careful with "." prefix - we want to ignore most hidden files,
	// but some like .gitignore or .dockerignore might be wanted
	if strings.HasPrefix(name, ".") {
		// Allow some common config files that define project structure
		allowedDotFiles := []string{
			".gitignore",
			".dockerignore",
			".eslintrc",
			".prettierrc",
			".editorconfig",
			".gitattributes",
			".npmrc",
			".nvmrc",
			".ruby-version",
			".python-version",
			".tool-versions",
		}

		if !isDir {
			for _, allowed := range allowedDotFiles {
				if lowerName == strings.ToLower(allowed) ||
					strings.HasPrefix(lowerName, strings.ToLower(allowed)+".") {
					return false // Don't filter these
				}
			}
		}

		// Filter other dot files/folders
		return true
	}

	// Check other prefixes
	for _, prefix := range f.ignorePrefixes[1:] { // Skip "." as we handled it above
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	// Special case: filter files/dirs with certain patterns
	// e.g., anything ending with .min.js, .bundle.js, etc.
	if !isDir {
		if strings.Contains(lowerName, ".min.") ||
			strings.Contains(lowerName, ".bundle.") ||
			strings.Contains(lowerName, ".packed.") ||
			strings.Contains(lowerName, ".compiled.") {
			return true
		}
	}

	return false
}

// GetFilteredExtension checks if an extension should be filtered
func (f *Filter) GetFilteredExtension(filename string) (string, bool) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "", false
	}

	lowerExt := strings.ToLower(ext)
	for _, ignored := range f.ignoreExtensions {
		if lowerExt == strings.ToLower(ignored) {
			return ext, true
		}
	}

	return ext, false
}
