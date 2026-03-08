package parser

// PageSpec represents a single STML page parsed from an HTML file.
type PageSpec struct {
	Name     string        // page name derived from filename (e.g. "login-page")
	FileName string        // original filename (e.g. "login-page.html")
	Fetches  []FetchBlock  // data-fetch blocks (for validation)
	Actions  []ActionBlock // data-action blocks (for validation)
	Children []ChildNode   // all top-level children in DOM order (for codegen)
}

// FetchBlock represents a data-fetch element and its descendant bindings.
type FetchBlock struct {
	Tag         string         // original HTML tag (e.g. "section")
	ClassName   string         // class attribute value
	OperationID string         // data-fetch value (e.g. "ListMyReservations")
	Params      []ParamBind    // data-param-* attributes on this element
	Binds       []FieldBind    // descendant data-bind attributes (for validation)
	Eaches      []EachBlock    // descendant data-each blocks (for validation)
	States      []StateBind    // descendant data-state attributes (for validation)
	Components  []ComponentRef // descendant data-component attributes (for validation)
	Children    []ChildNode    // all children in DOM order (for codegen)
	NestedFetches []FetchBlock // nested data-fetch blocks (for validation)

	// Phase 5: infra params
	Paginate bool       // data-paginate present
	Sort     *SortDecl  // data-sort parsed result
	Filters  []string   // data-filter comma-separated columns
	Includes []string   // data-include comma-separated resources
}

// SortDecl represents a parsed data-sort value.
type SortDecl struct {
	Column    string // default sort column
	Direction string // "asc" or "desc"
}

// ActionBlock represents a data-action element and its descendant fields.
type ActionBlock struct {
	Tag         string      // original HTML tag
	ClassName   string      // class attribute value
	OperationID string      // data-action value (e.g. "CreateReservation")
	Params      []ParamBind // data-param-* attributes on this element
	Fields      []FieldBind // descendant data-field attributes (for validation)
	Children    []ChildNode // all children in DOM order (for codegen)
	SubmitText  string      // text of button[type=submit]
}

// ParamBind represents a data-param-* attribute.
type ParamBind struct {
	Name   string // parameter name (e.g. "ReservationID")
	Source string // value source (e.g. "route.ReservationID")
}

// FieldBind represents a data-bind or data-field attribute.
type FieldBind struct {
	Name        string // field name (e.g. "Status", "Email")
	Tag         string // HTML tag (e.g. "span", "input")
	Type        string // input type attribute (e.g. "hidden", "number", "")
	ClassName   string // class attribute value
	Placeholder string // placeholder attribute value
}

// EachBlock represents a data-each element and its descendant bindings.
type EachBlock struct {
	Tag           string         // original HTML tag (e.g. "ul")
	ClassName     string         // class attribute value
	ItemTag       string         // repeated item tag (e.g. "li")
	ItemClassName string         // repeated item class
	Field         string         // array field name (e.g. "reservations")
	Binds         []FieldBind    // data-bind inside the loop (for validation)
	States        []StateBind    // data-state inside the loop (for validation)
	Components    []ComponentRef // data-component inside the loop (for validation)
	Children      []ChildNode    // item children in DOM order (for codegen)
}

// StateBind represents a data-state attribute.
type StateBind struct {
	Tag       string      // original HTML tag (e.g. "p", "footer")
	ClassName string      // class attribute value
	Condition string      // condition expression (e.g. "reservations.empty", "canDelete")
	Text      string      // text content (e.g. "예약이 없습니다")
	Children  []ChildNode // nested children (e.g. action buttons inside state)
}

// ComponentRef represents a data-component attribute.
type ComponentRef struct {
	Name      string // component name (e.g. "DatePicker")
	Bind      string // data-bind value if present
	Field     string // data-field value if present
	ClassName string // class attribute value
}

// StaticElement represents a non-binding HTML element to preserve in output.
type StaticElement struct {
	Tag       string          // HTML tag (e.g. "header", "h1")
	ClassName string          // class attribute value
	Text      string          // direct text content
	Children  []ChildNode     // nested children
}

// ChildNode represents any child element inside a block, preserving DOM order.
type ChildNode struct {
	Kind      string          // "bind", "each", "state", "component", "static", "action", "fetch"
	Bind      *FieldBind
	Each      *EachBlock
	State     *StateBind
	Component *ComponentRef
	Static    *StaticElement
	Action    *ActionBlock
	Fetch     *FetchBlock
}
