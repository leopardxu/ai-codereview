package tools

import "strings"

// FileFilter provides common file filtering logic
type FileFilter struct{}

// ShouldSkipFile returns true if the file should be skipped from review
func (f *FileFilter) ShouldSkipFile(path string) bool {
	// Skip binary files
	if f.IsBinary(path) {
		return true
	}

	// Skip special files
	if f.IsSpecialFile(path) {
		return true
	}

	return false
}

// IsBinary checks if a file is a binary file
func (f *FileFilter) IsBinary(path string) bool {
	binaryExtensions := []string{
		".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico",
		".jar", ".so", ".dll", ".exe", ".bin",
		".zip", ".tar", ".gz", ".7z", ".rar",
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
		".o", ".a", ".pyc", ".class",
	}

	for _, ext := range binaryExtensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	return false
}

// IsSpecialFile checks if a file is a special file that should be skipped
func (f *FileFilter) IsSpecialFile(path string) bool {
	specialFiles := []string{
		"/COMMIT_MSG",
		"/MERGE_LIST",
		".gitignore",
		".gitmodules",
	}

	for _, special := range specialFiles {
		if path == special || strings.HasSuffix(path, special) {
			return true
		}
	}

	return false
}

// IsSourceCode checks if a file is a source code file
func (f *FileFilter) IsSourceCode(path string) bool {
	sourceExtensions := []string{
		".c", ".h", ".cpp", ".hpp", ".cc", ".cxx",
		".java", ".kt", ".kts",
		".go",
		".py",
		".js", ".ts", ".jsx", ".tsx",
		".rs",
		".swift",
		".m", ".mm",
		".xml", ".json", ".yaml", ".yml",
		".sh", ".bash",
		".gradle", ".mk", ".cmake",
	}

	for _, ext := range sourceExtensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	return false
}
