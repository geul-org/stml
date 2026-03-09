package generator

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/geul-org/stml/parser"
)

// renderUseQuery generates a useQuery hook call.
func renderUseQuery(f parser.FetchBlock) string {
	alias := toLowerFirst(f.OperationID) + "Data"
	paramValues := renderParamValues(f.Params)
	paramArgs := renderParamArgs(f.Params)

	// queryKey parts
	queryKey := fmt.Sprintf("'%s'", f.OperationID)
	if paramValues != "" {
		queryKey += ", " + paramValues
	}
	// Phase 5: infra params in queryKey
	if f.Paginate {
		queryKey += ", page, limit"
	}
	if f.Sort != nil {
		queryKey += ", sortBy, sortDir"
	}
	if len(f.Filters) > 0 {
		queryKey += ", filters"
	}

	// API call args
	hasInfra := f.Paginate || f.Sort != nil || len(f.Filters) > 0
	apiArgs := paramArgs
	if hasInfra {
		apiArgs = renderInfraApiArgs(f, paramArgs)
	}
	apiCall := fmt.Sprintf("api.%s(%s)", f.OperationID, apiArgs)

	return fmt.Sprintf(`const { data: %s, isLoading: %sLoading, error: %sError } = useQuery({
    queryKey: [%s],
    queryFn: () => %s,
  })`, alias, alias, alias, queryKey, apiCall)
}

// renderInfraApiArgs builds the API call arguments with infra params.
func renderInfraApiArgs(f parser.FetchBlock, paramArgs string) string {
	var parts []string

	// Spread existing params
	if paramArgs != "" {
		// paramArgs is "{ key: val, ... }" — strip braces and spread
		inner := strings.TrimPrefix(paramArgs, "{ ")
		inner = strings.TrimSuffix(inner, " }")
		parts = append(parts, inner)
	}

	if f.Paginate {
		parts = append(parts, "page", "limit")
	}
	if f.Sort != nil {
		parts = append(parts, "sortBy", "sortDir")
	}
	if len(f.Filters) > 0 {
		parts = append(parts, "...filters")
	}

	return "{ " + strings.Join(parts, ", ") + " }"
}

// renderUseMutation generates a useMutation hook call.
func renderUseMutation(a parser.ActionBlock, fetchOps []string) string {
	mutName := toLowerFirst(a.OperationID) + "Mutation"
	paramArgs := renderParamArgs(a.Params)

	apiArgs := "data"
	if paramArgs != "" {
		inner := strings.TrimPrefix(paramArgs, "{ ")
		inner = strings.TrimSuffix(inner, " }")
		apiArgs = "{ ...data, " + inner + " }"
	}

	// onSuccess: invalidate related queries
	invalidate := "queryClient.invalidateQueries()"
	if len(fetchOps) > 0 {
		var parts []string
		for _, op := range fetchOps {
			parts = append(parts, fmt.Sprintf("queryClient.invalidateQueries({ queryKey: ['%s'] })", op))
		}
		invalidate = strings.Join(parts, "\n      ")
	}

	return fmt.Sprintf(`const %s = useMutation({
    mutationFn: (data: any) => api.%s(%s),
    onSuccess: () => {
      %s
    },
  })`, mutName, a.OperationID, apiArgs, invalidate)
}

// renderFormHook generates a useForm hook call.
func renderFormHook(a parser.ActionBlock) string {
	formName := toLowerFirst(a.OperationID) + "Form"
	return fmt.Sprintf(`const %s = useForm()`, formName)
}

// --- JSX rendering with ChildNode tree ---

// renderFetchJSX generates JSX for a FetchBlock using ChildNode tree.
func renderFetchJSX(f parser.FetchBlock, indent int) string {
	alias := toLowerFirst(f.OperationID) + "Data"
	ind := indentStr(indent)
	tag := orDefault(f.Tag, "div")
	cls := clsAttr(f.ClassName)

	var lines []string
	lines = append(lines, fmt.Sprintf("%s{%sLoading && <div>로딩 중...</div>}", ind, alias))
	lines = append(lines, fmt.Sprintf("%s{%sError && <div>오류가 발생했습니다</div>}", ind, alias))
	lines = append(lines, fmt.Sprintf("%s{%s && (", ind, alias))
	lines = append(lines, fmt.Sprintf("%s  <%s%s>", ind, tag, cls))

	// Phase 5: filter UI
	if len(f.Filters) > 0 {
		lines = append(lines, renderFilterUI(f.Filters, indent+4)...)
	}

	// Phase 5: sort UI
	if f.Sort != nil {
		lines = append(lines, renderSortUI(f.Sort, indent+4)...)
	}

	if len(f.Children) > 0 {
		lines = append(lines, renderChildNodes(f.Children, alias, "item", indent+4)...)
	} else {
		// Fallback: use flat slices (backward compat)
		for _, b := range f.Binds {
			lines = append(lines, renderBindJSX(b, alias, indent+4))
		}
		for _, e := range f.Eaches {
			lines = append(lines, renderEachJSX(e, alias, indent+4))
		}
		for _, s := range f.States {
			lines = append(lines, renderStateJSX(s, alias, indent+4))
		}
		for _, c := range f.Components {
			lines = append(lines, renderComponentJSX(c, alias, indent+4))
		}
	}

	// Phase 5: pagination UI
	if f.Paginate {
		lines = append(lines, renderPaginationUI(alias, indent+4)...)
	}

	lines = append(lines, fmt.Sprintf("%s  </%s>", ind, tag))
	lines = append(lines, fmt.Sprintf("%s)}", ind))

	return strings.Join(lines, "\n")
}

