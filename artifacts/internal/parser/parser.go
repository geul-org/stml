package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

// ParseDir parses all .html files in the given directory and returns a PageSpec for each.
func ParseDir(dir string) ([]PageSpec, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	var pages []PageSpec
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		page, err := ParseFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		pages = append(pages, page)
	}
	return pages, nil
}

// ParseFile parses a single HTML file and returns a PageSpec.
func ParseFile(path string) (PageSpec, error) {
	f, err := os.Open(path)
	if err != nil {
		return PageSpec{}, err
	}
	defer f.Close()

	return ParseReader(filepath.Base(path), f)
}

// ParseReader parses HTML from a reader and returns a PageSpec.
func ParseReader(filename string, r io.Reader) (PageSpec, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return PageSpec{}, fmt.Errorf("html parse: %w", err)
	}

	name := strings.TrimSuffix(filename, ".html")
	page := PageSpec{
		Name:     name,
		FileName: filename,
	}

	walkTopLevel(doc, &page)
	return page, nil
}

// isImplicitTag returns true for tags the HTML parser auto-generates.
func isImplicitTag(tag string) bool {
	return tag == "html" || tag == "head" || tag == "body"
}

// walkTopLevel traverses the DOM tree collecting top-level blocks.
func walkTopLevel(n *html.Node, page *page) {
	if n.Type == html.ElementNode && isImplicitTag(n.Data) {
		// Skip implicit html/head/body, walk their children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkTopLevel(c, page)
		}
		return
	}
	if n.Type == html.ElementNode {
		if op := getAttr(n, "data-fetch"); op != "" {
			fb := parseFetchBlock(n, op)
			page.Fetches = append(page.Fetches, fb)
			page.Children = append(page.Children, ChildNode{Kind: "fetch", Fetch: &fb})
			return
		}
		if op := getAttr(n, "data-action"); op != "" {
			ab := parseActionBlock(n, op)
			page.Actions = append(page.Actions, ab)
			page.Children = append(page.Children, ChildNode{Kind: "action", Action: &ab})
			return
		}
		// Static element with possible nested data-* children
		if hasDescendantData(n) {
			se := parseStaticWithDataChildren(n, page)
			page.Children = append(page.Children, ChildNode{Kind: "static", Static: &se})
			return
		}
		// Pure static
		if hasContent(n) {
			se := parseStaticElement(n)
			page.Children = append(page.Children, ChildNode{Kind: "static", Static: &se})
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkTopLevel(c, page)
	}
}

// parseFetchBlock builds a FetchBlock from a data-fetch element and its descendants.
func parseFetchBlock(n *html.Node, operationID string) FetchBlock {
	fb := FetchBlock{
		Tag:         n.Data,
		ClassName:   getAttr(n, "class"),
		OperationID: operationID,
		Params:      extractParams(n),
	}

	// Phase 5: infra params
	if hasAttr(n, "data-paginate") {
		fb.Paginate = true
	}
	if v := getAttr(n, "data-sort"); v != "" {
		fb.Sort = parseSortDecl(v)
	}
	if v := getAttr(n, "data-filter"); v != "" {
		fb.Filters = splitTrim(v)
	}
	if v := getAttr(n, "data-include"); v != "" {
		fb.Includes = splitTrim(v)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkFetchChildren(c, &fb)
	}
	return fb
}

