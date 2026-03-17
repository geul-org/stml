# Phase 1: Parse

> `specs/dummy-study/frontend/*.html`의 data-* 속성을 파싱하여 Go 구조체 리스트로 변환한다.

## 목표

HTML5 파일을 읽어 `[]PageSpec` 슬라이스를 산출한다. 이후 Phase 2(codegen)와 Phase 3(validate)의 입력이 된다.

## 입력

```
specs/dummy-study/frontend/*.html
```

각 파일은 하나의 페이지를 나타내며, data-* 속성으로 API 바인딩을 선언한다.

### HTML 예시

```html
<main class="flex flex-col h-screen">
  <section data-fetch="ListSessions" data-param-project-id="currentProject.id" class="flex-1 p-4">
    <ul data-each="sessions" class="space-y-2">
      <li class="flex justify-between p-3 border rounded">
        <span data-bind="command" class="font-mono"></span>
        <span data-bind="status" class="text-sm text-gray-500"></span>
      </li>
    </ul>
    <p data-state="sessions.empty" class="text-gray-400">세션이 없습니다</p>
  </section>

  <footer class="p-4 border-t">
    <div data-action="CreateSession" class="flex gap-2">
      <input data-field="ProjectID" type="hidden" />
      <input data-field="Command" placeholder="명령어 입력" />
      <button type="submit">생성</button>
    </div>
  </footer>
</main>
```

## 산출

`[]PageSpec` — 파일 단위 파싱 결과

```go
type PageSpec struct {
    Name     string       // 페이지명 (파일명에서 파생, e.g. "session-page")
    FileName string       // 원본 파일명 (e.g. "session-page.html")
    Fetches  []FetchBlock // data-fetch 블록 리스트
    Actions  []ActionBlock // data-action 블록 리스트
}
```

`FetchBlock` — data-fetch가 선언된 요소와 하위 바인딩

```go
type FetchBlock struct {
    OperationID string       // data-fetch 값 (e.g. "ListSessions")
    Params      []ParamBind  // data-param-* 속성들
    Binds       []FieldBind  // 하위 data-bind 속성들
    Eaches      []EachBlock  // 하위 data-each 속성들
    States      []StateBind  // 하위 data-state 속성들
    Components  []ComponentRef // 하위 data-component 속성들
    Children    []FetchBlock // 중첩 data-fetch (부모 응답 참조)
}
```

`ActionBlock` — data-action이 선언된 요소와 하위 필드

```go
type ActionBlock struct {
    OperationID string       // data-action 값 (e.g. "CreateSession")
    Params      []ParamBind  // data-param-* 속성들
    Fields      []FieldBind  // 하위 data-field 속성들
}
```

공통 타입

```go
type ParamBind struct {
    Name   string // 파라미터명 (e.g. "projectId", data-param-project-id에서 추출)
    Source string // 값 소스 (e.g. "currentProject.id", "route.projectId")
}

type FieldBind struct {
    Name string // 필드명 (e.g. "Command")
    Tag  string // HTML 태그 (e.g. "input", "span")
    Type string // input type 속성 (e.g. "hidden", "number", "")
}

type EachBlock struct {
    Field  string      // 배열 필드명 (e.g. "sessions")
    Binds  []FieldBind // 반복 내부 data-bind
    States []StateBind // 반복 내부 data-state
    Components []ComponentRef // 반복 내부 data-component
}

type StateBind struct {
    Condition string // 조건 표현 (e.g. "sessions.empty", "canDelete")
}

type ComponentRef struct {
    Name  string // 컴포넌트명 (e.g. "Chart", "KanbanBoard")
    Bind  string // data-bind 값 (있을 경우, e.g. "usageData")
    Field string // data-field 값 (있을 경우, e.g. "ReservedAt")
}
```

## 구현 계획

### 1. 프로젝트 초기화

- `go mod init github.com/park-jun-woo/stml`
- 디렉토리 구조:
  ```
  artifacts/
    cmd/stml/main.go         # CLI 진입점
    internal/
      parser/
        parser.go            # 파싱 로직
        types.go             # PageSpec, FetchBlock 등 구조체 정의
      parser_test.go         # 테스트
  ```

### 2. 구조체 정의 (`types.go`)

- `PageSpec`, `FetchBlock`, `ActionBlock`, `ParamBind`, `FieldBind`, `EachBlock`, `StateBind`, `ComponentRef` 정의

### 3. 파서 구현 (`parser.go`)

- `golang.org/x/net/html`로 HTML 파일 파싱
- DOM 트리를 재귀 순회하며 data-* 속성 추출:
  1. `data-fetch` → 새 FetchBlock 시작, 하위 트리에서 bind/each/state/component 수집
  2. `data-action` → 새 ActionBlock 시작, 하위 트리에서 field 수집
  3. `data-param-*` → 속성명에서 파라미터명 추출 (kebab-case → camelCase)
  4. 중첩 data-fetch는 Children으로 트리 구성

### 4. 파라미터명 변환

- `data-param-project-id` → Name: `projectId` (kebab → camelCase)
- 속성값은 그대로 Source로 저장

### 5. 테스트

- STML.md의 두 예시(session-page, project-detail-page)를 `specs/dummy-study/frontend/`에 배치
- 파싱 결과를 기대값과 비교

## 완료 기준

- [ ] `specs/dummy-study/frontend/*.html` 파일을 읽어 `[]PageSpec`를 반환
- [ ] 8종 data-* 속성 모두 파싱 가능
- [ ] 중첩 data-fetch 처리
- [ ] data-param-* kebab-case → camelCase 변환
- [ ] 기획서 예시 2개(session-page, project-detail-page) 파싱 테스트 통과
