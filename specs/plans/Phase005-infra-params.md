# Phase 5: 인프라 파라미터 (OpenAPI x- 확장 지원)

> data-fetch 목록 엔드포인트에 페이지네이션, 정렬, 필터, 관계 포함을 선언하고, OpenAPI x- 확장과 교차 검증한다.

## 배경

`files/OpenAPI-x-extensions.md`에 정의된 4개 x- 확장(x-pagination, x-sort, x-filter, x-include)이 OpenAPI 엔드포인트에 선언된다. 현재 STML은 이를 인지하지 못하므로:

- 목록 API의 페이지네이션 UI를 생성할 수 없다
- 정렬/필터 UI를 생성할 수 없다
- 관계 포함 파라미터를 전달할 수 없다
- STML에서 사용한 sort/filter 컬럼이 OpenAPI allowed에 포함되는지 검증할 수 없다

## 설계 원칙

STML의 기존 철학을 유지한다:

1. **what만 선언** — HTML에 "이 목록은 페이지네이션한다, 이 컬럼으로 정렬한다"를 선언. how(페이지 버튼 UI, 드롭다운 UI)는 코드젠이 결정
2. **OpenAPI가 진실** — x-pagination의 style/defaultLimit, x-sort의 allowed 목록은 OpenAPI에서 읽는다. STML은 사용 여부와 기본값만 선언
3. **교차 검증** — STML에서 선언한 sort/filter/include 컬럼이 OpenAPI x- allowed에 포함되는지 validate 단계에서 검증

## 새 data-* 속성

기존 8개 속성에 4개 추가. 모두 `data-fetch` 요소의 수식어(modifier)로만 사용한다.

| 속성 | 역할 | 값 | 예시 |
|---|---|---|---|
| `data-paginate` | 페이지네이션 활성 | (값 없음 = boolean) | `data-paginate` |
| `data-sort` | 기본 정렬 컬럼 | `컬럼명` 또는 `컬럼명:desc` | `data-sort="start_at:desc"` |
| `data-filter` | 필터 UI 노출 컬럼 | 쉼표 구분 컬럼 목록 | `data-filter="status,room_id"` |
| `data-include` | 관계 포함 | 쉼표 구분 리소스 목록 | `data-include="room,user"` |

### 속성 상세

**data-paginate**
- 값 없이 존재하면 활성. OpenAPI x-pagination에서 style, defaultLimit, maxLimit을 읽는다
- `style: offset` → useQuery + useState(page, limit) + 페이지 컨트롤 UI
- `style: cursor` → useInfiniteQuery + "더 보기" 버튼

**data-sort**
- 값 = 기본 정렬 컬럼. 방향은 `:`으로 구분 (생략 시 x-sort.direction 또는 `asc`)
- 코드젠: useState(sortBy, sortDir) + 정렬 토글 UI
- 검증: 값이 x-sort.allowed에 포함되는지

**data-filter**
- 값 = 필터 UI를 노출할 컬럼 쉼표 목록
- 코드젠: 각 컬럼에 대한 input/select + useState + queryKey 포함
- 검증: 각 컬럼이 x-filter.allowed에 포함되는지

**data-include**
- 값 = 함께 요청할 관계 리소스 쉼표 목록
- 코드젠: queryFn에 include 파라미터 전달 (UI 없음, 데이터 보강)
- 검증: 각 리소스가 x-include.allowed에 포함되는지

## STML 예시

```html
<!-- 기존 (단순 목록) -->
<section data-fetch="ListMyReservations">
  <ul data-each="reservations">
    <li><span data-bind="RoomID"></span></li>
  </ul>
</section>

<!-- Phase 5 (인프라 파라미터 추가) -->
<section data-fetch="ListMyReservations"
         data-paginate
         data-sort="start_at:desc"
         data-filter="status,room_id"
         data-include="room">
  <ul data-each="reservations">
    <li>
      <span data-bind="room.Name"></span>
      <span data-bind="StartAt"></span>
      <span data-bind="Status"></span>
    </li>
  </ul>
</section>
```