// walkFetchChildren recursively collects bindings inside a fetch block.
func walkFetchChildren(n *html.Node, fb *FetchBlock) {
	if n.Type == html.TextNode {
		return
	}
	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkFetchChildren(c, fb)
		}
		return
	}

	// Nested data-fetch
	if op := getAttr(n, "data-fetch"); op != "" {
		child := parseFetchBlock(n, op)
		fb.NestedFetches = append(fb.NestedFetches, child)
		fb.Children = append(fb.Children, ChildNode{Kind: "fetch", Fetch: &child})
		return
	}

	// data-action inside fetch
	if op := getAttr(n, "data-action"); op != "" {
		ab := parseActionBlock(n, op)
		fb.Children = append(fb.Children, ChildNode{Kind: "action", Action: &ab})
		return
	}

	// data-each
	if field := getAttr(n, "data-each"); field != "" {
		eb := parseEachBlock(n, field)
		fb.Eaches = append(fb.Eaches, eb)
		fb.Children = append(fb.Children, ChildNode{Kind: "each", Each: &eb})
		return
	}

	// data-bind
	if field := getAttr(n, "data-bind"); field != "" {
		bind := FieldBind{
			Name:      field,
			Tag:       n.Data,
			Type:      getAttr(n, "type"),
			ClassName: getAttr(n, "class"),
		}
		fb.Binds = append(fb.Binds, bind)
		fb.Children = append(fb.Children, ChildNode{Kind: "bind", Bind: &bind})
		return
	}

	// data-state
	if cond := getAttr(n, "data-state"); cond != "" {
		sb := parseStateBind(n, cond)
		fb.States = append(fb.States, sb)
		fb.Children = append(fb.Children, ChildNode{Kind: "state", State: &sb})
		return
	}

	// data-component
	if comp := getAttr(n, "data-component"); comp != "" {
		cr := ComponentRef{
			Name:      comp,
			Bind:      getAttr(n, "data-bind"),
			Field:     getAttr(n, "data-field"),
			ClassName: getAttr(n, "class"),
		}
		fb.Components = append(fb.Components, cr)
		fb.Children = append(fb.Children, ChildNode{Kind: "component", Component: &cr})
		// walk children in case HTML parser didn't self-close
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkFetchChildren(c, fb)
		}
		return
	}

	// Static element — preserve if it has content
	if hasContent(n) || hasDescendantDataInFetch(n) {
		se := parseStaticInFetch(n, fb)
		fb.Children = append(fb.Children, ChildNode{Kind: "static", Static: &se})
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkFetchChildren(c, fb)
	}
}

// parseStaticInFetch parses a static element inside a fetch block, but still
// collects any nested data-* elements into the parent fb for validation.
func parseStaticInFetch(n *html.Node, fb *FetchBlock) StaticElement {
	se := StaticElement{
		Tag:       n.Data,
		ClassName: getAttr(n, "class"),
		Text:      directText(n),
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			continue
		}
		if c.Type != html.ElementNode {
			continue
		}

		if op := getAttr(c, "data-fetch"); op != "" {
			child := parseFetchBlock(c, op)
			fb.NestedFetches = append(fb.NestedFetches, child)
			se.Children = append(se.Children, ChildNode{Kind: "fetch", Fetch: &child})
		} else if op := getAttr(c, "data-action"); op != "" {
			ab := parseActionBlock(c, op)
			se.Children = append(se.Children, ChildNode{Kind: "action", Action: &ab})
		} else if field := getAttr(c, "data-each"); field != "" {
			eb := parseEachBlock(c, field)
			fb.Eaches = append(fb.Eaches, eb)
			se.Children = append(se.Children, ChildNode{Kind: "each", Each: &eb})
		} else if field := getAttr(c, "data-bind"); field != "" {
			bind := FieldBind{Name: field, Tag: c.Data, Type: getAttr(c, "type"), ClassName: getAttr(c, "class")}
			fb.Binds = append(fb.Binds, bind)
			se.Children = append(se.Children, ChildNode{Kind: "bind", Bind: &bind})
		} else if cond := getAttr(c, "data-state"); cond != "" {
			sb := parseStateBind(c, cond)
			fb.States = append(fb.States, sb)
			se.Children = append(se.Children, ChildNode{Kind: "state", State: &sb})
		} else if comp := getAttr(c, "data-component"); comp != "" {
			cr := ComponentRef{Name: comp, Bind: getAttr(c, "data-bind"), Field: getAttr(c, "data-field"), ClassName: getAttr(c, "class")}
			fb.Components = append(fb.Components, cr)
			se.Children = append(se.Children, ChildNode{Kind: "component", Component: &cr})
		} else if hasContent(c) || hasDescendantDataInFetch(c) {
			childStatic := parseStaticInFetch(c, fb)
			se.Children = append(se.Children, ChildNode{Kind: "static", Static: &childStatic})
		}
	}
	return se
}

