# Phase 4: Codegen Upgrade 01

> HTML 원본의 시각적 정보를 보존하여 코드젠 출력 품질을 프로덕션 수준으로 올린다.

## 목표

현재 코드젠은 API 바인딩(훅, 폼, mutation)은 정확하지만 HTML 원본의 구조, 스타일, 텍스트를 전부 버린다. 파서의 정보 수집량을 늘리고 템플릿을 개선하여, STML HTML에 작성된 모든 시각적 의도를 TSX에 보존한다.

## 범위

코드젠개선방안.md의 항목 1~7 (타입 생성 제외)

| # | 항목 | 변경 대상 |
|---|---|---|
| 1 | HTML 구조 보존 | 파서 + 템플릿 |
| 2 | Tailwind 클래스 보존 | 파서 + 템플릿 |
| 3 | 텍스트 콘텐츠 보존 | 파서 + 템플릿 |
| 4 | input 속성 보존 | 파서 + 템플릿 |
| 5 | 필드 없는 action → 버튼 | 템플릿 |
| 6 | data-state 조건부 렌더링 | 파서 + 템플릿 |
| 7 | onSuccess 범위 한정 | 템플릿 |

## Step 1. 파서 구조체 확장 (`types.go`)

### 공통: 모든 블록에 Tag, ClassName 추가

```go
type FetchBlock struct {
    Tag         string         // 원본 태그 (e.g. "section", "article")
    ClassName   string         // class 속성값 (e.g. "flex-1 p-4")
    OperationID string
    Params      []ParamBind
    Binds       []FieldBind
    Eaches      []EachBlock
    States      []StateBind
    Components  []ComponentRef
    Children    []FetchBlock
}

type ActionBlock struct {
    Tag         string         // 원본 태그 (e.g. "div", "form")
    ClassName   string         // class 속성값
    OperationID string
    Params      []ParamBind
    Fields      []FieldBind
    SubmitText  string         // button[type=submit]의 텍스트 (e.g. "예약하기")
}

type EachBlock struct {
    Tag        string         // 원본 태그 (e.g. "ul")
    ClassName  string
    ItemTag    string         // 반복 항목 태그 (e.g. "li")
    ItemClassName string
    Field      string
    Binds      []FieldBind
    States     []StateBind
    Components []ComponentRef
}

type StateBind struct {
    Tag       string  // 원본 태그 (e.g. "p", "footer")
    ClassName string
    Condition string
    Text      string  // 텍스트 콘텐츠 (e.g. "예약이 없습니다")
    Children  []any   // 내부에 data-action 등이 있을 수 있음
}
```

### FieldBind 확장

```go
type FieldBind struct {
    Name        string
    Tag         string
    Type        string
    ClassName   string            // class 속성값
    Placeholder string            // placeholder 속성값
}
```

### 비-바인딩 요소 보존

data-* 속성이 없는 순수 HTML 요소(header, h1, h2, span 텍스트 등)도 보존해야 한다. 새 타입 도입:

```go
// StaticElement represents a non-binding HTML element to preserve in output.
type StaticElement struct {
    Tag       string
    ClassName string
    Text      string           // 텍스트 콘텐츠
    Children  []StaticElement
}
```

FetchBlock, ActionBlock, EachBlock에 `Statics []StaticElement` 필드 추가. 코드젠 시 바인딩 요소와 정적 요소를 원본 순서대로 출력한다.

### 순서 보존 문제

현재 파서는 Binds, Eaches, States, Components를 별도 슬라이스에 수집하므로 원본 DOM 순서가 사라진다. 순서를 보존하려면 통합 자식 타입이 필요하다:

```go
// ChildNode represents any child element inside a block, preserving DOM order.
type ChildNode struct {
    Kind      string         // "bind", "each", "state", "component", "static", "action", "fetch"
    Bind      *FieldBind
    Each      *EachBlock
    State     *StateBind
    Component *ComponentRef
    Static    *StaticElement
    Action    *ActionBlock
    Fetch     *FetchBlock
}
```

FetchBlock과 ActionBlock에 `Children []ChildNode` 를 추가하고, 기존 개별 슬라이스(Binds, Eaches 등)는 검증용으로 유지하되 코드젠은 Children 순서로 출력한다.

