package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/geul-org/stml/artifacts/internal/parser"
)

// Generate produces React TSX files from parsed PageSpecs.
func Generate(pages []parser.PageSpec, specsDir, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outDir, err)
	}

	for _, page := range pages {
		code := GeneratePage(page, specsDir)
		path := filepath.Join(outDir, page.Name+".tsx")
		if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

// GeneratePage generates the TSX source code for a single page.
func GeneratePage(page parser.PageSpec, specsDir string) string {
	is := collectImports(page, specsDir)
	componentName := toComponentName(page.Name)

	// Collect fetch operationIds for mutation onSuccess
	var fetchOps []string
	for _, f := range page.Fetches {
		fetchOps = collectFetchOps(f, fetchOps)
	}

	// Collect ALL actions including nested ones (e.g. inside fetch > state)
	allActions := append([]parser.ActionBlock{}, page.Actions...)
	allActions = append(allActions, collectAllActions(page.Children)...)
	// Deduplicate by OperationID
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

	// Update import flags if nested actions exist
	if len(allActions) > 0 {
		is.useMutation = true
		is.useQueryClient = true
	}

	var sb strings.Builder

	// Imports
	sb.WriteString(renderImports(is))
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
		// Use ChildNode tree for DOM-order rendering
		// If single static root (e.g. <main>), unwrap it to avoid duplication
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

// renderFetchHooks writes useState + useQuery hook declarations.
func renderFetchHooks(f parser.FetchBlock, sb *strings.Builder) {
	// Phase 5: useState hooks for infra params
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

// collectAllParams gathers all ParamBinds from the page (including nested actions).
func collectAllParams(page parser.PageSpec) []parser.ParamBind {
	var params []parser.ParamBind
	for _, f := range page.Fetches {
		params = collectFetchParamBinds(f, params)
	}
	for _, a := range page.Actions {
		params = append(params, a.Params...)
	}
	// Also collect params from nested actions in ChildNode tree
	for _, a := range collectAllActions(page.Children) {
		params = append(params, a.Params...)
	}
	return params
}

// deduplicateActions removes duplicate actions by OperationID.
func deduplicateActions(actions []parser.ActionBlock) []parser.ActionBlock {
	seen := map[string]bool{}
	var result []parser.ActionBlock
	for _, a := range actions {
		if !seen[a.OperationID] {
			seen[a.OperationID] = true
			result = append(result, a)
		}
	}
	return result
}

func collectFetchParamBinds(f parser.FetchBlock, params []parser.ParamBind) []parser.ParamBind {
	params = append(params, f.Params...)
	for _, child := range f.NestedFetches {
		params = collectFetchParamBinds(child, params)
	}
	return params
}

func collectFetchOps(f parser.FetchBlock, ops []string) []string {
	ops = append(ops, f.OperationID)
	for _, child := range f.NestedFetches {
		ops = collectFetchOps(child, ops)
	}
	return ops
}

// findRootElement determines the root JSX element from page children.
func findRootElement(page parser.PageSpec) (string, string) {
	// If children start with a single static wrapper (e.g. <main>), use it
	if len(page.Children) == 1 && page.Children[0].Kind == "static" {
		se := page.Children[0].Static
		return se.Tag, se.ClassName
	}
	return "div", ""
}

// collectAllActions walks the ChildNode tree and collects all ActionBlocks
// (including those nested inside fetch/state/static blocks).
func collectAllActions(nodes []parser.ChildNode) []parser.ActionBlock {
	var actions []parser.ActionBlock
	for _, ch := range nodes {
		switch ch.Kind {
		case "action":
			actions = append(actions, *ch.Action)
		case "fetch":
			actions = append(actions, collectAllActions(ch.Fetch.Children)...)
		case "state":
			actions = append(actions, collectAllActions(ch.State.Children)...)
		case "static":
			actions = append(actions, collectAllActions(ch.Static.Children)...)
		case "each":
			actions = append(actions, collectAllActions(ch.Each.Children)...)
		}
	}
	return actions
}

// toComponentName converts "my-reservations-page" to "MyReservationsPage".
func toComponentName(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		parts[i] = toUpperFirst(p)
	}
	return strings.Join(parts, "")
}
