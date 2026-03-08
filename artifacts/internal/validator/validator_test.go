package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/geul-org/stml/artifacts/internal/parser"
)

// setupTestProject creates a temporary project with openapi.yaml and optional files.
func setupTestProject(t *testing.T, openapi string, customTS map[string]string, components []string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")

	os.MkdirAll(filepath.Join(root, "api"), 0o755)
	os.MkdirAll(filepath.Join(root, "frontend", "components"), 0o755)

	os.WriteFile(filepath.Join(root, "api", "openapi.yaml"), []byte(openapi), 0o644)

	for name, content := range customTS {
		os.WriteFile(filepath.Join(root, "frontend", name), []byte(content), 0o644)
	}

	for _, comp := range components {
		os.WriteFile(filepath.Join(root, "frontend", "components", comp+".tsx"), []byte("export default {}"), 0o644)
	}

	return root
}

const dummyOpenAPI = `
openapi: "3.0.3"
info:
  title: Test API
  version: "1.0.0"
paths:
  /reservations:
    post:
      operationId: CreateReservation
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                RoomID:
                  type: integer
                StartAt:
                  type: string
                EndAt:
                  type: string
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  reservation:
                    type: object
  /me/reservations:
    get:
      operationId: ListMyReservations
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  reservations:
                    type: array
                    items:
                      type: object
  /reservations/{ReservationID}:
    get:
      operationId: GetReservation
      parameters:
        - name: ReservationID
          in: path
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  reservation:
                    type: object
  /login:
    post:
      operationId: Login
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                Email:
                  type: string
                Password:
                  type: string
      responses:
        "200":
          description: ok
`

func TestValidateDummyStudyPass(t *testing.T) {
	root := setupTestProject(t, dummyOpenAPI, nil, []string{"DatePicker"})

	pages := []parser.PageSpec{
		{
			Name: "login-page", FileName: "login-page.html",
			Actions: []parser.ActionBlock{
				{OperationID: "Login", Fields: []parser.FieldBind{
					{Name: "Email", Tag: "input", Type: "email"},
					{Name: "Password", Tag: "input", Type: "password"},
				}},
			},
		},
		{
			Name: "my-reservations-page", FileName: "my-reservations-page.html",
			Fetches: []parser.FetchBlock{
				{
					OperationID: "ListMyReservations",
					Eaches: []parser.EachBlock{
						{Field: "reservations", Binds: []parser.FieldBind{{Name: "RoomID", Tag: "span"}}},
					},
					States: []parser.StateBind{{Condition: "reservations.empty"}},
				},
			},
			Actions: []parser.ActionBlock{
				{OperationID: "CreateReservation", Fields: []parser.FieldBind{
					{Name: "RoomID", Tag: "input", Type: "number"},
					{Name: "StartAt", Tag: "data-component:DatePicker"},
					{Name: "EndAt", Tag: "data-component:DatePicker"},
				}},
			},
		},
		{
			Name: "reservation-detail-page", FileName: "reservation-detail-page.html",
			Fetches: []parser.FetchBlock{
				{
					OperationID: "GetReservation",
					Params:      []parser.ParamBind{{Name: "reservationId", Source: "route.ReservationID"}},
					Binds:       []parser.FieldBind{{Name: "reservation.Status", Tag: "span"}},
				},
			},
		},
	}

	errs := Validate(pages, root)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Error(e.Error())
		}
	}
}

func TestValidateOperationNotFound(t *testing.T) {
	root := setupTestProject(t, dummyOpenAPI, nil, nil)

	pages := []parser.PageSpec{
		{
			Name: "test-page", FileName: "test-page.html",
			Fetches: []parser.FetchBlock{
				{OperationID: "NonExistent"},
			},
		},
	}

	errs := Validate(pages, root)
	assertHasError(t, errs, "NonExistent")
	assertHasError(t, errs, "operationId가 없습니다")
}

func TestValidateWrongMethod(t *testing.T) {
	root := setupTestProject(t, dummyOpenAPI, nil, nil)

	pages := []parser.PageSpec{
		{
			Name: "test-page", FileName: "test-page.html",
			Fetches: []parser.FetchBlock{
				{OperationID: "Login"}, // Login is POST, not GET
			},
		},
	}

	errs := Validate(pages, root)
	assertHasError(t, errs, "POST")
	assertHasError(t, errs, "GET이어야 함")
}