## Step 2. 파서 수정 (`parser.go`)

### 2-1. 속성 수집 확장

모든 data-* 요소에서 추가 수집:
- `n.Data` → Tag
- `getAttr(n, "class")` → ClassName
- `getAttr(n, "placeholder")` → Placeholder

### 2-2. 텍스트 수집

button[type=submit]의 첫 텍스트 자식 노드 → ActionBlock.SubmitText
data-state 요소의 텍스트 자식 노드 → StateBind.Text

```go
func extractText(n *html.Node) string {
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
```

### 2-3. ChildNode 순서 보존

walkFetchChildren, walkActionChildren에서 자식을 순회할 때, 각 요소를 해당 타입의 슬라이스에 추가하는 동시에 ChildNode로도 추가한다.

### 2-4. StaticElement 수집

data-* 속성이 없는 요소 중 텍스트나 자식이 있는 것은 StaticElement로 수집:

```go
if !hasDataAttr(n) {
    static := StaticElement{
        Tag:       n.Data,
        ClassName: getAttr(n, "class"),
        Text:      extractText(n),
    }
    // 재귀로 자식 StaticElement 수집
    for c := n.FirstChild; c != nil; c = c.NextSibling {
        // ...
    }
    parent.Children = append(parent.Children, ChildNode{Kind: "static", Static: &static})
}
```

### 2-5. data-state 내부 action 수집

data-state 요소 안에 data-action이 있으면(예: canCancel 안에 CancelReservation), StateBind.Children에 ActionBlock으로 수집한다. 이를 위해 StateBind의 Children을 `[]ChildNode`로 정의한다.

## Step 3. 템플릿 수정 (`templates.go`)

### 3-1. 태그 + className 출력

모든 JSX 렌더링 함수에서:

```go
// 현재
fmt.Sprintf("%s<div>", ind)

// 개선
tag := f.Tag
if tag == "" { tag = "div" }
cls := ""
if f.ClassName != "" { cls = fmt.Sprintf(` className="%s"`, f.ClassName) }
fmt.Sprintf("%s<%s%s>", ind, tag, cls)
```

### 3-2. ChildNode 순서 출력

renderFetchJSX를 ChildNode 기반으로 재작성:

```go
for _, child := range f.Children {
    switch child.Kind {
    case "bind":
        lines = append(lines, renderBindJSX(*child.Bind, alias, indent+2))
    case "each":
        lines = append(lines, renderEachJSX(*child.Each, alias, indent+2))
    case "state":
        lines = append(lines, renderStateJSX(*child.State, alias, indent+2))
    case "component":
        lines = append(lines, renderComponentJSX(*child.Component, alias, indent+2))
    case "static":
        lines = append(lines, renderStaticJSX(*child.Static, indent+2))
    case "action":
        lines = append(lines, renderActionJSX(*child.Action, indent+2))
    }
}
```

### 3-3. 텍스트 보존

```go
// ActionBlock 제출 버튼
submitText := a.SubmitText
if submitText == "" { submitText = "제출" }
fmt.Sprintf(`<button type="submit"%s>%s</button>`, cls, submitText)

// StateBind 텍스트
fmt.Sprintf(`<%s%s>%s</%s>`, s.Tag, cls, s.Text, s.Tag)
```

### 3-4. input 속성 보존

```go
func renderFieldJSX(f FieldBind, formName string, indent int) string {
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
    // register
    // ...
}
```

### 3-5. 필드 없는 action → 버튼

```go
func renderActionJSX(a ActionBlock, indent int) string {
    if len(a.Fields) == 0 {
        // button onClick
        mutName := toLowerFirst(a.OperationID) + "Mutation"
        return fmt.Sprintf(`%s<button onClick={() => %s.mutate({})}%s>%s</button>`,
            ind, mutName, cls, a.SubmitText)
    }
    // 기존 form 렌더링
}
```

### 3-6. data-state 조건부 렌더링

