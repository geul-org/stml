package validator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SymbolTable holds all symbols extracted from OpenAPI and custom.ts.
type SymbolTable struct {
	Operations map[string]APISymbol // operationId → APISymbol
}

// APISymbol represents a single OpenAPI operation.
type APISymbol struct {
	Method         string                 // "get", "post", "put", "delete"
	Parameters     []ParamSymbol          // path/query parameters
	RequestFields  map[string]string      // field name → type
	ResponseFields map[string]FieldSymbol // field name → type info

	// Phase 5: x- extensions
	Pagination *PaginationExt
	Sort       *SortExt
	Filter     *FilterExt
}

// PaginationExt represents x-pagination extension.
type PaginationExt struct {
	Style        string // "offset" or "cursor"
	DefaultLimit int
	MaxLimit     int
}

// SortExt represents x-sort extension.
type SortExt struct {
	Allowed   []string
	Default   string
	Direction string
}

// FilterExt represents x-filter extension.
type FilterExt struct {
	Allowed []string
}

// ParamSymbol represents an OpenAPI parameter.
type ParamSymbol struct {
	Name string // parameter name
	In   string // "path" or "query"
}

// FieldSymbol represents a response field with type info.
type FieldSymbol struct {
	Type     string // "string", "integer", "array", "object"
	ItemType string // item type if array
}

// CustomSymbol holds exported function names from a custom.ts file.
type CustomSymbol struct {
	Functions map[string]bool
}

// LoadOpenAPI parses an OpenAPI YAML file and builds a SymbolTable.
func LoadOpenAPI(path string) (*SymbolTable, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read openapi: %w", err)
	}

	var doc openAPIDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi: %w", err)
	}

	st := &SymbolTable{Operations: make(map[string]APISymbol)}

	for _, pathItem := range doc.Paths {
		for method, op := range pathItem {
			if op.OperationID == "" {
				continue
			}
			api := APISymbol{
				Method:         strings.ToUpper(method),
				RequestFields:  make(map[string]string),
				ResponseFields: make(map[string]FieldSymbol),
			}

			// Parameters
			for _, p := range op.Parameters {
				api.Parameters = append(api.Parameters, ParamSymbol{
					Name: p.Name,
					In:   p.In,
				})
			}

			// Request body fields
			if op.RequestBody.Content != nil {
				for _, ct := range op.RequestBody.Content {
					ref := ct.Schema.Ref
					if ref != "" {
						resolveSchemaFields(doc.Components.Schemas, ref, api.RequestFields)
					} else {
						for fname, fprop := range ct.Schema.Properties {
							api.RequestFields[fname] = fprop.Type
						}
					}
				}
			}

			// Response fields (from 200 response)
			if resp, ok := op.Responses["200"]; ok {
				for _, ct := range resp.Content {
					ref := ct.Schema.Ref
					if ref != "" {
						resolveResponseFields(doc.Components.Schemas, ref, api.ResponseFields)
					} else {
						for fname, fprop := range ct.Schema.Properties {
							api.ResponseFields[fname] = toFieldSymbol(fprop, doc.Components.Schemas)
						}
					}
				}
			}

			// Phase 5: x- extensions
			if op.XPagination != nil {
				api.Pagination = &PaginationExt{
					Style:        op.XPagination.Style,
					DefaultLimit: op.XPagination.DefaultLimit,
					MaxLimit:     op.XPagination.MaxLimit,
				}
			}
			if op.XSort != nil {
				dir := op.XSort.Direction
				if dir == "" {
					dir = "asc"
				}
				api.Sort = &SortExt{
					Allowed:   op.XSort.Allowed,
					Default:   op.XSort.Default,
					Direction: dir,
				}
			}
			if op.XFilter != nil {
				api.Filter = &FilterExt{Allowed: op.XFilter.Allowed}
			}
			st.Operations[op.OperationID] = api
		}
	}

	return st, nil
}

