# STML — AI Manual

> Go CLI that parses HTML5 `data-*` attributes, validates them against OpenAPI specs, and generates React TSX components.

## What STML Does

STML takes **HTML5 + Tailwind + `data-*` attributes** as input and produces **React + TypeScript + TanStack Query** components as output. It bridges OpenAPI endpoints to frontend UI declaratively.

```
specs/frontend/*.html  →  stml parse  →  PageSpec (JSON)
                       →  stml validate  →  OpenAPI cross-check
                       →  stml gen  →  artifacts/frontend/*.tsx
```

## data-* Attribute Reference

### Core Attributes (8)

| Attribute | Purpose | Value | Context |
|---|---|---|---|
| `data-fetch` | Bind GET endpoint | `operationId` | Element becomes useQuery block |
| `data-action` | Bind POST/PUT/DELETE endpoint | `operationId` | Element becomes useMutation + form/button |
| `data-field` | Bind request body field | Field name | Inside `data-action`. Generates `<input {...register()}>` |
| `data-bind` | Bind response field | Field name (dot notation OK) | Inside `data-fetch`. Generates `{data.field}` |
| `data-param-*` | Bind path/query parameter | `route.ParamName` or expression | On `data-fetch` or `data-action` element |
| `data-each` | Iterate array field | Array field name | Inside `data-fetch`. Generates `.map()` |
| `data-state` | Conditional render | Condition expression | Suffixes: `.empty`, `.loading`, `.error`, or plain boolean |
| `data-component` | Delegate to React component | Component name | Wrapper file must exist in `components/` |

### Infrastructure Attributes (4)

| Attribute | Purpose | Value | Requires |
|---|---|---|---|
| `data-paginate` | Enable pagination UI | (boolean, no value) | `x-pagination` on OpenAPI operation |
| `data-sort` | Default sort column + direction | `column` or `column:desc` | `x-sort` on OpenAPI operation |
| `data-filter` | Filter input columns | Comma-separated column names | `x-filter` on OpenAPI operation |
| `data-include` | Include related resources | Comma-separated resource names | `x-include` on OpenAPI operation |

## Codegen Mapping

| data-* | React Output |
|---|---|
| `data-fetch="Op"` | `useQuery({ queryKey: ['Op'], queryFn: () => api.Op() })` |
| `data-action="Op"` (with fields) | `useForm()` + `useMutation()` + `<form onSubmit>` |
| `data-action="Op"` (no fields) | `useMutation()` + `<button onClick>` |
| `data-field="Name"` | `<input {...form.register('Name')} />` |
| `data-bind="name"` | `<tag>{data.name}</tag>` |
| `data-param-x="route.X"` | `useParams()` + pass to API call |
| `data-each="items"` | `data.items?.map((item) => ...)` |
| `data-state="x.empty"` | `{data.x?.length === 0 && ...}` |
| `data-state="canX"` | `{data.canX && ...}` |
| `data-component="X"` | `import X from '@/components/X'` + `<X />` |
| `data-paginate` | `useState(page, limit)` + prev/next buttons |
| `data-sort="col:dir"` | `useState(sortBy, sortDir)` + toggle button |
| `data-filter="a,b"` | `useState(filters)` + filter inputs |
| `data-include="rel"` | `include: 'rel'` in API call args |

## Preserved in Codegen

- HTML tags (section, article, ul, li, header, footer, dl, dt, dd)
- Tailwind className attributes
- Placeholder text on inputs
- Button text from HTML
- DOM order of elements
- Static text content

## Validation Rules

| # | Check | Error Condition |
|---|---|---|
| 1 | operationId exists | Not in OpenAPI |
| 2 | HTTP method matches | data-fetch on non-GET, data-action on GET |
| 3 | Parameter exists | data-param-* name not in OpenAPI parameters |
| 4 | Request field exists | data-field name not in request schema |
| 5 | Response field exists | data-bind name not in response schema or custom.ts |
| 6 | Array field type | data-each field is not array type |
| 7 | Component file exists | No .tsx file in components/ |
| 8 | Custom.ts fallback | data-bind not in response AND not in custom.ts |
| 9 | Pagination ext exists | data-paginate but no x-pagination on endpoint |
| 10 | Sort column allowed | data-sort column not in x-sort.allowed |
| 11 | Filter column allowed | data-filter column not in x-filter.allowed |
| 12 | Include resource allowed | data-include resource not in x-include.allowed |

## CLI Usage

```bash
# Parse HTML to JSON PageSpec
stml parse <frontend-dir>

# Validate against OpenAPI
stml validate <project-root>

# Validate + Generate TSX (validate runs first, aborts on failure)
stml gen <project-root> [output-dir]
```

## Project Layout

```
specs/
  <project>/
    api/openapi.yaml          # OpenAPI 3.x with x- extensions
    frontend/
      <page-name>.html        # STML source (input)
      <page-name>.custom.ts   # Frontend calculation functions (optional)
      components/
        <Name>.tsx             # Component wrappers or implementations

cmd/stml/main.go              # CLI entry point
parser/                       # HTML → PageSpec
validator/                    # OpenAPI cross-validation
generator/                    # PageSpec → framework code (Target interface)
  target.go                   #   Target interface + DefaultTarget()
  react_target.go             #   ReactTarget implementation
  react_imports.go            #   React import analysis
  react_templates.go          #   React JSX templates
  generator.go                #   Common types + delegation wrappers

artifacts/
  <project>/
    frontend/
      <page-name>.tsx         # Generated output (do not edit)
```

## Custom.ts Convention

When `data-bind` references a field not in the OpenAPI response, the validator checks for a matching exported function in `<page>.custom.ts`:

```ts
// specs/frontend/cart-page.custom.ts
export function totalPrice(items) {
  return items.reduce((sum, item) => sum + item.price * item.quantity, 0)
}
```

## Component Tiers

| Tier | Coverage | Declaration | Example |
|---|---|---|---|
| data-* | 60-70% | 12 fixed attributes | Lists, forms, CRUD, pagination |
| React wrapper | 20-30% | `data-component` + re-export | DatePicker, RichEditor, Chart |
| Custom + JSDoc | 5-10% | `data-component` + implementation | KanbanBoard, CodeReviewTimeline |