## 기대 코드젠 출력

### offset 페이지네이션 + 정렬 + 필터

```tsx
export default function MyReservationsPage() {
  const [page, setPage] = useState(1)
  const [limit] = useState(20)
  const [sortBy, setSortBy] = useState('start_at')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')
  const [filters, setFilters] = useState<Record<string, string>>({})

  const { data, isLoading, error } = useQuery({
    queryKey: ['ListMyReservations', page, limit, sortBy, sortDir, filters],
    queryFn: () => api.ListMyReservations({
      page, limit, sortBy, sortDir,
      include: 'room',
      ...filters,
    }),
  })

  return (
    <section>
      {/* 필터 UI */}
      <div className="flex gap-2 mb-4">
        <input placeholder="status" value={filters.status ?? ''}
               onChange={(e) => setFilters(f => ({ ...f, status: e.target.value }))} />
        <input placeholder="room_id" value={filters.room_id ?? ''}
               onChange={(e) => setFilters(f => ({ ...f, room_id: e.target.value }))} />
      </div>

      {/* 정렬 토글 */}
      <div className="flex gap-2 mb-4">
        <button onClick={() => { setSortBy('start_at'); setSortDir(d => d === 'asc' ? 'desc' : 'asc') }}>
          시작일 {sortBy === 'start_at' ? (sortDir === 'asc' ? '↑' : '↓') : ''}
        </button>
      </div>

      {/* 목록 */}
      <ul>
        {data?.reservations?.map((item, index) => (
          <li key={index}>
            <span>{item.room.Name}</span>
            <span>{item.StartAt}</span>
            <span>{item.Status}</span>
          </li>
        ))}
      </ul>

      {/* 페이지네이션 */}
      <div className="flex justify-between mt-4">
        <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}>이전</button>
        <span>{page} / {Math.ceil((data?.total ?? 0) / limit)}</span>
        <button disabled={!data?.total || page * limit >= data.total}
                onClick={() => setPage(p => p + 1)}>다음</button>
      </div>
    </section>
  )
}
```

### cursor 페이지네이션

```tsx
const { data, fetchNextPage, hasNextPage, isLoading } = useInfiniteQuery({
  queryKey: ['ListMyReservations'],
  queryFn: ({ pageParam }) => api.ListMyReservations({ cursor: pageParam, limit: 20 }),
  getNextPageParam: (lastPage) => lastPage.nextCursor,
})

// ...
{hasNextPage && <button onClick={() => fetchNextPage()}>더 보기</button>}
```

## Step 1. 파서 구조체 확장 (`types.go`)

### FetchBlock에 인프라 파라미터 필드 추가

```go
type FetchBlock struct {
    // 기존 필드 유지 ...

    // Phase 5: 인프라 파라미터
    Paginate    bool          // data-paginate 존재 여부
    Sort        *SortDecl     // data-sort 파싱 결과
    Filters     []string      // data-filter 쉼표 파싱 결과
    Includes    []string      // data-include 쉼표 파싱 결과
}

type SortDecl struct {
    Column    string  // 기본 정렬 컬럼
    Direction string  // "asc" 또는 "desc"
}
```

## Step 2. 파서 수정 (`parser.go`)

### 2-1. data-paginate/sort/filter/include 속성 수집

`parseFetchBlock()`에서 추가 속성을 읽는다:

```go
if hasAttr(n, "data-paginate") {
    fb.Paginate = true
}
if v := getAttr(n, "data-sort"); v != "" {
    fb.Sort = parseSortDecl(v) // "start_at:desc" → SortDecl{"start_at", "desc"}
}
if v := getAttr(n, "data-filter"); v != "" {
    fb.Filters = strings.Split(v, ",")
}
if v := getAttr(n, "data-include"); v != "" {
    fb.Includes = strings.Split(v, ",")
}
```

### 2-2. parseSortDecl 헬퍼

```go
func parseSortDecl(v string) *SortDecl {
    parts := strings.SplitN(v, ":", 2)
    sd := &SortDecl{Column: strings.TrimSpace(parts[0]), Direction: "asc"}
    if len(parts) == 2 {
        sd.Direction = strings.TrimSpace(parts[1])
    }
    return sd
}
```