// parseEachBlock builds an EachBlock from a data-each element.
func parseEachBlock(n *html.Node, field string) EachBlock {
	eb := EachBlock{
		Tag:       n.Data,
		ClassName: getAttr(n, "class"),
		Field:     field,
	}
	// Find the first element child as the item template
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			eb.ItemTag = c.Data
			eb.ItemClassName = getAttr(c, "class")
			// Collect item children
			for gc := c.FirstChild; gc != nil; gc = gc.NextSibling {
				walkEachChildren(gc, &eb)
			}
			break
		}
	}
	return eb
}

// walkEachChildren recursively collects bindings inside an each block's item.
func walkEachChildren(n *html.Node, eb *EachBlock) {
	if n.Type != html.ElementNode {
		return
	}

	if field := getAttr(n, "data-bind"); field != "" {
		bind := FieldBind{
			Name:      field,
			Tag:       n.Data,
			Type:      getAttr(n, "type"),
			ClassName: getAttr(n, "class"),
		}
		eb.Binds = append(eb.Binds, bind)
		eb.Children = append(eb.Children, ChildNode{Kind: "bind", Bind: &bind})
		return
	}

	if cond := getAttr(n, "data-state"); cond != "" {
		sb := parseStateBind(n, cond)
		eb.States = append(eb.States, sb)
		eb.Children = append(eb.Children, ChildNode{Kind: "state", State: &sb})
		return
	}

	if comp := getAttr(n, "data-component"); comp != "" {
		cr := ComponentRef{
			Name:      comp,
			Bind:      getAttr(n, "data-bind"),
			Field:     getAttr(n, "data-field"),
			ClassName: getAttr(n, "class"),
		}
		eb.Components = append(eb.Components, cr)
		eb.Children = append(eb.Children, ChildNode{Kind: "component", Component: &cr})
		return
	}

	// Static wrapper inside each item
	if hasContent(n) {
		se := StaticElement{
			Tag:       n.Data,
			ClassName: getAttr(n, "class"),
			Text:      directText(n),
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				if bf := getAttr(c, "data-bind"); bf != "" {
					bind := FieldBind{Name: bf, Tag: c.Data, Type: getAttr(c, "type"), ClassName: getAttr(c, "class")}
					eb.Binds = append(eb.Binds, bind)
					se.Children = append(se.Children, ChildNode{Kind: "bind", Bind: &bind})
				} else {
					childSe := parseStaticElement(c)
					se.Children = append(se.Children, ChildNode{Kind: "static", Static: &childSe})
				}
			}
		}
		eb.Children = append(eb.Children, ChildNode{Kind: "static", Static: &se})
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkEachChildren(c, eb)
	}
}

// parseActionBlock builds an ActionBlock from a data-action element.
func parseActionBlock(n *html.Node, operationID string) ActionBlock {
	ab := ActionBlock{
		Tag:         n.Data,
		ClassName:   getAttr(n, "class"),
		OperationID: operationID,
		Params:      extractParams(n),
	}
	// If the action element is a button itself, capture its text
	if n.Data == "button" {
		ab.SubmitText = directText(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkActionChildren(c, &ab)
	}
	return ab
}

// walkActionChildren recursively collects fields inside an action block.
func walkActionChildren(n *html.Node, ab *ActionBlock) {
	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkActionChildren(c, ab)
		}
		return
	}

	// data-component with data-field takes priority
	if comp := getAttr(n, "data-component"); comp != "" {
		if f := getAttr(n, "data-field"); f != "" {
			bind := FieldBind{
				Name:      f,
				Tag:       "data-component:" + comp,
				ClassName: getAttr(n, "class"),
			}
			ab.Fields = append(ab.Fields, bind)
			ab.Children = append(ab.Children, ChildNode{Kind: "bind", Bind: &bind})
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkActionChildren(c, ab)
		}
		return
	}

	if field := getAttr(n, "data-field"); field != "" {
		bind := FieldBind{
			Name:        field,
			Tag:         n.Data,
			Type:        getAttr(n, "type"),
			ClassName:   getAttr(n, "class"),
			Placeholder: getAttr(n, "placeholder"),
		}
		ab.Fields = append(ab.Fields, bind)
		ab.Children = append(ab.Children, ChildNode{Kind: "bind", Bind: &bind})
		return
	}

	// button[type=submit] — extract text
	if n.Data == "button" && getAttr(n, "type") == "submit" {
		ab.SubmitText = directText(n)
		return
	}

	// Static elements inside action (e.g. labels, divs wrapping fields)
	if hasContent(n) || hasDescendantField(n) {
		se := parseStaticInAction(n, ab)
		ab.Children = append(ab.Children, ChildNode{Kind: "static", Static: &se})
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkActionChildren(c, ab)
	}
}