// renderFilterUI generates filter input controls.
func renderFilterUI(filters []string, indent int) []string {
	ind := indentStr(indent)
	var lines []string
	lines = append(lines, fmt.Sprintf(`%s<div className="flex gap-2 mb-4">`, ind))
	for _, col := range filters {
		lines = append(lines, fmt.Sprintf(`%s  <input placeholder="%s" value={filters.%s ?? ''} className="px-3 py-2 border rounded" onChange={(e) => setFilters(f => ({ ...f, %s: e.target.value }))} />`, ind, col, col, col))
	}
	lines = append(lines, fmt.Sprintf(`%s</div>`, ind))
	return lines
}

// renderSortUI generates sort toggle controls.
func renderSortUI(sort *parser.SortDecl, indent int) []string {
	ind := indentStr(indent)
	var lines []string
	lines = append(lines, fmt.Sprintf(`%s<div className="flex gap-2 mb-4">`, ind))
	lines = append(lines, fmt.Sprintf(`%s  <button onClick={() => { setSortBy('%s'); setSortDir(d => d === 'asc' ? 'desc' : 'asc') }}>`, ind, sort.Column))
	lines = append(lines, fmt.Sprintf(`%s    %s {sortBy === '%s' ? (sortDir === 'asc' ? '↑' : '↓') : ''}`, ind, sort.Column, sort.Column))
	lines = append(lines, fmt.Sprintf(`%s  </button>`, ind))
	lines = append(lines, fmt.Sprintf(`%s</div>`, ind))
	return lines
}

// renderPaginationUI generates pagination controls.
func renderPaginationUI(alias string, indent int) []string {
	ind := indentStr(indent)
	var lines []string
	lines = append(lines, fmt.Sprintf(`%s<div className="flex justify-between items-center mt-4">`, ind))
	lines = append(lines, fmt.Sprintf(`%s  <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}>이전</button>`, ind))
	lines = append(lines, fmt.Sprintf(`%s  <span>{page} / {Math.ceil((%s?.total ?? 0) / limit)}</span>`, ind, alias))
	lines = append(lines, fmt.Sprintf(`%s  <button disabled={!%s?.total || page * limit >= %s.total} onClick={() => setPage(p => p + 1)}>다음</button>`, ind, alias, alias))
	lines = append(lines, fmt.Sprintf(`%s</div>`, ind))
	return lines
}

// renderActionJSX generates JSX for an ActionBlock.
func renderActionJSX(a parser.ActionBlock, indent int) string {
	ind := indentStr(indent)
	mutName := toLowerFirst(a.OperationID) + "Mutation"

	// No fields → button onClick (no form needed)
	if len(a.Fields) == 0 {
		tag := orDefault(a.Tag, "button")
		cls := clsAttr(a.ClassName)
		text := orDefault(a.SubmitText, a.OperationID)
		if tag == "button" {
			return fmt.Sprintf(`%s<button onClick={() => %s.mutate({})}%s>%s</button>`, ind, mutName, cls, text)
		}
		return fmt.Sprintf(`%s<%s%s><button onClick={() => %s.mutate({})}>%s</button></%s>`, ind, tag, cls, mutName, text, tag)
	}

	formName := toLowerFirst(a.OperationID) + "Form"
	cls := clsAttr(a.ClassName)
	submitText := orDefault(a.SubmitText, "제출")

	var lines []string
	lines = append(lines, fmt.Sprintf(`%s<form onSubmit={%s.handleSubmit((data) => %s.mutate(data))}%s>`, ind, formName, mutName, cls))

	if len(a.Children) > 0 {
		lines = append(lines, renderActionChildNodes(a.Children, formName, indent+2)...)
	} else {
		for _, f := range a.Fields {
			lines = append(lines, renderFieldJSX(f, formName, indent+2))
		}
	}

	lines = append(lines, fmt.Sprintf(`%s  <button type="submit">%s</button>`, ind, submitText))
	lines = append(lines, fmt.Sprintf(`%s</form>`, ind))

	return strings.Join(lines, "\n")
}

