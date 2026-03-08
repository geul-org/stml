package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/geul-org/stml/parser"
)

// importSet collects unique imports for a page.
type importSet struct {
	react        bool
	useQuery     bool
	useMutation  bool
	useQueryClient bool
	useParams    bool
	useForm      bool
	useState     bool
	components   []string // unique component names
	customFile   string   // non-empty if custom.ts exists
}

// collectImports analyzes a PageSpec and determines required imports.
func collectImports(page parser.PageSpec, specsDir string) importSet {
	is := importSet{react: true}
	compSet := map[string]bool{}

	if len(page.Fetches) > 0 {
		is.useQuery = true
	}
	if len(page.Actions) > 0 {
		is.useMutation = true
		is.useQueryClient = true
		is.useForm = true
	}

	for _, f := range page.Fetches {
		collectFetchImports(f, &is, compSet)
	}
	for _, a := range page.Actions {
		collectActionImports(a, &is, compSet)
	}

	for comp := range compSet {
		is.components = append(is.components, comp)
	}

	// Check for custom.ts
	if specsDir != "" {
		customPath := filepath.Join(specsDir, page.Name+".custom.ts")
		if _, err := os.Stat(customPath); err == nil {
			is.customFile = page.Name + ".custom"
		}
	}

	return is
}

func collectFetchImports(f parser.FetchBlock, is *importSet, compSet map[string]bool) {
	for _, p := range f.Params {
		if strings.HasPrefix(p.Source, "route.") {
			is.useParams = true
		}
	}
	for _, c := range f.Components {
		compSet[c.Name] = true
	}
	// Phase 5: infra params need useState
	if f.Paginate || f.Sort != nil || len(f.Filters) > 0 {
		is.useState = true
	}
	for _, child := range f.NestedFetches {
		collectFetchImports(child, is, compSet)
	}
}

func collectActionImports(a parser.ActionBlock, is *importSet, compSet map[string]bool) {
	for _, p := range a.Params {
		if strings.HasPrefix(p.Source, "route.") {
			is.useParams = true
		}
	}
	for _, f := range a.Fields {
		if strings.HasPrefix(f.Tag, "data-component:") {
			comp := strings.TrimPrefix(f.Tag, "data-component:")
			compSet[comp] = true
		}
	}
}

// renderImports generates the import block string.
func renderImports(is importSet, opt GenerateOptions) string {
	var lines []string

	if opt.UseClient {
		lines = append(lines, "'use client'\n")
	}
	if is.useState {
		lines = append(lines, "import React, { useState } from 'react'")
	} else {
		lines = append(lines, "import React from 'react'")
	}

	// tanstack query
	var queryImports []string
	if is.useQuery {
		queryImports = append(queryImports, "useQuery")
	}
	if is.useMutation {
		queryImports = append(queryImports, "useMutation")
	}
	if is.useQueryClient {
		queryImports = append(queryImports, "useQueryClient")
	}
	if len(queryImports) > 0 {
		lines = append(lines, fmt.Sprintf("import { %s } from '@tanstack/react-query'", strings.Join(queryImports, ", ")))
	}

	// react-router
	if is.useParams {
		lines = append(lines, "import { useParams } from 'react-router-dom'")
	}

	// react-hook-form
	if is.useForm {
		lines = append(lines, "import { useForm } from 'react-hook-form'")
	}

	// api client
	lines = append(lines, fmt.Sprintf("import { api } from '%s'", opt.APIImportPath))

	// components
	for _, comp := range is.components {
		lines = append(lines, fmt.Sprintf("import %s from '@/components/%s'", comp, comp))
	}

	// custom.ts
	if is.customFile != "" {
		lines = append(lines, fmt.Sprintf("import * as custom from './%s'", is.customFile))
	}

	return strings.Join(lines, "\n")
}
