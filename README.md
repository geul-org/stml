# STML

**SSOT Template Markup Language** — a framework-independent declaration of what your frontend pages show and do.

## Why

Frontend code mixes two things: **your decisions** and **framework wiring**.

Your decisions — which API to call, which fields to show, in what order, with what layout and style, under what conditions — are buried inside React hooks, Vue composables, or Svelte load functions. When you switch frameworks, refactor, or ask AI to rewrite, those decisions get re-interpreted, guessed at, or lost.

STML separates them. Your decisions go into plain HTML with `data-*` attributes. Framework wiring is generated or delegated.

```
┌─────────────────────────────────┐
│  STML (specs/)                  │  Your decisions. Framework-independent.
│  what to fetch, show, submit    │  Survives any rewrite.
├─────────────────────────────────┤
│  Codegen / LLM                  │  Generates framework wiring.
│  React, Vue, Svelte, anything   │  Replaceable.
├─────────────────────────────────┤
│  Runtime (artifacts/)           │  React TSX, Vue SFC, etc.
│  hooks, state, rendering        │  Generated. Do not edit.
└─────────────────────────────────┘
```

## What Gets Preserved

Everything in STML HTML is a user decision:

- **Which API endpoints** to call (`data-fetch`, `data-action`)
- **Which fields** to display or collect (`data-bind`, `data-field`)
- **What order** elements appear (DOM structure)
- **What it looks like** (Tailwind classes, HTML tags)
- **What text** users see (headings, placeholders, button labels)
- **What conditions** control visibility (`data-state`)
- **What components** handle special UX (`data-component`)
- **How lists behave** — pagination, sorting, filtering

None of this is React. None of this is Vue. It's what your page does, declared in standard HTML5.

## Example

```html
<main class="max-w-4xl mx-auto p-6">
  <h1 class="text-2xl font-bold mb-6">My Reservations</h1>

  <section data-fetch="ListMyReservations" data-paginate data-sort="StartAt:desc" data-filter="Status">
    <ul data-each="reservations" class="space-y-3">
      <li class="flex justify-between p-4 border rounded">
        <span data-bind="RoomID" class="font-semibold"></span>
        <span data-bind="Status" class="text-sm text-gray-500"></span>
      </li>
    </ul>
    <p data-state="reservations.empty" class="text-gray-400">No reservations</p>
  </section>

  <div data-action="CreateReservation">
    <input data-field="RoomID" type="number" placeholder="Room number" />
    <button type="submit">Reserve</button>
  </div>
</main>
```

This same file can produce:

| Target | How |
|---|---|
| React TSX | `stml gen` (built-in deterministic codegen) |
| Vue SFC | Give STML + OpenAPI to LLM: "implement in Vue" |
| Svelte | Give STML + OpenAPI to LLM: "implement in Svelte" |
| Flutter | Give STML + OpenAPI to LLM: "implement in Flutter" |

Your decisions are preserved. Only the framework wiring changes.

## Validation

The built-in validator cross-checks STML against OpenAPI — before any code is generated or written:

- Does the operationId exist? Is the HTTP method correct?
- Do request fields, response fields, and parameters match the schema?
- Are sort/filter/include columns within the allowed lists?
- Do referenced components exist?

This catches frontend-API mismatches at CI time, not at runtime.

```bash
stml validate specs/my-project    # 11 symbolic checks against OpenAPI
```

## Built-in Codegen (React PoC)

The Go CLI includes deterministic React codegen as a proof of concept:

```bash
stml gen specs/my-project artifacts/my-project/frontend
```

Generates React TSX with useQuery, useMutation, useForm, useState, pagination controls, sort toggles, and filter inputs. All Tailwind classes, tags, and text preserved from the source HTML.

This is one possible output. STML is the source of truth, not the generated code.

## data-* Attributes

| Attribute | What it declares |
|---|---|
| `data-fetch` | This element loads data from a GET endpoint |
| `data-action` | This element submits data to a POST/PUT/DELETE endpoint |
| `data-field` | This input collects a request body field |
| `data-bind` | This element displays a response field |
| `data-param-*` | This operation needs a path/query parameter |
| `data-each` | This container repeats over an array field |
| `data-state` | This element shows conditionally |
| `data-component` | This element delegates to a custom component |
| `data-paginate` | This list is paginated |
| `data-sort` | This list is sortable (default column and direction) |
| `data-filter` | This list is filterable (which columns) |

## Project Structure

```
specs/<project>/
  api/openapi.yaml              # API contract
  frontend/*.html               # STML declarations (your decisions)
  frontend/*.custom.ts          # Frontend calculations (optional)
  frontend/components/*.tsx     # Custom component specs

artifacts/<project>/
  frontend/*.tsx                # Generated output (do not edit)
```

## Install

```bash
go install github.com/geul-org/stml/cmd/stml@latest
```

## Documentation

- [`artifacts/manual-for-ai.md`](artifacts/manual-for-ai.md) — AI-consumable reference
- [`artifacts/manual-for-human.md`](artifacts/manual-for-human.md) — User guide (Korean)

## License

[MIT](LICENSE)