// renderChildNodes renders ChildNode slice in DOM order for fetch context.
func renderChildNodes(nodes []parser.ChildNode, dataVar, itemVar string, indent int) []string {
	var lines []string
	for _, ch := range nodes {
		switch ch.Kind {
		case "bind":
			lines = append(lines, renderBindJSX(*ch.Bind, dataVar, indent))
		case "each":
			lines = append(lines, renderEachJSX(*ch.Each, dataVar, indent))
		case "state":
			lines = append(lines, renderStateJSX(*ch.State, dataVar, indent))
		case "component":
			lines = append(lines, renderComponentJSX(*ch.Component, dataVar, indent))
		case "static":
			lines = append(lines, renderStaticJSX(*ch.Static, dataVar, itemVar, indent))
		case "action":
			lines = append(lines, renderActionJSX(*ch.Action, indent))
		case "fetch":
			lines = append(lines, renderFetchJSX(*ch.Fetch, indent))
		}
	}
	return lines
}

// renderActionChildNodes renders ChildNode slice in DOM order for action context.
func renderActionChildNodes(nodes []parser.ChildNode, formName string, indent int) []string {
	var lines []string
	for _, ch := range nodes {
		switch ch.Kind {
		case "bind":
			lines = append(lines, renderFieldJSX(*ch.Bind, formName, indent))
		case "static":
			lines = append(lines, renderStaticActionJSX(*ch.Static, formName, indent))
		}
	}
	return lines
}

