package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/geul-org/stml/artifacts/internal/parser"
)

// Validate checks parsed PageSpecs against an OpenAPI spec and custom.ts files.
// projectRoot is the project root containing api/openapi.yaml and frontend/.
func Validate(pages []parser.PageSpec, projectRoot string) []ValidationError {
	openAPIPath := filepath.Join(projectRoot, "api", "openapi.yaml")
	st, err := LoadOpenAPI(openAPIPath)
	if err != nil {
		return []ValidationError{{
			File:    openAPIPath,
			Attr:    "openapi",
			Message: fmt.Sprintf("OpenAPI 파일을 읽을 수 없습니다: %v", err),
		}}
	}

	frontendDir := filepath.Join(projectRoot, "frontend")
	var errs []ValidationError

	for _, page := range pages {
		// Load custom.ts for this page
		customPath := filepath.Join(frontendDir, page.Name+".custom.ts")
		cs, _ := LoadCustomTS(customPath)

		for _, f := range page.Fetches {
			errs = append(errs, validateFetchBlock(f, page.FileName, st, cs, frontendDir)...)
		}
		for _, a := range page.Actions {
			errs = append(errs, validateActionBlock(a, page.FileName, st, frontendDir)...)
		}
	}

	return errs
}

func validateFetchBlock(f parser.FetchBlock, file string, st *SymbolTable, cs *CustomSymbol, frontendDir string) []ValidationError {
	var errs []ValidationError
	attr := fmt.Sprintf("data-fetch=%q", f.OperationID)

	// 1. operationId existence
	api, ok := st.Operations[f.OperationID]
	if !ok {
		return append(errs, errOpNotFound(file, attr, f.OperationID))
	}

	// 2. HTTP method check
	if api.Method != "GET" {
		errs = append(errs, errWrongMethod(file, attr, f.OperationID, api.Method, "GET"))
	}

	// 3. parameter check
	errs = append(errs, validateParams(f.Params, f.OperationID, file, api)...)

	// 5. response bind check
	for _, b := range f.Binds {
		fieldName := b.Name
		// handle dot notation: "reservation.Status" → check "reservation" in response
		if idx := strings.IndexByte(fieldName, '.'); idx >= 0 {
			fieldName = fieldName[:idx]
		}
		if _, ok := api.ResponseFields[fieldName]; !ok {
			if cs == nil || !cs.Functions[fieldName] {
				errs = append(errs, errBindNotFound(file, f.OperationID, b.Name))
			}
		}
	}

	// 6. each array check
	for _, e := range f.Eaches {
		if fs, ok := api.ResponseFields[e.Field]; ok {
			if fs.Type != "array" {
				errs = append(errs, errEachNotArray(file, f.OperationID, e.Field))
			}
		} else {
			errs = append(errs, errEachNotFound(file, f.OperationID, e.Field))
		}
	}

	// 7. component existence
	for _, c := range f.Components {
		errs = append(errs, validateComponent(c.Name, file, frontendDir)...)
	}

	// Phase 5: infra param validation
	errs = append(errs, validateInfraParams(f, file, api)...)

	// Recurse into nested fetches
	for _, child := range f.NestedFetches {
		errs = append(errs, validateFetchBlock(child, file, st, cs, frontendDir)...)
	}

	return errs
}

func validateInfraParams(f parser.FetchBlock, file string, api APISymbol) []ValidationError {
	var errs []ValidationError

	// data-paginate requires x-pagination
	if f.Paginate && api.Pagination == nil {
		errs = append(errs, errPaginateNoExt(file, f.OperationID))
	}

	// data-sort column must be in x-sort.allowed
	if f.Sort != nil {
		if api.Sort == nil {
			errs = append(errs, errSortNotAllowed(file, f.OperationID, f.Sort.Column))
		} else if !containsStr(api.Sort.Allowed, f.Sort.Column) {
			errs = append(errs, errSortNotAllowed(file, f.OperationID, f.Sort.Column))
		}
	}

	// data-filter columns must be in x-filter.allowed
	for _, col := range f.Filters {
		if api.Filter == nil || !containsStr(api.Filter.Allowed, col) {
			errs = append(errs, errFilterNotAllowed(file, f.OperationID, col))
		}
	}

	// data-include resources must be in x-include.allowed
	for _, res := range f.Includes {
		if api.Include == nil || !containsStr(api.Include.Allowed, res) {
			errs = append(errs, errIncludeNotAllowed(file, f.OperationID, res))
		}
	}

	return errs
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func validateActionBlock(a parser.ActionBlock, file string, st *SymbolTable, frontendDir string) []ValidationError {
	var errs []ValidationError
	attr := fmt.Sprintf("data-action=%q", a.OperationID)

	// 1. operationId existence
	api, ok := st.Operations[a.OperationID]
	if !ok {
		return append(errs, errOpNotFound(file, attr, a.OperationID))
	}

	// 2. HTTP method check
	if api.Method == "GET" {
		errs = append(errs, errWrongMethod(file, attr, a.OperationID, api.Method, "POST/PUT/DELETE"))
	}

	// 3. parameter check
	errs = append(errs, validateParams(a.Params, a.OperationID, file, api)...)

	// 4. request field check
	for _, f := range a.Fields {
		fieldName := f.Name
		// skip component fields (e.g. tag="data-component:DatePicker")
		if strings.HasPrefix(f.Tag, "data-component:") {
			comp := strings.TrimPrefix(f.Tag, "data-component:")
			errs = append(errs, validateComponent(comp, file, frontendDir)...)
		}
		if _, ok := api.RequestFields[fieldName]; !ok {
			errs = append(errs, errFieldNotFound(file, a.OperationID, fieldName))
		}
	}

	return errs
}

func validateParams(params []parser.ParamBind, opID, file string, api APISymbol) []ValidationError {
	var errs []ValidationError
	for _, p := range params {
		found := false
		for _, ap := range api.Parameters {
			if strings.EqualFold(ap.Name, p.Name) {
				found = true
				break
			}
		}
		if !found {
			errs = append(errs, errParamNotFound(file, opID, p.Name))
		}
	}
	return errs
}

func validateComponent(name, file, frontendDir string) []ValidationError {
	compPath := filepath.Join(frontendDir, "components", name+".tsx")
	if _, err := os.Stat(compPath); os.IsNotExist(err) {
		relPath := filepath.Join("frontend", "components", name+".tsx")
		return []ValidationError{errComponentNotFound(file, name, relPath)}
	}
	return nil
}