## Step 3. 심볼 테이블 확장 (`symbol.go`)

### 3-1. OpenAPI x- 확장 파싱

`APISymbol`에 인프라 파라미터 필드 추가:

```go
type APISymbol struct {
    Method         string
    Parameters     []ParamSymbol
    RequestFields  []FieldSymbol
    ResponseFields []FieldSymbol

    // Phase 5
    Pagination *PaginationExt
    Sort       *SortExt
    Filter     *FilterExt
    Include    *IncludeExt
}

type PaginationExt struct {
    Style        string // "offset" or "cursor"
    DefaultLimit int
    MaxLimit     int
}

type SortExt struct {
    Allowed   []string
    Default   string
    Direction string
}

type FilterExt struct {
    Allowed []string
}

type IncludeExt struct {
    Allowed []string
}
```

### 3-2. YAML 파싱

OpenAPI operation 파싱 시 x- 키 존재 여부를 확인하고 구조체로 변환:

```go
if xp, ok := opMap["x-pagination"]; ok { ... }
if xs, ok := opMap["x-sort"]; ok { ... }
if xf, ok := opMap["x-filter"]; ok { ... }
if xi, ok := opMap["x-include"]; ok { ... }
```

## Step 4. 검증 확장 (`validator.go`)

### 4-1. 새 검증 규칙 4개

| # | 검증 | 조건 | 에러 |
|---|---|---|---|
| 9 | data-paginate 사용 시 x-pagination 존재 | fb.Paginate && sym.Pagination == nil | "paginate declared but no x-pagination on {op}" |
| 10 | data-sort 컬럼이 x-sort.allowed에 포함 | fb.Sort.Column ∉ sym.Sort.Allowed | "sort column '{col}' not in x-sort.allowed of {op}" |
| 11 | data-filter 컬럼이 x-filter.allowed에 포함 | fb.Filter[i] ∉ sym.Filter.Allowed | "filter column '{col}' not in x-filter.allowed of {op}" |
| 12 | data-include 리소스가 x-include.allowed에 포함 | fb.Include[i] ∉ sym.Include.Allowed | "include '{res}' not in x-include.allowed of {op}" |

### 4-2. 역방향 검증 (선택)

x- 확장이 있는데 STML에서 사용하지 않는 경우는 경고(warning)로 처리. 에러는 아니다 — 모든 인프라 기능을 사용할 의무는 없다.

## Step 5. 코드젠 확장

### 5-1. imports 확장 (`imports.go`)

```go
// data-paginate + offset → useState 필요
if hasPaginate(page) {
    is.useState = true
}
// data-sort → useState
if hasSort(page) {
    is.useState = true
}
// data-filter → useState
if hasFilter(page) {
    is.useState = true
}
// cursor 스타일 → useInfiniteQuery
if hasCursorPaginate(page) {
    is.useInfiniteQuery = true
}
```

### 5-2. hooks 생성 (`generator.go`)

페이지네이션/정렬/필터가 있는 fetch에 대해:

```go
// useState hooks
if fb.Paginate { renderPaginationState(fb, sym) }
if fb.Sort != nil { renderSortState(fb) }
if len(fb.Filters) > 0 { renderFilterState(fb) }

// useQuery에 파라미터 추가
queryKey에 page, limit, sortBy, sortDir, filters 포함
queryFn에 해당 파라미터 전달
```

### 5-3. JSX 생성 (`templates.go`)

```go
func renderFetchJSX(f FetchBlock, ...) {
    // 기존 로딩/에러/데이터 블록

    if len(f.Filters) > 0 {
        // 필터 입력 UI 렌더링
    }
    if f.Sort != nil {
        // 정렬 토글 UI 렌더링
    }

    // 기존 children 렌더링

    if f.Paginate {
        // 페이지네이션 컨트롤 렌더링 (style에 따라 offset/cursor)
    }
}
```

## Step 6. 더미 스터디 업데이트