// parseStaticInAction parses a static element inside an action block.
func parseStaticInAction(n *html.Node, ab *ActionBlock) StaticElement {
	se := StaticElement{
		Tag:       n.Data,
		ClassName: getAttr(n, "class"),
		Text:      directText(n),
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		if comp := getAttr(c, "data-component"); comp != "" {
			if f := getAttr(c, "data-field"); f != "" {
				bind := FieldBind{Name: f, Tag: "data-component:" + comp, ClassName: getAttr(c, "class")}
				ab.Fields = append(ab.Fields, bind)
				se.Children = append(se.Children, ChildNode{Kind: "bind", Bind: &bind})
			}
			// Walk children for nested components (HTML5 self-close issue)
			for gc := c.FirstChild; gc != nil; gc = gc.NextSibling {
				if gc.Type == html.ElementNode {
					walkStaticActionChild(gc, ab, &se)
				}
			}
		} else if field := getAttr(c, "data-field"); field != "" {
			bind := FieldBind{Name: field, Tag: c.Data, Type: getAttr(c, "type"), ClassName: getAttr(c, "class"), Placeholder: getAttr(c, "placeholder")}
			ab.Fields = append(ab.Fields, bind)
			se.Children = append(se.Children, ChildNode{Kind: "bind", Bind: &bind})
		} else if c.Data == "button" && getAttr(c, "type") == "submit" {
			ab.SubmitText = directText(c)
		} else if hasContent(c) || hasDescendantField(c) {
			childSe := parseStaticInAction(c, ab)
			se.Children = append(se.Children, ChildNode{Kind: "static", Static: &childSe})
		}
	}
	return se
}

// walkStaticActionChild handles a single child element inside a static action wrapper.
func walkStaticActionChild(c *html.Node, ab *ActionBlock, se *StaticElement) {
	if comp := getAttr(c, "data-component"); comp != "" {
		if f := getAttr(c, "data-field"); f != "" {
			bind := FieldBind{Name: f, Tag: "data-component:" + comp, ClassName: getAttr(c, "class")}
			ab.Fields = append(ab.Fields, bind)
			se.Children = append(se.Children, ChildNode{Kind: "bind", Bind: &bind})
		}
		for gc := c.FirstChild; gc != nil; gc = gc.NextSibling {
			if gc.Type == html.ElementNode {
				walkStaticActionChild(gc, ab, se)
			}
		}
	} else if field := getAttr(c, "data-field"); field != "" {
		bind := FieldBind{Name: field, Tag: c.Data, Type: getAttr(c, "type"), ClassName: getAttr(c, "class"), Placeholder: getAttr(c, "placeholder")}
		ab.Fields = append(ab.Fields, bind)
		se.Children = append(se.Children, ChildNode{Kind: "bind", Bind: &bind})
	} else if c.Data == "button" && getAttr(c, "type") == "submit" {
		ab.SubmitText = directText(c)
	}
}

// parseStateBind builds a StateBind from a data-state element.
func parseStateBind(n *html.Node, condition string) StateBind {
	sb := StateBind{
		Tag:       n.Data,
		ClassName: getAttr(n, "class"),
		Condition: condition,
		Text:      directText(n),
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		if op := getAttr(c, "data-action"); op != "" {
			ab := parseActionBlock(c, op)
			sb.Children = append(sb.Children, ChildNode{Kind: "action", Action: &ab})
		} else if hasContent(c) {
			se := parseStaticElement(c)
			sb.Children = append(sb.Children, ChildNode{Kind: "static", Static: &se})
		}
	}
	return sb
}

