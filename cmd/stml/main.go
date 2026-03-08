package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/geul-org/stml/generator"
	"github.com/geul-org/stml/parser"
	"github.com/geul-org/stml/validator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: stml <command> [dir]")
		fmt.Fprintln(os.Stderr, "commands: parse, validate, gen")
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "parse":
		runParse()
	case "validate":
		runValidate()
	case "gen":
		runGen()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func runParse() {
	dir := "specs/frontend"
	if len(os.Args) >= 3 {
		dir = os.Args[2]
	}

	pages, err := parser.ParseDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(pages)
}

func runValidate() {
	projectRoot := "specs/dummy-study"
	if len(os.Args) >= 3 {
		projectRoot = os.Args[2]
	}

	frontendDir := filepath.Join(projectRoot, "frontend")
	pages, err := parser.ParseDir(frontendDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	errs := validator.Validate(pages, projectRoot)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e.Error())
		}
		os.Exit(1)
	}

	fmt.Printf("validate OK (%d pages)\n", len(pages))
}

func runGen() {
	projectRoot := "specs/dummy-study"
	if len(os.Args) >= 3 {
		projectRoot = os.Args[2]
	}
	outDir := "artifacts/dummy-study/frontend"
	if len(os.Args) >= 4 {
		outDir = os.Args[3]
	}

	frontendDir := filepath.Join(projectRoot, "frontend")
	pages, err := parser.ParseDir(frontendDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Validate first
	errs := validator.Validate(pages, projectRoot)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e.Error())
		}
		fmt.Fprintln(os.Stderr, "validation failed, codegen aborted")
		os.Exit(1)
	}

	if err := generator.Generate(pages, frontendDir, outDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("generated %d files in %s\n", len(pages), outDir)
}