### 6-1. OpenAPI에 x- 확장 추가

`specs/dummy-study/api/openapi.yaml`의 `ListMyReservations`에:

```yaml
/me/reservations:
  get:
    operationId: ListMyReservations
    x-pagination:
      style: offset
      defaultLimit: 20
      maxLimit: 100
    x-sort:
      allowed: [StartAt, CreatedAt]
      default: StartAt
      direction: desc
    x-filter:
      allowed: [Status, RoomID]
```

### 6-2. STML HTML 업데이트

`specs/dummy-study/frontend/my-reservations-page.html`에 인프라 속성 추가:

```html
<section data-fetch="ListMyReservations"
         data-paginate
         data-sort="StartAt:desc"
         data-filter="Status">
```

### 6-3. 기대 TSX

기존 my-reservations-page.tsx에 페이지네이션 컨트롤, 정렬 토글, 필터 입력이 추가된 결과.

## Step 7. 테스트

### 파서 테스트

- data-paginate가 있으면 `fb.Paginate == true`
- data-sort="start_at:desc" → `fb.Sort.Column == "start_at"`, `fb.Sort.Direction == "desc"`
- data-sort="start_at" → `fb.Sort.Direction == "asc"` (기본값)
- data-filter="status,room_id" → `fb.Filters == ["status", "room_id"]`
- data-include="room" → `fb.Includes == ["room"]`
- 인프라 속성 없으면 모두 zero value

### 검증 테스트

- data-paginate + x-pagination 있음 → pass
- data-paginate + x-pagination 없음 → error
- data-sort 컬럼이 x-sort.allowed에 있음 → pass
- data-sort 컬럼이 x-sort.allowed에 없음 → error
- data-filter 컬럼이 x-filter.allowed에 없음 → error
- data-include 리소스가 x-include.allowed에 없음 → error

### 코드젠 테스트

- data-paginate(offset) → assertContains: `useState`, `setPage`, `이전`, `다음`
- data-paginate(cursor) → assertContains: `useInfiniteQuery`, `fetchNextPage`, `더 보기`
- data-sort → assertContains: `setSortBy`, `setSortDir`
- data-filter → assertContains: `setFilters`, `onChange`
- data-include → assertContains: `include: 'room'`
- queryKey에 인프라 파라미터 포함 확인

## 구현 순서

```
1. types.go       — SortDecl, FetchBlock 확장
2. parser.go      — data-paginate/sort/filter/include 파싱
3. parser_test.go  — 인프라 속성 파싱 테스트
4. symbol.go      — PaginationExt, SortExt, FilterExt, IncludeExt 타입 + x- 파싱
5. errors.go      — 4개 에러 생성자 추가
6. validator.go   — 4개 검증 규칙 추가
7. validator_test.go — 6개 검증 테스트
8. imports.go     — useState, useInfiniteQuery 임포트 조건
9. templates.go   — 페이지네이션/정렬/필터/include JSX 렌더링
10. generator.go   — useState hooks, queryKey/queryFn 파라미터 확장
11. generator_test.go — 코드젠 테스트
12. 더미 스터디    — OpenAPI x- 추가 + STML 업데이트 + gen 실행
```

## 완료 기준

- [ ] data-paginate(offset) → useState + 페이지 컨트롤 UI 생성
- [ ] data-paginate(cursor) → useInfiniteQuery + "더 보기" 생성
- [ ] data-sort → useState + 정렬 토글 UI 생성
- [ ] data-filter → useState + 필터 입력 UI 생성
- [ ] data-include → queryFn에 include 파라미터 전달
- [ ] queryKey에 page/sort/filter 파라미터 포함
- [ ] STML sort/filter/include 값이 OpenAPI x- allowed와 교차 검증
- [ ] data-paginate 사용 시 x-pagination 존재 검증
- [ ] 더미 스터디에 인프라 파라미터 적용 + 코드젠 결과 확인
- [ ] 기존 테스트 전부 통과 (하위 호환)
- [ ] 인프라 속성 없는 기존 fetch는 동작 변경 없음