// parseStaticElement recursively parses a non-binding HTML element.
func parseStaticElement(n *html.Node) StaticElement {
	se := StaticElement{
		Tag:       n.Data,
		ClassName: getAttr(n, "class"),
		Text:      directText(n),
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && hasContent(c) {
			child := parseStaticElement(c)
			se.Children = append(se.Children, ChildNode{Kind: "static", Static: &child})
		}
	}
	return se
}

// parseStaticWithDataChildren parses a static element that has data-* descendants.
func parseStaticWithDataChildren(n *html.Node, page *page) StaticElement {
	se := StaticElement{
		Tag:       n.Data,
		ClassName: getAttr(n, "class"),
		Text:      directText(n),
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		if op := getAttr(c, "data-fetch"); op != "" {
			fb := parseFetchBlock(c, op)
			page.Fetches = append(page.Fetches, fb)
			se.Children = append(se.Children, ChildNode{Kind: "fetch", Fetch: &fb})
		} else if op := getAttr(c, "data-action"); op != "" {
			ab := parseActionBlock(c, op)
			page.Actions = append(page.Actions, ab)
			se.Children = append(se.Children, ChildNode{Kind: "action", Action: &ab})
		} else if hasDescendantData(c) {
			child := parseStaticWithDataChildren(c, page)
			se.Children = append(se.Children, ChildNode{Kind: "static", Static: &child})
		} else if hasContent(c) {
			child := parseStaticElement(c)
			se.Children = append(se.Children, ChildNode{Kind: "static", Static: &child})
		}
	}
	return se
}

// --- type alias to avoid import cycle in walkTopLevel ---
type page = PageSpec

// --- helpers ---

// extractParams extracts data-param-* attributes from an element.
func extractParams(n *html.Node) []ParamBind {
	var params []ParamBind
	for _, attr := range n.Attr {
		if strings.HasPrefix(attr.Key, "data-param-") {
			paramName := attr.Key[len("data-param-"):]
			params = append(params, ParamBind{
				Name:   kebabToCamel(paramName),
				Source: attr.Val,
			})
		}
	}
	return params
}

// kebabToCamel converts kebab-case to camelCase.
func kebabToCamel(s string) string {
	if !strings.Contains(s, "-") {
		return s
	}
	parts := strings.Split(s, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// getAttr returns the value of the named attribute, or "" if not found.
func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// directText extracts the first non-empty direct text child of an element.
func directText(n *html.Node) string {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text := strings.TrimSpace(c.Data)
			if text != "" {
				return text
			}
		}
	}
	return ""
}

// hasContent returns true if the element has text or element children.
func hasContent(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
			return true
		}
		if c.Type == html.ElementNode {
			return true
		}
	}
	return false
}

// hasDescendantData checks if any descendant has a data-fetch or data-action attribute.
func hasDescendantData(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			if getAttr(c, "data-fetch") != "" || getAttr(c, "data-action") != "" {
				return true
			}
			if hasDescendantData(c) {
				return true
			}
		}
	}
	return false
}

// hasDescendantDataInFetch checks if any descendant has data-* attributes relevant to fetch.
func hasDescendantDataInFetch(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			for _, attr := range c.Attr {
				if strings.HasPrefix(attr.Key, "data-") {
					return true
				}
			}
			if hasDescendantDataInFetch(c) {
				return true
			}
		}
	}
	return false
}

// hasAttr returns true if the element has the named attribute (regardless of value).
func hasAttr(n *html.Node, key string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return true
		}
	}
	return false
}

// parseSortDecl parses "column:direction" into a SortDecl.
func parseSortDecl(v string) *SortDecl {
	parts := strings.SplitN(v, ":", 2)
	sd := &SortDecl{Column: strings.TrimSpace(parts[0]), Direction: "asc"}
	if len(parts) == 2 {
		sd.Direction = strings.TrimSpace(parts[1])
	}
	return sd
}

// splitTrim splits a comma-separated string and trims whitespace.
func splitTrim(v string) []string {
	raw := strings.Split(v, ",")
	var result []string
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// hasDescendantField checks if any descendant has data-field or data-component with data-field.
func hasDescendantField(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			if getAttr(c, "data-field") != "" {
				return true
			}
			if getAttr(c, "data-component") != "" && getAttr(c, "data-field") != "" {
				return true
			}
			if hasDescendantField(c) {
				return true
			}
		}
	}
	return false
}