// renderStaticJSX renders a StaticElement preserving structure.
func renderStaticJSX(se parser.StaticElement, dataVar, itemVar string, indent int) string {
	ind := indentStr(indent)
	tag := se.Tag
	cls := clsAttr(se.ClassName)

	if len(se.Children) == 0 {
		if se.Text != "" {
			return fmt.Sprintf("%s<%s%s>%s</%s>", ind, tag, cls, se.Text, tag)
		}
		return fmt.Sprintf("%s<%s%s />", ind, tag, cls)
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%s<%s%s>", ind, tag, cls))
	if se.Text != "" {
		lines = append(lines, fmt.Sprintf("%s  %s", ind, se.Text))
	}
	lines = append(lines, renderChildNodes(se.Children, dataVar, itemVar, indent+2)...)
	lines = append(lines, fmt.Sprintf("%s</%s>", ind, tag))
	return strings.Join(lines, "\n")
}

// renderStaticActionJSX renders a StaticElement inside an action form.
func renderStaticActionJSX(se parser.StaticElement, formName string, indent int) string {
	ind := indentStr(indent)
	tag := se.Tag
	cls := clsAttr(se.ClassName)

	if len(se.Children) == 0 {
		if se.Text != "" {
			return fmt.Sprintf("%s<%s%s>%s</%s>", ind, tag, cls, se.Text, tag)
		}
		return fmt.Sprintf("%s<%s%s />", ind, tag, cls)
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%s<%s%s>", ind, tag, cls))
	lines = append(lines, renderActionChildNodes(se.Children, formName, indent+2)...)
	lines = append(lines, fmt.Sprintf("%s</%s>", ind, tag))
	return strings.Join(lines, "\n")
}

// renderBindJSX generates JSX for a data-bind field.
func renderBindJSX(b parser.FieldBind, dataVar string, indent int) string {
	ind := indentStr(indent)
	tag := orDefault(b.Tag, "span")
	cls := clsAttr(b.ClassName)
	return fmt.Sprintf("%s<%s%s>{%s.%s}</%s>", ind, tag, cls, dataVar, b.Name, tag)
}

// renderFieldJSX generates JSX for a form field.
func renderFieldJSX(f parser.FieldBind, formName string, indent int) string {
	ind := indentStr(indent)

	// data-component field
	if strings.HasPrefix(f.Tag, "data-component:") {
		comp := strings.TrimPrefix(f.Tag, "data-component:")
		return fmt.Sprintf("%s<%s {...%s.register('%s')} />", ind, comp, formName, f.Name)
	}

	var attrs []string
	if f.Type != "" {
		attrs = append(attrs, fmt.Sprintf(`type="%s"`, f.Type))
	}
	if f.Placeholder != "" {
		attrs = append(attrs, fmt.Sprintf(`placeholder="%s"`, f.Placeholder))
	}
	if f.ClassName != "" {
		attrs = append(attrs, fmt.Sprintf(`className="%s"`, f.ClassName))
	}

	reg := fmt.Sprintf("{...%s.register('%s'", formName, f.Name)
	if f.Type == "number" {
		reg += ", { valueAsNumber: true }"
	}
	reg += ")}"

	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}

	return fmt.Sprintf("%s<input%s %s />", ind, attrStr, reg)
}

// renderEachJSX generates JSX for an EachBlock.
func renderEachJSX(e parser.EachBlock, dataVar string, indent int) string {
	ind := indentStr(indent)
	tag := orDefault(e.Tag, "div")
	cls := clsAttr(e.ClassName)
	itemTag := orDefault(e.ItemTag, "div")
	itemCls := clsAttr(e.ItemClassName)

	var lines []string
	lines = append(lines, fmt.Sprintf("%s<%s%s>", ind, tag, cls))
	lines = append(lines, fmt.Sprintf("%s  {%s.%s?.map((item: any, index: number) => (", ind, dataVar, e.Field))
	lines = append(lines, fmt.Sprintf("%s    <%s key={index}%s>", ind, itemTag, itemCls))

	if len(e.Children) > 0 {
		lines = append(lines, renderChildNodes(e.Children, "item", "item", indent+6)...)
	} else {
		for _, b := range e.Binds {
			lines = append(lines, renderBindJSX(b, "item", indent+6))
		}
	}

	lines = append(lines, fmt.Sprintf("%s    </%s>", ind, itemTag))
	lines = append(lines, fmt.Sprintf("%s  ))}", ind))
	lines = append(lines, fmt.Sprintf("%s</%s>", ind, tag))

	return strings.Join(lines, "\n")
}

// renderStateJSX generates JSX for a StateBind.
func renderStateJSX(s parser.StateBind, dataVar string, indent int) string {
	ind := indentStr(indent)
	tag := orDefault(s.Tag, "div")
	cls := clsAttr(s.ClassName)

	cond := ""
	switch {
	case strings.HasSuffix(s.Condition, ".empty"):
		field := strings.TrimSuffix(s.Condition, ".empty")
		cond = fmt.Sprintf("%s.%s?.length === 0", dataVar, field)
	case strings.HasSuffix(s.Condition, ".loading"):
		cond = dataVar + "Loading"
	case strings.HasSuffix(s.Condition, ".error"):
		cond = dataVar + "Error"
	default:
		cond = fmt.Sprintf("%s.%s", dataVar, s.Condition)
	}

	// Has children (e.g. action inside state)
	if len(s.Children) > 0 {
		var lines []string
		lines = append(lines, fmt.Sprintf("%s{%s && (", ind, cond))
		lines = append(lines, fmt.Sprintf("%s  <%s%s>", ind, tag, cls))
		lines = append(lines, renderChildNodes(s.Children, dataVar, "item", indent+4)...)
		lines = append(lines, fmt.Sprintf("%s  </%s>", ind, tag))
		lines = append(lines, fmt.Sprintf("%s)}", ind))
		return strings.Join(lines, "\n")
	}

	// Simple text
	text := orDefault(s.Text, "")
	if text != "" {
		return fmt.Sprintf("%s{%s && <%s%s>%s</%s>}", ind, cond, tag, cls, text, tag)
	}

	return fmt.Sprintf("%s{%s && <%s%s />}", ind, cond, tag, cls)
}

// renderComponentJSX generates JSX for a ComponentRef.
func renderComponentJSX(c parser.ComponentRef, dataVar string, indent int) string {
	ind := indentStr(indent)
	if c.Bind != "" {
		return fmt.Sprintf("%s<%s data={%s.%s} />", ind, c.Name, dataVar, c.Bind)
	}
	return fmt.Sprintf("%s<%s />", ind, c.Name)
}

// renderUseParams generates useParams destructuring for route params.
func renderUseParams(params []parser.ParamBind) string {
	var routeParams []string
	seen := map[string]bool{}
	for _, p := range params {
		if strings.HasPrefix(p.Source, "route.") {
			name := strings.TrimPrefix(p.Source, "route.")
			if !seen[name] {
				routeParams = append(routeParams, name)
				seen[name] = true
			}
		}
	}
	if len(routeParams) == 0 {
		return ""
	}
	return fmt.Sprintf("const { %s } = useParams()", strings.Join(routeParams, ", "))
}

// --- helpers ---

func renderParamValues(params []parser.ParamBind) string {
	var parts []string
	for _, p := range params {
		parts = append(parts, paramSourceExpr(p))
	}
	return strings.Join(parts, ", ")
}

func renderParamArgs(params []parser.ParamBind) string {
	if len(params) == 0 {
		return ""
	}
	var parts []string
	for _, p := range params {
		parts = append(parts, fmt.Sprintf("%s: %s", p.Name, paramSourceExpr(p)))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func paramSourceExpr(p parser.ParamBind) string {
	if strings.HasPrefix(p.Source, "route.") {
		return strings.TrimPrefix(p.Source, "route.")
	}
	return p.Source
}

func toLowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func toUpperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func indentStr(n int) string {
	return strings.Repeat(" ", n)
}

func clsAttr(className string) string {
	if className == "" {
		return ""
	}
	return fmt.Sprintf(` className="%s"`, className)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