```go
func renderStateJSX(s StateBind, dataVar string, indent int) string {
    cond := ""
    switch {
    case strings.HasSuffix(s.Condition, ".empty"):
        field := strings.TrimSuffix(s.Condition, ".empty")
        cond = fmt.Sprintf("%s.%s?.length === 0", dataVar, field)
    case strings.HasSuffix(s.Condition, ".loading"):
        // ...
    default:
        // plain boolean: 부모 fetch 데이터에서 참조
        cond = fmt.Sprintf("%s.%s", dataVar, s.Condition)
    }

    // 자식이 있으면 블록으로, 없으면 텍스트 한 줄
    if len(s.Children) > 0 {
        // ChildNode 기반 렌더링
    } else {
        return fmt.Sprintf("%s{%s && <%s%s>%s</%s>}", ind, cond, s.Tag, cls, s.Text, s.Tag)
    }
}
```

### 3-7. onSuccess 범위 한정

```go
func renderUseMutation(a ActionBlock, fetchOps []string) string {
    // ...
    invalidate := ""
    if len(fetchOps) > 0 {
        keys := make([]string, len(fetchOps))
        for i, op := range fetchOps {
            keys[i] = fmt.Sprintf("{ queryKey: ['%s'] }", op)
        }
        for _, k := range keys {
            invalidate += fmt.Sprintf("\n      queryClient.invalidateQueries(%s)", k)
        }
    } else {
        invalidate = "\n      queryClient.invalidateQueries()"
    }
    // ...
}
```

GeneratePage에서 fetchOps 목록을 수집하여 전달:

```go
var fetchOps []string
for _, f := range page.Fetches {
    fetchOps = append(fetchOps, f.OperationID)
}
```

## Step 4. 테스트 업데이트

### 파서 테스트

기존 4개 테스트에 새 필드 검증 추가:
- Tag, ClassName 값 확인
- SubmitText 값 확인
- FieldBind.Placeholder, ClassName 확인
- StateBind.Text 확인
- ChildNode 순서 확인 (DOM 순서 보존 검증)

### 코드젠 테스트

기대 결과를 개선된 TSX로 업데이트:
- `assertContains(t, code, "className=")` 추가
- `assertContains(t, code, "placeholder=")` 추가
- `assertContains(t, code, "예약하기")` (버튼 텍스트)
- `assertNotContains(t, code, "제출")` (하드코딩 텍스트 없음)
- DeleteRoom: `assertContains(t, code, "onClick")`, `assertNotContains(t, code, "<form")`
- onSuccess: `assertContains(t, code, "queryKey: ['ListMyReservations']")`

### 더미 스터디 통합 테스트

`stml gen specs/dummy-study` 실행 후 생성된 TSX를 기대값과 비교:
- 코드젠개선방안.md의 기대 결과(my-reservations-page.tsx)와 diff가 없어야 한다

## Step 5. CLI 변경

변경 없음. `stml gen`은 parse → validate → generate 파이프라인 그대로.

## 구현 순서

```
1. types.go     — ChildNode, StaticElement 정의, 기존 구조체에 Tag/ClassName/Text 필드 추가
2. parser.go    — 속성 수집 확장, ChildNode 순서 수집, StaticElement 수집, 텍스트 수집
3. parser_test  — 새 필드 검증 추가
4. templates.go — 태그/className 출력, ChildNode 순서 출력, 텍스트 보존, 필드없는 action, data-state, onSuccess
5. generator.go — fetchOps 목록 전달
6. generator_test — 개선된 기대값으로 테스트 업데이트
7. 더미 스터디 통합 테스트 — gen 실행 + 결과 비교
```

## 완료 기준

- [ ] 생성된 TSX에 원본 HTML 태그 보존 (section, article, ul, li 등)
- [ ] 생성된 TSX에 Tailwind className 보존
- [ ] 버튼 텍스트가 원본과 일치 ("예약하기", "스터디룸 삭제" 등)
- [ ] input에 placeholder, className 보존
- [ ] 필드 없는 action이 button onClick으로 렌더링
- [ ] data-state가 조건부 렌더링으로 구현 (주석 아님)
- [ ] mutation onSuccess가 관련 queryKey만 무효화
- [ ] DOM 순서가 원본 HTML과 일치
- [ ] 기존 테스트 전부 통과 (하위 호환)
- [ ] 더미 스터디 코드젠 결과가 기대값과 일치