func TestValidateParamNotFound(t *testing.T) {
	root := setupTestProject(t, dummyOpenAPI, nil, nil)

	pages := []parser.PageSpec{
		{
			Name: "test-page", FileName: "test-page.html",
			Fetches: []parser.FetchBlock{
				{
					OperationID: "GetReservation",
					Params:      []parser.ParamBind{{Name: "nonExistentParam", Source: "route.foo"}},
				},
			},
		},
	}

	errs := Validate(pages, root)
	assertHasError(t, errs, "nonExistentParam")
}

func TestValidateFieldNotFound(t *testing.T) {
	root := setupTestProject(t, dummyOpenAPI, nil, nil)

	pages := []parser.PageSpec{
		{
			Name: "test-page", FileName: "test-page.html",
			Actions: []parser.ActionBlock{
				{OperationID: "Login", Fields: []parser.FieldBind{
					{Name: "NonExistentField", Tag: "input"},
				}},
			},
		},
	}

	errs := Validate(pages, root)
	assertHasError(t, errs, "NonExistentField")
	assertHasError(t, errs, "request schema")
}

func TestValidateBindNotFound(t *testing.T) {
	root := setupTestProject(t, dummyOpenAPI, nil, nil)

	pages := []parser.PageSpec{
		{
			Name: "test-page", FileName: "test-page.html",
			Fetches: []parser.FetchBlock{
				{
					OperationID: "GetReservation",
					Binds:       []parser.FieldBind{{Name: "nonExistent", Tag: "span"}},
				},
			},
		},
	}

	errs := Validate(pages, root)
	assertHasError(t, errs, "nonExistent")
	assertHasError(t, errs, "custom.ts에도")
}

func TestValidateEachNotArray(t *testing.T) {
	root := setupTestProject(t, dummyOpenAPI, nil, nil)

	pages := []parser.PageSpec{
		{
			Name: "test-page", FileName: "test-page.html",
			Fetches: []parser.FetchBlock{
				{
					OperationID: "GetReservation",
					Eaches:      []parser.EachBlock{{Field: "reservation"}}, // reservation is object, not array
				},
			},
		},
	}

	errs := Validate(pages, root)
	assertHasError(t, errs, "배열이 아닙니다")
}

func TestValidateComponentNotFound(t *testing.T) {
	root := setupTestProject(t, dummyOpenAPI, nil, nil) // no components created

	pages := []parser.PageSpec{
		{
			Name: "test-page", FileName: "test-page.html",
			Fetches: []parser.FetchBlock{
				{
					OperationID: "ListMyReservations",
					Components:  []parser.ComponentRef{{Name: "MissingComponent"}},
				},
			},
		},
	}

	errs := Validate(pages, root)
	assertHasError(t, errs, "MissingComponent")
	assertHasError(t, errs, "파일이 없습니다")
}

func TestValidateCustomTSFallback(t *testing.T) {
	customFiles := map[string]string{
		"test-page.custom.ts": `export function totalPrice(items) {
  return items.reduce((sum, item) => sum + item.price, 0)
}`,
	}
	root := setupTestProject(t, dummyOpenAPI, customFiles, nil)

	pages := []parser.PageSpec{
		{
			Name: "test-page", FileName: "test-page.html",
			Fetches: []parser.FetchBlock{
				{
					OperationID: "GetReservation",
					Binds:       []parser.FieldBind{{Name: "totalPrice", Tag: "span"}},
				},
			},
		},
	}

	errs := Validate(pages, root)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Error(e.Error())
		}
		t.Fatal("expected no errors with custom.ts fallback")
	}
}

// --- Phase 5: infra param validation tests ---

const infraOpenAPI = `
openapi: "3.0.3"
info:
  title: Test API
  version: "1.0.0"
paths:
  /items:
    get:
      operationId: ListItems
      x-pagination:
        style: offset
        defaultLimit: 20
        maxLimit: 100
      x-sort:
        allowed: [name, created_at]
        default: name
        direction: asc
      x-filter:
        allowed: [status, category]
      x-include:
        allowed: [author, tags]
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array
                    items:
                      type: object
                  total:
                    type: integer
  /simple:
    get:
      operationId: ListSimple
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array
                    items:
                      type: object
`

