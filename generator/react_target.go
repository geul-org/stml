package generator

import (
	"fmt"
	"strings"

	"github.com/geul-org/stml/parser"
)

// ReactTarget generates React TSX components.
type ReactTarget struct{}

// renderFetchHooks writes useState + useQuery hook declarations.
func renderFetchHooks(f parser.FetchBlock, sb *strings.Builder) {
	if f.Paginate {
		defaultLimit := 20
		sb.WriteString(fmt.Sprintf("  const [page, setPage] = useState(1)\n"))
		sb.WriteString(fmt.Sprintf("  const [limit] = useState(%d)\n", defaultLimit))
	}
	if f.Sort != nil {
		sb.WriteString(fmt.Sprintf("  const [sortBy, setSortBy] = useState('%s')\n", f.Sort.Column))
		sb.WriteString(fmt.Sprintf("  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('%s')\n", f.Sort.Direction))
	}
	if len(f.Filters) > 0 {
		sb.WriteString("  const [filters, setFilters] = useState<Record<string, string>>({})\n")
	}
	if f.Paginate || f.Sort != nil || len(f.Filters) > 0 {
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("  %s\n\n", renderUseQuery(f)))
	for _, child := range f.NestedFetches {
		renderFetchHooks(child, sb)
	}
}

func (r *ReactTarget) FileExtension() string {
	return ".tsx"
}

func (r *ReactTarget) Dependencies(pages []parser.PageSpec) map[string]string {
	deps := map[string]string{}
	for _, page := range pages {
		is := collectImports(page, "")
		if is.useQuery || is.useMutation || is.useQueryClient {
			deps["@tanstack/react-query"] = "^5"
		}
		if is.useForm {
			deps["react-hook-form"] = "^7"
		}
		if is.useParams {
			deps["react-router-dom"] = "^6"
		}
	}
	return deps
}

func (r *ReactTarget) GeneratePage(page parser.PageSpec, specsDir string, opt GenerateOptions) string {
	is := collectImports(page, specsDir)
	componentName := toComponentName(page.Name)

	// Collect fetch operationIds for mutation onSuccess
	var fetchOps []string
	for _, f := range page.Fetches {
		fetchOps = collectFetchOps(f, fetchOps)
	}

	// Collect ALL actions including nested ones
	allActions := append([]parser.ActionBlock{}, page.Actions...)
	allActions = append(allActions, collectAllActions(page.Children)...)
	allActions = deduplicateActions(allActions)

	// Check if any action needs a form
	needsForm := false
	for _, a := range allActions {
		if len(a.Fields) > 0 {
			needsForm = true
			break
		}
	}
	is.useForm = needsForm

	if len(allActions) > 0 {
		is.useMutation = true
		is.useQueryClient = true
	}

	var sb strings.Builder

	// Imports
	sb.WriteString(renderImports(is, opt))
	sb.WriteString("\n\n")

	// Component
	sb.WriteString(fmt.Sprintf("export default function %s() {\n", componentName))

	// useParams
	allParams := collectAllParams(page)
	if up := renderUseParams(allParams); up != "" {
		sb.WriteString(fmt.Sprintf("  %s\n", up))
	}

	// useQueryClient
	if is.useQueryClient {
		sb.WriteString("  const queryClient = useQueryClient()\n")
	}

	sb.WriteString("\n")

	// useQuery hooks
	for _, f := range page.Fetches {
		renderFetchHooks(f, &sb)
	}

	// useForm + useMutation hooks
	for _, a := range allActions {
		if len(a.Fields) > 0 {
			sb.WriteString(fmt.Sprintf("  %s\n", renderFormHook(a)))
		}
		sb.WriteString(fmt.Sprintf("  %s\n\n", renderUseMutation(a, fetchOps)))
	}

	// JSX return
	sb.WriteString("  return (\n")

	if len(page.Children) > 0 {
		children := page.Children
		rootTag := "div"
		rootCls := ""
		if len(children) == 1 && children[0].Kind == "static" {
			root := children[0].Static
			rootTag = root.Tag
			rootCls = root.ClassName
			children = root.Children
		}
		sb.WriteString(fmt.Sprintf("    <%s%s>\n", rootTag, clsAttr(rootCls)))
		for _, line := range renderChildNodes(children, "", "item", 6) {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("    </%s>\n", rootTag))
	} else {
		sb.WriteString("    <div>\n")
		for _, f := range page.Fetches {
			sb.WriteString(renderFetchJSX(f, 6))
			sb.WriteString("\n")
		}
		for _, a := range page.Actions {
			sb.WriteString(renderActionJSX(a, 6))
			sb.WriteString("\n")
		}
		sb.WriteString("    </div>\n")
	}

	sb.WriteString("  )\n")
	sb.WriteString("}\n")

	return sb.String()
}
