# STML

**SSOT Template Markup Language** — declarative UI-to-API binding with HTML5 `data-*` attributes, symbolic validation against OpenAPI, and React codegen.

## What It Does

```
HTML5 + Tailwind + data-*  →  Parse  →  Validate (OpenAPI)  →  Generate (React TSX)
```

Write **what** the page shows and does in plain HTML. STML generates the **how** — React hooks, forms, mutations, pagination, and all the wiring.

## Example

**Input** (`specs/frontend/my-page.html`):
```html
<main class="max-w-4xl mx-auto p-6">
  <section data-fetch="ListItems" data-paginate data-sort="name:asc">
    <ul data-each="items" class="space-y-2">
      <li class="p-3 border rounded">
        <span data-bind="name" class="font-semibold"></span>
        <span data-bind="status" class="text-sm text-gray-500"></span>
      </li>
    </ul>
    <p data-state="items.empty">항목이 없습니다</p>
  </section>

  <div data-action="CreateItem">
    <input data-field="Name" placeholder="이름" class="border rounded px-3 py-2" />
    <button type="submit">생성</button>
  </div>
</main>
```

**Output** (`artifacts/frontend/my-page.tsx`):
- `useQuery` with pagination, sorting state
- `useMutation` + `useForm` for the create action
- Filter/sort/pagination UI controls
- All Tailwind classes, tags, and text preserved

## Install

```bash
go install github.com/geul-org/stml/artifacts/cmd/stml@latest
```

## Usage

```bash
stml parse <frontend-dir>                  # HTML → JSON
stml validate <project-root>               # OpenAPI cross-check
stml gen <project-root> [output-dir]       # Validate + Generate TSX
```

## data-* Attributes

| Attribute | Purpose |
|---|---|
| `data-fetch` | Connect GET endpoint (useQuery) |
| `data-action` | Connect POST/PUT/DELETE endpoint (useMutation) |
| `data-field` | Bind request body field (input) |
| `data-bind` | Bind response field (display) |
| `data-param-*` | Bind path/query parameter |
| `data-each` | Iterate array field |
| `data-state` | Conditional rendering |
| `data-component` | Delegate to React component |
| `data-paginate` | Enable pagination UI |
| `data-sort` | Default sort column |
| `data-filter` | Filter input columns |
| `data-include` | Include related resources |

## Validation

STML performs 12 symbolic cross-validation checks against OpenAPI:

- operationId existence and HTTP method correctness
- Request/response field existence
- Parameter name matching
- Array type checking for `data-each`
- Component file existence
- Infrastructure params against `x-pagination`, `x-sort`, `x-filter`, `x-include` allowed lists
- Custom.ts function fallback for computed fields

## Project Structure

```
specs/<project>/
  api/openapi.yaml            # OpenAPI 3.x spec
  frontend/*.html             # STML source (SSOT)
  frontend/*.custom.ts        # Frontend calculations (optional)
  frontend/components/*.tsx   # Component wrappers

artifacts/<project>/
  frontend/*.tsx              # Generated output (do not edit)
```

## Tech Stack

| Area | Technology |
|---|---|
| CLI | Go |
| HTML Parsing | golang.org/x/net/html |
| OpenAPI Parsing | gopkg.in/yaml.v3 |
| Output | React + TypeScript + TanStack Query + react-hook-form |

## Documentation

- [`artifacts/manual-for-ai.md`](artifacts/manual-for-ai.md) — AI-oriented reference
- [`artifacts/manual-for-human.md`](artifacts/manual-for-human.md) — User guide (Korean)

## License

[MIT](LICENSE)
