package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/park-jun-woo/stml/parser"
)

// GenerateOptions configures code generation behavior.
type GenerateOptions struct {
	APIImportPath string // import path for api module (default: "@/lib/api")
	UseClient     bool   // emit 'use client' directive (default: true)
}

// GenerateResult contains generation output metadata.
type GenerateResult struct {
	Pages        int
	Dependencies map[string]string // package name → version range
}

// DefaultOptions returns GenerateOptions with default values.
func DefaultOptions() GenerateOptions {
	return GenerateOptions{
		APIImportPath: "@/lib/api",
		UseClient:     true,
	}
}

func mergeOpt(base, override GenerateOptions) GenerateOptions {
	if override.APIImportPath != "" {
		base.APIImportPath = override.APIImportPath
	}
	base.UseClient = override.UseClient
	return base
}

// Generate produces framework-specific files using the default target.
func Generate(pages []parser.PageSpec, specsDir, outDir string, opts ...GenerateOptions) (*GenerateResult, error) {
	return GenerateWith(DefaultTarget(), pages, specsDir, outDir, opts...)
}

// GeneratePage generates source code for a single page using the default target.
func GeneratePage(page parser.PageSpec, specsDir string, opts ...GenerateOptions) string {
	opt := DefaultOptions()
	if len(opts) > 0 {
		opt = mergeOpt(opt, opts[0])
	}
	return DefaultTarget().GeneratePage(page, specsDir, opt)
}

// GenerateWith produces files using the given Target.
func GenerateWith(t Target, pages []parser.PageSpec, specsDir, outDir string, opts ...GenerateOptions) (*GenerateResult, error) {
	opt := DefaultOptions()
	if len(opts) > 0 {
		opt = mergeOpt(opt, opts[0])
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", outDir, err)
	}

	for _, page := range pages {
		code := t.GeneratePage(page, specsDir, opt)
		path := filepath.Join(outDir, page.Name+t.FileExtension())
		if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", path, err)
		}
	}

	return &GenerateResult{
		Pages:        len(pages),
		Dependencies: t.Dependencies(pages),
	}, nil
}

// --- common utils ---

// toComponentName converts "my-reservations-page" to "MyReservationsPage".
func toComponentName(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		parts[i] = toUpperFirst(p)
	}
	return strings.Join(parts, "")
}

// collectAllActions walks the ChildNode tree and collects all ActionBlocks.
func collectAllActions(nodes []parser.ChildNode) []parser.ActionBlock {
	var actions []parser.ActionBlock
	for _, ch := range nodes {
		switch ch.Kind {
		case "action":
			actions = append(actions, *ch.Action)
		case "fetch":
			actions = append(actions, collectAllActions(ch.Fetch.Children)...)
		case "state":
			actions = append(actions, collectAllActions(ch.State.Children)...)
		case "static":
			actions = append(actions, collectAllActions(ch.Static.Children)...)
		case "each":
			actions = append(actions, collectAllActions(ch.Each.Children)...)
		}
	}
	return actions
}

// deduplicateActions removes duplicate actions by OperationID.
func deduplicateActions(actions []parser.ActionBlock) []parser.ActionBlock {
	seen := map[string]bool{}
	var result []parser.ActionBlock
	for _, a := range actions {
		if !seen[a.OperationID] {
			seen[a.OperationID] = true
			result = append(result, a)
		}
	}
	return result
}

// collectAllParams gathers all ParamBinds from the page.
func collectAllParams(page parser.PageSpec) []parser.ParamBind {
	var params []parser.ParamBind
	for _, f := range page.Fetches {
		params = collectFetchParamBinds(f, params)
	}
	for _, a := range page.Actions {
		params = append(params, a.Params...)
	}
	for _, a := range collectAllActions(page.Children) {
		params = append(params, a.Params...)
	}
	return params
}

func collectFetchParamBinds(f parser.FetchBlock, params []parser.ParamBind) []parser.ParamBind {
	params = append(params, f.Params...)
	for _, child := range f.NestedFetches {
		params = collectFetchParamBinds(child, params)
	}
	return params
}

func collectFetchOps(f parser.FetchBlock, ops []string) []string {
	ops = append(ops, f.OperationID)
	for _, child := range f.NestedFetches {
		ops = collectFetchOps(child, ops)
	}
	return ops
}

func findRootElement(page parser.PageSpec) (string, string) {
	if len(page.Children) == 1 && page.Children[0].Kind == "static" {
		se := page.Children[0].Static
		return se.Tag, se.ClassName
	}
	return "div", ""
}
