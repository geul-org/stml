package generator

import "github.com/geul-org/stml/parser"

// Target abstracts the code generation backend.
// Implement this interface to support a new framework (e.g. Vue, Svelte).
type Target interface {
	GeneratePage(page parser.PageSpec, specsDir string, opts GenerateOptions) string
	FileExtension() string
	Dependencies(pages []parser.PageSpec) map[string]string
}

// DefaultTarget returns the built-in React/TSX target.
func DefaultTarget() Target {
	return &ReactTarget{}
}

// compile-time check
var _ Target = (*ReactTarget)(nil)
