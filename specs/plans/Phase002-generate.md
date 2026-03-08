# Phase 2: Generate

> 파싱된 `[]PageSpec`을 data-* 타입별 템플릿으로 매칭하여 React 컴포넌트를 생성한다.

## 목표

Phase 1의 산출물(`[]PageSpec`)을 입력받아, 각 data-* 속성에 대응하는 React + TypeScript 코드를 생성하고 `artifacts/frontend/`에 파일로 출력한다.

## 입력

Phase 1의 `[]PageSpec`

## 산출

```
artifacts/frontend/<page-name>.tsx
```

파일당 하나의 React 페이지 컴포넌트. `prettier` 적용 완료 상태.

## 속성별 템플릿 (8종)

### data-fetch → useQuery

```tsx
const { data: {{.Alias}}, isLoading: {{.Alias}}Loading, error: {{.Alias}}Error } = useQuery({
  queryKey: ['{{.OperationID}}', {{.ParamValues}}],
  queryFn: () => api.{{.OperationID}}({{.ParamArgs}}),
})
```

JSX 영역:

```tsx
{{{.Alias}}Loading && <div>로딩 중...</div>}
{{{.Alias}}Error && <div>오류가 발생했습니다</div>}
{{{.Alias}} && (
  // 하위 바인딩 렌더링
)}
```

### data-action → useMutation

```tsx
const {{.MutationName}} = useMutation({
  mutationFn: (data: {{.RequestType}}) => api.{{.OperationID}}(data),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['{{.RelatedQueryKey}}'] })
  },
})
```

폼 래퍼:

```tsx
<form onSubmit={(e) => {
  e.preventDefault()
  {{.MutationName}}.mutate(formData)
}}>
  {/* data-field 입력들 */}
</form>
```

### data-field → 폼 입력 바인딩

```tsx
// type="hidden"
<input type="hidden" value={{{.Source}}} {...register('{{.Name}}')} />

// 일반 input
<input placeholder="{{.Placeholder}}" {...register('{{.Name}}')} className="{{.Classes}}" />

// type="number"
<input type="number" {...register('{{.Name}}', { valueAsNumber: true })} className="{{.Classes}}" />
```

### data-bind → 응답 필드 출력

```tsx
// 일반 텍스트
<span className="{{.Classes}}">{item.{{.Field}}}</span>

// custom.ts 함수 (OpenAPI response에 없는 필드)
<span className="{{.Classes}}">{custom.{{.Field}}({{.Args}})}</span>
```

### data-param-* → 파라미터 전달

쿼리/경로 파라미터를 API 호출 인자로 변환:

```tsx
// data-param-project-id="currentProject.id"
api.ListSessions({ projectId: currentProject.id })

// data-param-project-id="route.projectId"
const { projectId } = useParams()
api.GetProject({ projectId })
```

`route.*` 소스는 `useParams()` 훅 호출을 생성하고, 그 외는 컨텍스트나 상위 fetch 응답에서 참조한다.

### data-each → 배열 반복

```tsx
{{{.Data}}.{{.Field}}?.map((item) => (
  <{{.Tag}} key={item.id} className="{{.Classes}}">
    {/* 내부 data-bind 렌더링 */}
  </{{.Tag}}>
))}
```

### data-state → 조건부 표시

```tsx
// data-state="sessions.empty"
{{{.Data}}.sessions?.length === 0 && (
  <p className="{{.Classes}}">세션이 없습니다</p>
)}

// data-state="canDelete"
{canDelete && (
  // 하위 요소 렌더링
)}
```

조건 표현 매핑:
- `<field>.empty` → `data?.field?.length === 0`
- `<field>.loading` → `fieldLoading`
- `<field>.error` → `fieldError`
- 그 외 → 로컬 상태 변수

### data-component → 컴포넌트 위임

```tsx
import Chart from '@/components/Chart'
import KanbanBoard from '@/components/KanbanBoard'

// data-component + data-bind
<Chart data={fetchData.usageData} />

// data-component + data-field
<DatePicker {...register('ReservedAt')} />

// data-component + children
<PageTransition>
  {/* 하위 요소 렌더링 */}
</PageTransition>
```

## 구현 계획

### 1. 디렉토리 구조

```
artifacts/
  internal/
    generator/
      generator.go       # 메인 생성 로직
      templates.go       # 속성별 템플릿 정의
      imports.go         # import 자동 수집
    generator_test.go    # 테스트
```

### 2. 템플릿 엔진

- `text/template`으로 속성별 템플릿 등록
- FetchBlock, ActionBlock 각각을 템플릿 데이터로 변환하는 어댑터 함수

### 3. 코드 생성 흐름

```
PageSpec 순회
  → 파일 헤더(import) 수집
    → data-fetch → useQuery import
    → data-action → useMutation import
    → data-component → 컴포넌트 import
    → route.* 참조 → useParams import
    → custom.ts 존재 → custom import
  → 훅 선언 생성 (useQuery, useMutation, useParams)
  → JSX 트리 생성
    → FetchBlock → 로딩/에러/데이터 분기
    → ActionBlock → form + onSubmit
    → EachBlock → .map()
    → StateBind → 조건부 렌더링
    → ComponentRef → 컴포넌트 렌더링
  → 파일 쓰기
```

### 4. custom.ts 병합

- `specs/frontend/<page-name>.custom.ts` 존재 시:
  - `import * as custom from './<page-name>.custom'` 추가
  - data-bind 필드가 OpenAPI response에 없고 custom.ts에 있으면 함수 호출로 렌더링

### 5. 보조 로직

- `data-fetch="ListSessions"` → alias: `listSessionsData` (camelCase + Data 접미사)
- `data-action="CreateSession"` → mutation: `createSessionMutation`
- data-param의 `route.*` 감지 → `useParams()` 호출 생성
- 중첩 data-fetch → 부모 응답 데이터를 자식 파라미터로 전달

### 6. 테스트

- session-page, project-detail-page 예시의 코드젠 결과를 기대값과 비교
- custom.ts 병합 케이스 (cart-page) 테스트

## 완료 기준

- [ ] 8종 속성 모두 템플릿 구현
- [ ] `stml gen` 실행 시 `artifacts/frontend/`에 TSX 파일 생성
- [ ] custom.ts 동반 파일 자동 병합
- [ ] data-component import 자동 삽입
- [ ] route.* 파라미터 시 useParams 자동 생성
- [ ] 기획서 예시 3개(session-page, project-detail-page, cart-page) 코드젠 테스트 통과