func TestValidateInfraParamsPass(t *testing.T) {
	root := setupTestProject(t, infraOpenAPI, nil, nil)

	pages := []parser.PageSpec{{
		Name: "test-page", FileName: "test-page.html",
		Fetches: []parser.FetchBlock{{
			OperationID: "ListItems",
			Paginate:    true,
			Sort:        &parser.SortDecl{Column: "name", Direction: "desc"},
			Filters:     []string{"status"},
			Includes:    []string{"author"},
			Eaches:      []parser.EachBlock{{Field: "items"}},
		}},
	}}

	errs := Validate(pages, root)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Error(e.Error())
		}
	}
}

func TestValidatePaginateNoExt(t *testing.T) {
	root := setupTestProject(t, infraOpenAPI, nil, nil)

	pages := []parser.PageSpec{{
		Name: "test-page", FileName: "test-page.html",
		Fetches: []parser.FetchBlock{{
			OperationID: "ListSimple",
			Paginate:    true,
			Eaches:      []parser.EachBlock{{Field: "items"}},
		}},
	}}

	errs := Validate(pages, root)
	assertHasError(t, errs, "x-pagination이 선언되지 않았습니다")
}

func TestValidateSortNotAllowed(t *testing.T) {
	root := setupTestProject(t, infraOpenAPI, nil, nil)

	pages := []parser.PageSpec{{
		Name: "test-page", FileName: "test-page.html",
		Fetches: []parser.FetchBlock{{
			OperationID: "ListItems",
			Sort:        &parser.SortDecl{Column: "invalid_col", Direction: "asc"},
			Eaches:      []parser.EachBlock{{Field: "items"}},
		}},
	}}

	errs := Validate(pages, root)
	assertHasError(t, errs, "x-sort.allowed")
	assertHasError(t, errs, "invalid_col")
}

func TestValidateFilterNotAllowed(t *testing.T) {
	root := setupTestProject(t, infraOpenAPI, nil, nil)

	pages := []parser.PageSpec{{
		Name: "test-page", FileName: "test-page.html",
		Fetches: []parser.FetchBlock{{
			OperationID: "ListItems",
			Filters:     []string{"status", "bad_col"},
			Eaches:      []parser.EachBlock{{Field: "items"}},
		}},
	}}

	errs := Validate(pages, root)
	assertHasError(t, errs, "x-filter.allowed")
	assertHasError(t, errs, "bad_col")
}

func TestValidateIncludeNotAllowed(t *testing.T) {
	root := setupTestProject(t, infraOpenAPI, nil, nil)

	pages := []parser.PageSpec{{
		Name: "test-page", FileName: "test-page.html",
		Fetches: []parser.FetchBlock{{
			OperationID: "ListItems",
			Includes:    []string{"author", "bad_rel"},
			Eaches:      []parser.EachBlock{{Field: "items"}},
		}},
	}}

	errs := Validate(pages, root)
	assertHasError(t, errs, "x-include.allowed")
	assertHasError(t, errs, "bad_rel")
}

func TestValidateSortNoExtOnEndpoint(t *testing.T) {
	root := setupTestProject(t, infraOpenAPI, nil, nil)

	pages := []parser.PageSpec{{
		Name: "test-page", FileName: "test-page.html",
		Fetches: []parser.FetchBlock{{
			OperationID: "ListSimple",
			Sort:        &parser.SortDecl{Column: "name", Direction: "asc"},
			Eaches:      []parser.EachBlock{{Field: "items"}},
		}},
	}}

	errs := Validate(pages, root)
	assertHasError(t, errs, "x-sort.allowed")
}

func assertHasError(t *testing.T, errs []ValidationError, substr string) {
	t.Helper()
	for _, e := range errs {
		if strings.Contains(e.Error(), substr) {
			return
		}
	}
	t.Errorf("expected error containing %q, got: %v", substr, errs)
}