// LoadCustomTS parses a custom.ts file and extracts exported function names.
func LoadCustomTS(path string) (*CustomSymbol, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CustomSymbol{Functions: make(map[string]bool)}, nil
		}
		return nil, err
	}
	defer f.Close()

	cs := &CustomSymbol{Functions: make(map[string]bool)}
	re := regexp.MustCompile(`export\s+function\s+(\w+)`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := re.FindStringSubmatch(scanner.Text()); m != nil {
			cs.Functions[m[1]] = true
		}
	}
	return cs, scanner.Err()
}

// --- OpenAPI YAML types ---

type openAPIDoc struct {
	Paths      map[string]map[string]openAPIOperation `yaml:"paths"`
	Components struct {
		Schemas map[string]openAPISchema `yaml:"schemas"`
	} `yaml:"components"`
}

type openAPIOperation struct {
	OperationID  string                       `yaml:"operationId"`
	Parameters   []openAPIParam               `yaml:"parameters"`
	RequestBody  openAPIRequestBody           `yaml:"requestBody"`
	Responses    map[string]openAPIResponse   `yaml:"responses"`
	XPagination  *yamlPaginationExt           `yaml:"x-pagination"`
	XSort        *yamlSortExt                 `yaml:"x-sort"`
	XFilter      *yamlFilterExt               `yaml:"x-filter"`
}

type yamlPaginationExt struct {
	Style        string `yaml:"style"`
	DefaultLimit int    `yaml:"defaultLimit"`
	MaxLimit     int    `yaml:"maxLimit"`
}

type yamlSortExt struct {
	Allowed   []string `yaml:"allowed"`
	Default   string   `yaml:"default"`
	Direction string   `yaml:"direction"`
}

type yamlFilterExt struct {
	Allowed []string `yaml:"allowed"`
}

type openAPIParam struct {
	Name   string        `yaml:"name"`
	In     string        `yaml:"in"`
	Schema openAPISchema `yaml:"schema"`
}

type openAPIRequestBody struct {
	Content map[string]openAPIMediaType `yaml:"content"`
}

type openAPIResponse struct {
	Content map[string]openAPIMediaType `yaml:"content"`
}

type openAPIMediaType struct {
	Schema openAPISchema `yaml:"schema"`
}

type openAPISchema struct {
	Ref        string                     `yaml:"$ref"`
	Type       string                     `yaml:"type"`
	Properties map[string]openAPISchema   `yaml:"properties"`
	Items      *openAPISchema             `yaml:"items"`
}

// resolveSchemaFields resolves a $ref and collects field names into the map.
func resolveSchemaFields(schemas map[string]openAPISchema, ref string, fields map[string]string) {
	name := refName(ref)
	schema, ok := schemas[name]
	if !ok {
		return
	}
	for fname, fprop := range schema.Properties {
		typ := fprop.Type
		if fprop.Ref != "" {
			typ = "object"
		}
		fields[fname] = typ
	}
}

// resolveResponseFields resolves a $ref and collects response field symbols.
func resolveResponseFields(schemas map[string]openAPISchema, ref string, fields map[string]FieldSymbol) {
	name := refName(ref)
	schema, ok := schemas[name]
	if !ok {
		return
	}
	for fname, fprop := range schema.Properties {
		fields[fname] = toFieldSymbol(fprop, schemas)
	}
}

func toFieldSymbol(s openAPISchema, schemas map[string]openAPISchema) FieldSymbol {
	if s.Ref != "" {
		return FieldSymbol{Type: "object", ItemType: refName(s.Ref)}
	}
	if s.Type == "array" && s.Items != nil {
		itemType := s.Items.Type
		if s.Items.Ref != "" {
			itemType = refName(s.Items.Ref)
		}
		return FieldSymbol{Type: "array", ItemType: itemType}
	}
	return FieldSymbol{Type: s.Type}
}

func refName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}
